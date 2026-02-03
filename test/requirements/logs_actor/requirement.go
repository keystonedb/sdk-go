package logs_actor

import (
	"context"
	"errors"
	"strings"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
)

type Requirement struct {
	createdID keystone.ID
}

func (d *Requirement) Name() string {
	return "Actor Logs"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.User{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.createEntity(actor),
		d.writeLog(actor),
		d.writeLogWithOptions(actor),
		d.readLogs(actor),
		d.readLogsWithMinLevel(actor),
	}
}

func (d *Requirement) createEntity(actor *keystone.Actor) requirements.TestResult {
	usr := &models.User{
		Validate: "actor-logs-test",
	}

	createErr := actor.Mutate(context.Background(), usr, keystone.WithMutationComment("Create entity for actor logs test"))
	if createErr == nil {
		d.createdID = usr.GetKeystoneID()
	}

	return requirements.TestResult{
		Name:  "Create Entity",
		Error: createErr,
	}
}

func (d *Requirement) writeLog(actor *keystone.Actor) requirements.TestResult {
	if d.createdID == "" {
		return requirements.TestResult{
			Name:  "Write Log",
			Error: errors.New("no entity created"),
		}
	}

	err := actor.Log(context.Background(), d.createdID.String(), proto.LogLevel_Info, "Test log message from Actor.Log")

	return requirements.TestResult{
		Name:  "Write Log",
		Error: err,
	}
}

func (d *Requirement) writeLogWithOptions(actor *keystone.Actor) requirements.TestResult {
	if d.createdID == "" {
		return requirements.TestResult{
			Name:  "Write Log With Options",
			Error: errors.New("no entity created"),
		}
	}

	err := actor.Log(
		context.Background(),
		d.createdID.String(),
		proto.LogLevel_Warn,
		"Test warning message with options",
		keystone.WithLogReference("test-ref-123"),
		keystone.WithLogTraceID("trace-abc-456"),
		keystone.WithLogData(map[string]string{
			"key1": "value1",
			"key2": "value2",
		}),
	)

	return requirements.TestResult{
		Name:  "Write Log With Options",
		Error: err,
	}
}

func (d *Requirement) readLogs(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Read Logs"}

	if d.createdID == "" {
		return res.WithError(errors.New("no entity created"))
	}

	logs, err := actor.Logs(context.Background(), d.createdID.String())
	if err != nil {
		return res.WithError(err)
	}

	if len(logs) < 2 {
		return res.WithError(errors.New("expected at least 2 log entries"))
	}

	// Verify we can find our test messages
	foundInfo := false
	foundWarn := false
	for _, log := range logs {
		if log.Level == proto.LogLevel_Info && strings.Contains(log.Message, "Test log message from Actor.Log") {
			foundInfo = true
		}
		if log.Level == proto.LogLevel_Warn && strings.Contains(log.Message, "Test warning message with options") {
			foundWarn = true
			// Verify options were stored
			if log.Reference != "test-ref-123" {
				return res.WithError(errors.New("expected reference 'test-ref-123'"))
			}
			if log.TraceId != "trace-abc-456" {
				return res.WithError(errors.New("expected traceId 'trace-abc-456'"))
			}
			if log.Data == nil || log.Data["key1"] != "value1" {
				return res.WithError(errors.New("expected data key1=value1"))
			}
		}
	}

	if !foundInfo {
		return res.WithError(errors.New("info log message not found"))
	}
	if !foundWarn {
		return res.WithError(errors.New("warn log message not found"))
	}

	return res
}

func (d *Requirement) readLogsWithMinLevel(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Read Logs With Min Level"}

	if d.createdID == "" {
		return res.WithError(errors.New("no entity created"))
	}

	// Read only logs at Warn level or higher
	logs, err := actor.Logs(context.Background(), d.createdID.String(), keystone.WithLogsMinLevel(proto.LogLevel_Warn))
	if err != nil {
		return res.WithError(err)
	}

	// Should have at least the Warn log we wrote
	if len(logs) < 1 {
		return res.WithError(errors.New("expected at least 1 log entry with min level Warn"))
	}

	// Verify no Info level logs are returned
	for _, log := range logs {
		if log.Level < proto.LogLevel_Warn {
			return res.WithError(errors.New("received log entry below min level Warn"))
		}
	}

	return res
}
