package remote

import (
	"context"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"time"
)

const vendor2ID = "ven2"
const app2ID = "remote"

type Requirement struct {
	entityID         keystone.ID
	psn              *models.Person
	secondConnection *keystone.Connection
	secondActor      *keystone.Actor
}

type RemotePerson struct {
	keystone.Remote
}

func (d *Requirement) Name() string {
	return "Remote Entity"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	d.secondConnection = keystone.NewConnection(actor.Connection().DirectClient(), vendor2ID, app2ID, "test-access-token")
	act2 := d.secondConnection.Actor("tt", "127.0.0.2", "random-userid", "UserAgent")
	d.secondActor = &act2
	return []requirements.TestResult{
		d.prepare(actor),
		d.log(actor),
		d.upload(actor),
	}
}

func (d *Requirement) prepare(actor *keystone.Actor) requirements.TestResult {
	d.psn = &models.Person{
		BaseEntity:   keystone.BaseEntity{},
		Name:         "John",
		HeightInCm:   123,
		DOB:          time.Now(),
		BankBalance:  *keystone.NewAmount("USD", 345),
		FullName:     keystone.NewSecureString("John Doe", "Jo*** D***"),
		AccountPin:   keystone.NewVerifyString("1234"),
		SecretAnswer: keystone.NewSecureString("Pet Name", "Pe*******"),
	}
	mutateErr := actor.Mutate(context.Background(), d.psn)
	d.entityID = d.psn.GetKeystoneID()

	return requirements.TestResult{
		Name:  "Prepare Remote Entity",
		Error: mutateErr,
	}
}

func (d *Requirement) log(actor *keystone.Actor) requirements.TestResult {
	psn := keystone.RemoteEntity(d.entityID)
	psn.LogInfo("This is an info message", "ref1", "actor", "trace-123", map[string]string{"key1": "value1"})
	remoteMutateErr := d.secondActor.RemoteMutate(context.Background(), d.entityID, psn)

	if remoteMutateErr == nil {
		//TODO: Load Logs
	}

	return requirements.TestResult{
		Name:  "Log Remote Entity",
		Error: remoteMutateErr,
	}
}

func (d *Requirement) upload(actor *keystone.Actor) requirements.TestResult {

	psn := keystone.RemoteEntity(d.entityID)

	file1 := keystone.NewUpload("abc", proto.ObjectType_Standard)
	file1.SetData([]byte("Hello World"))

	remoteMutateErr := psn.Mutate(context.Background(), d.secondActor, keystone.PrepareUploads(file1))

	return requirements.TestResult{
		Name:  "Upload to Remote Entity",
		Error: remoteMutateErr,
	}
}
