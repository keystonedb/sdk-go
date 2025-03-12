package daily

import (
	"context"
	"errors"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"strconv"
	"time"
)

type Requirement struct {
	conn      *keystone.Connection
	createdID keystone.ID
}

func (d *Requirement) Name() string {
	return "Daily Entities"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	d.conn = conn
	conn.RegisterTypes(models.Person{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.create(actor),
		d.scanEntities(actor),
	}
}

func (d *Requirement) create(actor *keystone.Actor) requirements.TestResult {
	psn := &models.Person{
		BaseEntity: keystone.BaseEntity{},
		Name:       "Tom Daily",
	}

	createErr := actor.Mutate(context.Background(), psn, keystone.WithMutationComment("Known entity"))
	if createErr == nil {
		d.createdID = psn.GetKeystoneID()
	}

	return requirements.TestResult{
		Name:  "Create",
		Error: createErr,
	}
}

func (d *Requirement) scanEntities(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{
		Name: "Scan Entities",
	}

	if d.conn == nil {
		res.Error = errors.New("no connection available")
		return res
	}

	ks := d.conn.DirectClient()
	located := false
	fromId := ""
	loops := 0

	schema := &proto.Key{Key: keystone.Type(models.Person{}), Source: actor.VendorApp()}
	limit := int32(10)
	for {
		loops++
		entities, err := ks.DailyEntities(context.Background(), &proto.DailyEntityRequest{
			Authorization: actor.Authorization(),
			Schema:        schema,
			Date:          proto.CreateDate(time.Now()),
			AfterId:       fromId,
			Limit:         limit,
		})

		if err != nil {
			res.Error = err
			return res
		}

		fromId = entities.GetLastId()
		if len(entities.GetEntities()) == 0 {
			break
		}

		for _, eid := range entities.GetEntities() {
			if d.createdID.Matches(eid) {
				located = true
				break
			}
		}

		if located || len(entities.GetEntities()) < int(limit) {
			// Avoid doing an empty scan
			break
		}
	}

	res.Name += " (" + strconv.Itoa(loops) + ")"

	if !located {
		res.Error = errors.New("created entity not found")
	}

	return res
}
