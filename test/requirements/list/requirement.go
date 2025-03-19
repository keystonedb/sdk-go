package list

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"github.com/kubex/k4id"
	"time"
)

type Requirement struct {
	runID string
}

func (d *Requirement) Name() string {
	return "List Entities"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	d.runID = k4id.New().String()
	conn.RegisterTypes(models.FileData{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.create(actor),
		d.readPending(actor),
		d.readOneTwo(actor),
		d.readThree(actor),
		d.readComplex(actor),
	}
}

func (d *Requirement) readOneTwo(actor *keystone.Actor) requirements.TestResult {

	entities, err := actor.List(context.Background(), keystone.Type(models.FileData{}),
		[]string{"check_key", "line_information"}, keystone.Limit(2, 0), keystone.SortBy("modified", true),
		keystone.WhereIn("state", 1, 2))

	if err == nil && len(entities) < 2 {
		err = errors.New("not enough entities returned")
	}

	for _, entity := range entities {
		file := &models.FileData{}
		unErr := keystone.Unmarshal(entity, file)
		if unErr != nil {
			err = unErr
			break
		}

		if file.CheckKey != d.runID {
			err = errors.New("incorrect check key - " + file.CheckKey + " != " + d.runID)
			break
		}

		if file.LineInformation == "" {
			err = errors.New("line information is empty")
			break
		}
	}

	return requirements.TestResult{
		Name:  "readOneTwo",
		Error: err,
	}
}
func (d *Requirement) readThree(actor *keystone.Actor) requirements.TestResult {

	entities, err := actor.List(context.Background(), keystone.Type(models.FileData{}),
		[]string{"check_key"}, keystone.Limit(2, 0), keystone.SortBy("modified", true),
		keystone.WhereEquals("connector_id", "tester"))
	if err == nil && len(entities) < 2 {
		err = errors.New("not enough entities returned")
	}

	for _, entity := range entities {
		file := &models.FileData{}
		unErr := keystone.Unmarshal(entity, file)
		if unErr != nil {
			err = unErr
			break
		}

		if file.CheckKey != d.runID {
			err = errors.New("incorrect check key - " + file.CheckKey + " != " + d.runID)
			break
		}
	}

	return requirements.TestResult{
		Name:  "readThree",
		Error: err,
	}
}

func (d *Requirement) readComplex(actor *keystone.Actor) requirements.TestResult {

	entities, err := actor.List(context.Background(), keystone.Type(models.FileData{}),
		[]string{"check_key", "identifier"}, keystone.Limit(2, 0), keystone.SortBy("modified", true),
		keystone.WhereEquals("user_id", "usr1"),
		keystone.Or(keystone.WhereEquals("state", 1), keystone.WhereEquals("state", 2)))
	if err == nil && len(entities) < 2 {
		err = errors.New("not enough entities returned")
	}

	expect := map[string]bool{
		"fd0": false,
		"fd1": false,
	}

	for _, entity := range entities {
		file := &models.FileData{}
		unErr := keystone.Unmarshal(entity, file)
		if unErr != nil {
			err = unErr
			break
		}

		if file.CheckKey != d.runID {
			err = errors.New("incorrect check key - " + file.CheckKey + " != " + d.runID)
			break
		}

		expect[file.Identifier] = true
	}

	for k, v := range expect {
		if !v {
			err = errors.New("expected identifier not found - " + k)
			break
		}
	}

	return requirements.TestResult{
		Name:  "readComplex",
		Error: err,
	}
}

func (d *Requirement) readPending(actor *keystone.Actor) requirements.TestResult {

	entities, err := actor.List(context.Background(), keystone.Type(models.FileData{}),
		[]string{"check_key"}, keystone.Limit(3, 0), keystone.SortBy("modified", true),
		keystone.WhereEquals("is_pending", true))

	if err == nil && len(entities) < 3 {
		err = errors.New("not enough entities returned")
	}

	for _, entity := range entities {
		file := &models.FileData{}
		unErr := keystone.Unmarshal(entity, file)
		if unErr != nil {
			err = unErr
			break
		}

		if file.CheckKey != d.runID {
			err = errors.New("incorrect check key - " + file.CheckKey + " != " + d.runID)
			break
		}
	}

	return requirements.TestResult{
		Name:  "readPending",
		Error: err,
	}
}
func (d *Requirement) create(actor *keystone.Actor) requirements.TestResult {
	files := []*models.FileData{
		{
			UserID:          "usr1",
			Submitted:       time.Now(),
			State:           1,
			ConnectorID:     "connector-1",
			IsPending:       false,
			CheckKey:        d.runID,
			LineInformation: uuid.NewString(),
			Identifier:      "fd0",
		},
		{
			UserID:          "usr1",
			Submitted:       time.Now(),
			State:           2,
			ConnectorID:     "connector-1",
			IsPending:       false,
			CheckKey:        d.runID,
			LineInformation: uuid.NewString(),
			Identifier:      "fd1",
		},
		{
			UserID:          "usr1",
			Submitted:       time.Now(),
			State:           3,
			ConnectorID:     "connector-1",
			IsPending:       false,
			CheckKey:        d.runID,
			LineInformation: uuid.NewString(),
			Identifier:      "fd2",
		},
		{
			UserID:          "usr1",
			Submitted:       time.Now(),
			State:           3,
			ConnectorID:     "connector-1",
			IsPending:       true,
			CheckKey:        d.runID,
			LineInformation: uuid.NewString(),
			Identifier:      "fd3",
		},
		{
			UserID:          "usr1",
			Submitted:       time.Now(),
			State:           3,
			ConnectorID:     "tester",
			IsPending:       true,
			CheckKey:        d.runID,
			LineInformation: uuid.NewString(),
			Identifier:      "fd4",
		},
		{
			UserID:          "usr1",
			Submitted:       time.Now(),
			State:           3,
			ConnectorID:     "tester",
			IsPending:       true,
			CheckKey:        d.runID,
			LineInformation: uuid.NewString(),
			Identifier:      "fd5",
		},
	}

	var createErr error
	for _, file := range files {
		createErr = actor.Mutate(context.Background(), file, keystone.WithMutationComment("Create a file"))
		if createErr != nil {
			return requirements.TestResult{
				Name:  "Create",
				Error: createErr,
			}
		}
	}

	return requirements.TestResult{
		Name:  "Create",
		Error: createErr,
	}
}
