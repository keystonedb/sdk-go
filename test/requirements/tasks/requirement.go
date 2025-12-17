package tasks

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/keystonedb/sdk-go/test/requirements"
	"github.com/packaged/logger/v3/logger"
	"go.uber.org/zap"
)

type Requirement struct{}

func (d *Requirement) Name() string {
	return "TaskQ"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.pusher(actor),
		d.puller(actor),
	}
}

func (d *Requirement) pusher(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Task Pusher"}

	for i := 0; i < 10; i++ {
		pushErr := actor.TaskPush(context.Background(), "test-task", uuid.NewString(), map[string]string{
			"task_number": strconv.FormatInt(int64(i), 10),
		})
		if pushErr != nil {
			return res.WithError(pushErr)
		}
		time.Sleep(time.Millisecond * 10)
	}

	return res
}

func (d *Requirement) puller(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Task Puller"}

	counter := 0
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(20 * time.Second)
		cancel()
	}()

	return res.WithError(actor.TaskStream(ctx, "test-task", func(response *proto.TaskResponse) error {
		counter++
		logger.I().Info("Task Stream", zap.Int("Counter", counter), zap.Any("task_number", response))
		if counter >= 10 {
			time.Sleep(time.Second)
			cancel()
		}
		return nil
	}))
}
