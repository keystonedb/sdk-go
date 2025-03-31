package event_stream

import (
	"context"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/keystonedb/sdk-go/test/requirements"
	"github.com/packaged/logger/v3/logger"
	"go.uber.org/zap"
)

type Requirement struct{}

func (d *Requirement) Name() string {
	return "Event Stream"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.action(actor),
	}
}

func (d *Requirement) action(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Stream Receive"}

	err := actor.EventStream(context.Background(), d.Handle, "actions", true, keystone.OwnKey("tst2"))
	return res.WithError(err)
}

func (d *Requirement) Handle(response *proto.EventStreamResponse) error {
	logger.I().Info("Got Response", zap.Any("response", response))
	return nil
}
