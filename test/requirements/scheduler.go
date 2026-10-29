package requirements

import (
	"sort"
	"sync"
	"time"

	"github.com/keystonedb/sdk-go/keystone"
)

// ScheduledRequirement opts into early setup and resumable checks.
type ScheduledRequirement interface {
	Requirement
	VerifyScheduled(*keystone.Actor, Reporter, *Scheduler)
}

// ParallelRequirement opts independent fixtures into concurrent execution.
// Registration is always sequential. Requirements without this opt-in execute
// exclusively, and callbacks belonging to one requirement never overlap.
type ParallelRequirement interface {
	Requirement
	Parallel() bool
}

type Timing struct {
	// Effective is wall time with at least one active check, including RPC latency.
	Effective time.Duration
	// Work sums callback durations; it may exceed elapsed time under concurrency.
	Work    time.Duration
	Idle    time.Duration
	Elapsed time.Duration
}

type scheduledCheck struct {
	name     string
	ready    time.Time
	priority int
	parallel bool
	run      func(*Scheduler)
}
type activeInterval struct{ start, end time.Time }
type completedCheck struct {
	name     string
	interval activeInterval
}

// Scheduler runs ready checks before sleeping. Opted-in requirements can share
// worker slots, but each requirement retains serial callbacks. A parallel callback
// must schedule continuations through the scoped Scheduler passed to it.
type Scheduler struct {
	parent      *Scheduler
	name        string
	parallel    bool
	mu          sync.Mutex
	now         func() time.Time
	sleep       func(time.Duration)
	pending     []scheduledCheck
	current     string
	active      map[string]time.Duration
	wake        chan struct{}
	parallelism int
	running     bool
}

func NewScheduler() *Scheduler {
	return &Scheduler{now: time.Now, sleep: time.Sleep, active: make(map[string]time.Duration), wake: make(chan struct{}, 1), parallelism: 4}
}
func (s *Scheduler) root() *Scheduler {
	if s.parent != nil {
		return s.parent
	}
	return s
}
func (s *Scheduler) SetParallelism(limit int) {
	r := s.root()
	r.mu.Lock()
	defer r.mu.Unlock()
	if limit < 1 || r.running {
		panic("set a positive parallelism limit before running")
	}
	r.parallelism = limit
}
func (s *Scheduler) enqueue(name string, delay time.Duration, priority int, parallel bool, run func(*Scheduler)) {
	r := s.root()
	r.mu.Lock()
	r.pending = append(r.pending, scheduledCheck{name, r.now().Add(max(0, delay)), priority, parallel, run})
	r.mu.Unlock()
	select {
	case r.wake <- struct{}{}:
	default:
	}
}
func (s *Scheduler) Start(name string, run func()) {
	s.enqueue(name, 0, 1, false, func(*Scheduler) { run() })
}
func (s *Scheduler) StartEarly(name string, run func()) {
	s.enqueue(name, 0, -1, false, func(*Scheduler) { run() })
}
func (s *Scheduler) StartParallel(name string, run func(*Scheduler)) {
	s.enqueue(name, 0, 1, true, run)
}
func (s *Scheduler) StartEarlyParallel(name string, run func(*Scheduler)) {
	s.enqueue(name, 0, -1, true, run)
}
func (s *Scheduler) After(delay time.Duration, run func()) {
	r := s.root()
	name, parallel := s.name, s.parallel
	if s.parent == nil {
		r.mu.Lock()
		name = r.current
		r.mu.Unlock()
	}
	if name == "" {
		panic("schedule continuations from a running requirement")
	}
	s.enqueue(name, delay, 0, parallel, func(*Scheduler) { run() })
}
func (s *Scheduler) Active(name string) time.Duration {
	r := s.root()
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.active[name]
}

// activeDuration merges overlaps so concurrent work is never subtracted twice.
func activeDuration(intervals []activeInterval) time.Duration {
	if len(intervals) == 0 {
		return 0
	}
	sort.Slice(intervals, func(i, j int) bool { return intervals[i].start.Before(intervals[j].start) })
	current := intervals[0]
	var total time.Duration
	for _, next := range intervals[1:] {
		if next.start.After(current.end) {
			total += current.end.Sub(current.start)
			current = next
		} else if next.end.After(current.end) {
			current.end = next.end
		}
	}
	return total + current.end.Sub(current.start)
}

func (s *Scheduler) Run() Timing {
	r := s.root()
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		panic("scheduler is already running")
	}
	r.running = true
	limit := r.parallelism
	r.mu.Unlock()
	defer func() { r.mu.Lock(); r.running = false; r.mu.Unlock() }()
	started := r.now()
	done := make(chan completedCheck, limit)
	busy := map[string]bool{}
	var intervals []activeInterval
	var timing Timing
	record := func(completed completedCheck) {
		duration := completed.interval.end.Sub(completed.interval.start)
		intervals = append(intervals, completed.interval)
		timing.Work += duration
		r.mu.Lock()
		r.active[completed.name] += duration
		r.mu.Unlock()
	}
	for {
		// Release completed workers before looking for the next eligible callback.
		for draining := true; draining; {
			select {
			case completed := <-done:
				delete(busy, completed.name)
				record(completed)
			default:
				draining = false
			}
		}
		r.mu.Lock()
		now := r.now()
		ready := -1
		var next time.Time
		for i, job := range r.pending {
			if busy[job.name] || (!job.parallel && len(busy) > 0) || (job.parallel && len(busy) >= limit) {
				continue
			}
			if next.IsZero() || job.ready.Before(next) {
				next = job.ready
			}
			if job.ready.After(now) {
				continue
			}
			if ready < 0 || job.priority < r.pending[ready].priority || (job.priority == r.pending[ready].priority && job.ready.Before(r.pending[ready].ready)) {
				ready = i
			}
		}
		if ready >= 0 {
			job := r.pending[ready]
			r.pending = append(r.pending[:ready], r.pending[ready+1:]...)
			if !job.parallel {
				r.current = job.name
			}
			r.mu.Unlock()
			scoped := &Scheduler{parent: r, name: job.name, parallel: job.parallel}
			if job.parallel {
				busy[job.name] = true
				go func() {
					began := r.now()
					job.run(scoped)
					done <- completedCheck{job.name, activeInterval{began, r.now()}}
				}()
			} else {
				began := r.now()
				job.run(scoped)
				record(completedCheck{job.name, activeInterval{began, r.now()}})
				r.mu.Lock()
				r.current = ""
				r.mu.Unlock()
			}
			continue
		}
		empty := len(r.pending) == 0
		r.mu.Unlock()
		if empty && len(busy) == 0 {
			break
		}
		if len(busy) == 0 {
			// Retain the injected clock used by deterministic sequential timing tests.
			r.sleep(next.Sub(r.now()))
			continue
		}
		var timer *time.Timer
		var tick <-chan time.Time
		if !next.IsZero() {
			timer = time.NewTimer(max(0, next.Sub(r.now())))
			tick = timer.C
		}
		select {
		case completed := <-done:
			delete(busy, completed.name)
			record(completed)
		case <-r.wake:
		case <-tick:
		}
		if timer != nil {
			timer.Stop()
		}
	}
	timing.Elapsed = r.now().Sub(started)
	timing.Effective = activeDuration(intervals)
	timing.Idle = timing.Elapsed - timing.Effective
	return timing
}
