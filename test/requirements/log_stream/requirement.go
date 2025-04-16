package log_stream

import (
	"context"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"github.com/packaged/logger/v3/logger"
	"log"
)

type Requirement struct {
	createdID keystone.ID
	labelKey  string
}

func (d *Requirement) Name() string {
	return "Log Streams"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.User{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.store(actor),
	}
}

func (d *Requirement) store(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Stream"}
	usr := &models.User{
		Validate: "logger",
	}
	writeErr := actor.Mutate(context.Background(), usr)
	if writeErr != nil {
		return res.WithError(writeErr)
	}

	log.Println(usr.GetKeystoneID())

	stream, err := keystone.NewLogStream(actor)
	if err != nil {
		return res.WithError(err)
	}

	go func() {
		logger.I().ErrorIf(stream.Start(), "Start Streaming Logs")
	}()

	for i := 0; i < 1000; i++ {
		logBatch := keystone.NewEntityLog(usr.GetKeystoneID())

		logBatch.Debug("This is a debug message", "ref1", "actor", "trace-123", map[string]string{"key1": "value1"})
		logBatch.Info("This is an info message", "ref1", "actor", "trace-123", map[string]string{"key1": "value1"})
		logBatch.Notice("This is a notice message", "ref1", "actor", "trace-123", map[string]string{"key1": "value1"})
		logBatch.Warn("This is a warning message", "ref1", "actor", "trace-123", map[string]string{"key1": "value1"})
		logBatch.Error("This is an error message", "ref1", "actor", "trace-123", map[string]string{"key1": "value1"})
		logBatch.Critical("This is a critical message", "ref1", "actor", "trace-123", map[string]string{"key1": "value1"})
		logBatch.Fatal("This is a fatal message", "ref1", "actor", "trace-123", map[string]string{"key1": "value1"})
		logBatch.Alert("This is an alert message", "ref1", "actor", "trace-123", map[string]string{"key1": "value1"})
		//time.Sleep(time.Millisecond)
		logErr := stream.LogBatch(logBatch)
		if logErr != nil {
			logger.I().ErrorIf(logErr, "Log Batch Error")
			return res.WithError(logErr)
		}
	}

	return res.WithError(stream.Stop())
}
