package requirements

import (
	"reflect"
	"testing"
	"time"
)

func fakeScheduler() (*Scheduler, *time.Time) {
	now := time.Unix(0, 0)
	s := NewScheduler()
	s.now = func() time.Time { return now }
	s.sleep = func(d time.Duration) { now = now.Add(d) }
	return s, &now
}

func TestSchedulerOverlapsWaitsAndCountsOnlyActiveWork(t *testing.T) {
	s, now := fakeScheduler()
	var order []string
	s.Start("ordinary", func() { order = append(order, "ordinary"); *now = now.Add(3 * time.Second) })
	s.StartEarly("ttl", func() {
		order = append(order, "ttl setup")
		*now = now.Add(time.Second)
		s.After(5*time.Second, func() { order = append(order, "ttl check"); *now = now.Add(time.Second) })
	})
	s.StartEarly("background", func() {
		order = append(order, "background setup")
		s.After(2*time.Second, func() { order = append(order, "background check") })
	})
	timing := s.Run()
	if want := []string{"ttl setup", "background setup", "ordinary", "background check", "ttl check"}; !reflect.DeepEqual(order, want) {
		t.Fatal(order, want)
	}
	if timing.Effective != 5*time.Second || timing.Idle != 2*time.Second || timing.Elapsed != 7*time.Second {
		t.Fatal(timing)
	}
	if s.Active("ordinary") != 3*time.Second || s.Active("ttl") != 2*time.Second {
		t.Fatal(s.active)
	}
}

func TestSchedulerRunsDuePollBeforeNextOrdinaryRequirement(t *testing.T) {
	s, now := fakeScheduler()
	polls := 0
	var check func()
	check = func() {
		polls++
		if polls < 3 {
			s.After(time.Second, check)
		}
	}
	s.StartEarly("poll", func() { s.After(time.Second, check) })
	s.Start("one", func() { *now = now.Add(2 * time.Second) })
	s.Start("two", func() {
		if polls != 1 {
			t.Fatal("due poll delayed", polls)
		}
		*now = now.Add(2 * time.Second)
	})
	timing := s.Run()
	if polls != 3 || timing.Effective != 4*time.Second || timing.Idle != time.Second {
		t.Fatal(polls, timing)
	}
}

func TestSchedulerOrdinaryOrderAndEmptyRun(t *testing.T) {
	s, _ := fakeScheduler()
	var order []int
	for i := range 3 {
		s.Start("ordinary", func() { order = append(order, i) })
	}
	if timing := s.Run(); timing != (Timing{}) {
		t.Fatal(timing)
	}
	if !reflect.DeepEqual(order, []int{0, 1, 2}) {
		t.Fatal(order)
	}
	if timing := s.Run(); timing != (Timing{}) {
		t.Fatal(timing)
	}
}

func await[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(3 * time.Second):
		t.Fatal("scheduler did not make progress")
		var zero T
		return zero
	}
}

func TestSchedulerParallelLimitAndExclusiveBarrier(t *testing.T) {
	s := NewScheduler()
	s.SetParallelism(2)
	entered := make(chan string, 3)
	release := make(chan struct{})
	for _, name := range []string{"a", "b", "c"} {
		s.StartParallel(name, func(*Scheduler) { entered <- name; <-release })
	}
	exclusive := make(chan struct{}, 1)
	s.Start("exclusive", func() { exclusive <- struct{}{} })
	finished := make(chan Timing, 1)
	go func() { finished <- s.Run() }()
	await(t, entered)
	await(t, entered)
	select {
	case <-entered:
		t.Error("exceeded parallelism limit")
	default:
	}
	select {
	case <-exclusive:
		t.Error("exclusive requirement overlapped workers")
	default:
	}
	close(release)
	await(t, entered)
	await(t, exclusive)
	timing := await(t, finished)
	if timing.Effective > timing.Elapsed || timing.Work < timing.Effective || timing.Idle < 0 {
		t.Fatal(timing)
	}
}

func TestSchedulerSerializesCallbacksWithinParallelRequirement(t *testing.T) {
	s := NewScheduler()
	initial := make(chan struct{})
	releaseInitial := make(chan struct{})
	first := make(chan struct{})
	releaseFirst := make(chan struct{})
	second := make(chan struct{}, 1)
	s.StartParallel("same", func(scoped *Scheduler) {
		scoped.After(0, func() { close(first); <-releaseFirst })
		scoped.After(0, func() { second <- struct{}{} })
		close(initial)
		<-releaseInitial
	})
	finished := make(chan Timing, 1)
	go func() { finished <- s.Run() }()
	await(t, initial)
	select {
	case <-first:
		t.Error("continuation overlapped setup")
	default:
	}
	close(releaseInitial)
	await(t, first)
	select {
	case <-second:
		t.Error("callbacks from one requirement overlapped")
	default:
	}
	close(releaseFirst)
	await(t, second)
	await(t, finished)
	if s.Active("same") <= 0 {
		t.Fatal("missing timing attribution")
	}
}

func TestSchedulerWorkerCompletionWakesForDelayedContinuation(t *testing.T) {
	s := NewScheduler()
	s.SetParallelism(1)
	done := make(chan struct{}, 1)
	s.StartEarlyParallel("poll", func(scoped *Scheduler) { scoped.After(time.Millisecond, func() { done <- struct{}{} }) })
	finished := make(chan Timing, 1)
	go func() { finished <- s.Run() }()
	await(t, done)
	timing := await(t, finished)
	if timing.Idle <= 0 {
		t.Fatal("delay counted as active work", timing)
	}
}

func TestActiveDurationMergesConcurrentWork(t *testing.T) {
	base := time.Unix(0, 0)
	span := func(a, b int) activeInterval {
		return activeInterval{base.Add(time.Duration(a) * time.Second), base.Add(time.Duration(b) * time.Second)}
	}
	if got := activeDuration([]activeInterval{span(3, 7), span(1, 5), span(9, 10), span(2, 4)}); got != 7*time.Second {
		t.Fatal(got)
	}
}
