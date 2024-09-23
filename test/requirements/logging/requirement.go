package logging

import (
	"context"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
)

type Requirement struct {
	createdID string
	labelKey  string
}

func (d *Requirement) Name() string {
	return "Logging"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.User{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.store(actor),
		d.retrieve(actor),
	}
}

func (d *Requirement) store(actor *keystone.Actor) requirements.TestResult {

	usr := &models.User{
		Validate: "logger",
	}

	usr.LogDebug("This is a debug message", "ref1", "actor", "trace-123", map[string]string{"key1": "value1"})
	usr.LogInfo("This is an info message", "ref1", "actor", "trace-123", map[string]string{"key1": "value1"})
	usr.LogNotice("This is a notice message", "ref1", "actor", "trace-123", map[string]string{"key1": "value1"})
	usr.LogWarn("This is a warning message", "ref1", "actor", "trace-123", map[string]string{"key1": "value1"})
	usr.LogError("This is an error message", "ref1", "actor", "trace-123", map[string]string{"key1": "value1"})
	usr.LogCritical("This is a critical message", "ref1", "actor", "trace-123", map[string]string{"key1": "value1"})
	usr.LogFatal("This is a fatal message", "ref1", "actor", "trace-123", map[string]string{"key1": "value1"})
	usr.LogAlert("This is an alert message", "ref1", "actor", "trace-123", map[string]string{"key1": "value1"})

	createErr := actor.Mutate(context.Background(), usr, keystone.WithMutationComment("Create a user with logs"))
	if createErr == nil {
		d.createdID = usr.GetKeystoneID()
	}

	return requirements.TestResult{
		Name:  "Create",
		Error: createErr,
	}
}

func (d *Requirement) retrieve(actor *keystone.Actor) requirements.TestResult {
	usr := &models.User{}
	getErr := actor.Get(context.Background(), keystone.ByEntityID(usr, d.createdID), usr, keystone.WithProperties())
	return requirements.TestResult{
		Name:  "Read",
		Error: getErr,
	}
}
