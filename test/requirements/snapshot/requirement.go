package snapshot

import (
	"context"
	"errors"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"time"
)

var (
	String       = "snapshot"
	Integer      = int64(12355)
	Time         = time.Now()
	Amount       = keystone.NewAmount("USD", 130)
	Secret       = keystone.NewSecureString("secretval", "secre***")
	Verify       = keystone.NewVerifyString("toverify")
	MinMax       = keystone.NewMinMax(12, 18)
	Boolean      = true
	Float        = 19.85
	Map          = map[string]string{"key": "value", "key2": "value2"} // not setting
	StringSlice  = []string{"a", "b", "c", "c"}                        // not setting
	IntegerSlice = []int{1, 2, 3, 4, 4}                                // not setting
	StringSet    = keystone.NewStringSet("a", "b", "c", "c")           // not setting
	IntegerSet   = keystone.NewIntSet(1, 2, 3, 4, 4)                   // not setting
	RawData      = []byte("rawdata")
	EnumVal      = models.ENUM_VALUE1
	EnumVals     = []models.EnumValue{models.ENUM_VALUE0, models.ENUM_VALUE1}
	ExternalID   = keystone.NewExternalID("ven", "app", "etype", "external_id")
	MixedVal     = keystone.NewMixed(nil)
	MixedKey     = keystone.NewKeyMixed(nil)
)

type Requirement struct {
	createdID keystone.ID
}

func (d *Requirement) Name() string {
	return "Entity Reporting Snapshot"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.DataTypes{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	MixedVal.SetInt(12)
	MixedVal.SetString("stringval")
	MixedVal.SetBool(true)
	MixedVal.SetFloat(12.5)
	MixedVal.SetTime(time.Now())
	MixedVal.SetRaw([]byte("rawdata"))

	MixedKey.Set("first", MixedVal)
	MixedKey.Set("second", keystone.NewMixed("text"))

	return []requirements.TestResult{
		d.create(actor),
	}
}

func (d *Requirement) create(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Write Snapshot"}

	amnt := keystone.NewAmount(Amount.GetCurrency(), Amount.GetUnits())
	psn := &models.DataTypes{
		String:       String,
		Integer:      Integer,
		Time:         Time,
		Amount:       *amnt,
		AmountPt:     amnt,
		Secret:       Secret,
		Verify:       Verify,
		Boolean:      Boolean,
		MinMax:       MinMax,
		Float:        Float,
		Map:          Map,
		StringSlice:  StringSlice,
		IntegerSlice: IntegerSlice,
		StringSet:    StringSet,
		IntegerSet:   IntegerSet,
		RawData:      RawData,
		EnumValue:    EnumVal,
		Flags:        EnumVals,
		ExternalID:   ExternalID,
		Mixed:        MixedVal,
		MixedKey:     MixedKey,
	}

	createErr := actor.Mutate(context.Background(), psn, keystone.WithMutationComment("Create a default set"))
	if createErr == nil {
		d.createdID = psn.GetKeystoneID()
	} else {
		return res.WithError(createErr)
	}

	pass, err := actor.Snapshot(psn, psn.GetKeystoneID())
	if err != nil {
		return res.WithError(err)
	}
	if !pass {
		return res.WithError(errors.New("snapshot failed"))
	}
	return res
}
