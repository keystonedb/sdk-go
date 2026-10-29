package tasks

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/keystonedb/sdk-go/test/requirements"
	"github.com/packaged/logger/v3/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Requirement struct{ taskName string }

func (d *Requirement) Name() string {
	return "TaskQ"
}

// Parallel fixtures use unique sensor/entity identities or an isolated task queue.
func (d *Requirement) Parallel() bool { return true }

func (d *Requirement) Register(conn *keystone.Connection) error {
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor, report requirements.Reporter) {
	scheduler := requirements.NewScheduler()
	scheduler.StartEarly(d.Name(), func() { d.VerifyScheduled(actor, report, scheduler) })
	scheduler.Run()
}

func (d *Requirement) VerifyScheduled(actor *keystone.Actor, report requirements.Reporter, scheduler *requirements.Scheduler) {
	d.taskName = "sdk-test-task-" + uuid.NewString()
	pushed := d.pusher(actor)
	report(pushed)
	if pushed.Error == nil {
		scheduler.After(0, func() { report(d.puller(actor)) })
	}
}

func (d *Requirement) pusher(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Task Pusher"}

	for i := 0; i < 10; i++ {
		pushErr := actor.TaskPush(context.Background(), d.taskName, uuid.NewString(), map[string]string{
			"task_number": strconv.FormatInt(int64(i), 10),
		})
		if pushErr != nil {
			return res.WithError(pushErr)
		}
	}

	return res
}

func (d *Requirement) puller(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Task Puller"}

	counter := 0
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	err := actor.TaskStream(ctx, d.taskName, func(response *proto.TaskResponse) error {
		counter++
		logger.I().Info("Task Stream", zap.Int("Counter", counter), zap.Any("task_number", response))
		if counter >= 10 {
			cancel()
		}
		return nil
	})
	if counter != 10 {
		return res.WithError(fmt.Errorf("want 10 tasks, received %d: %v", counter, err))
	}
	if ctx.Err() == context.Canceled && (err == nil || status.Code(err) == codes.Canceled) {
		return res
	}
	return res.WithError(err)
}
