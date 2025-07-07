package dynamic_entity

import (
	"bytes"
	"context"
	"errors"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"log"
	"reflect"
	"strings"
	"time"
)

var (
	String       = "stringval"
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
	return "Remote Entity Properties"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
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
		d.read(actor),
		d.append(actor),
		d.readPostAppend(actor),
		d.reduce(actor),
		d.readPostReduce(actor),
	}
}

func (d *Requirement) create(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Create"}
	src := &models.Person{Name: "Remote Entity Properties Src"}
	createErr := actor.Mutate(context.Background(), src, keystone.WithMutationComment("Create a default set"))
	if createErr == nil {
		d.createdID = src.GetKeystoneID()
	} else {
		return res.WithError(createErr)
	}

	amnt := keystone.NewAmount(Amount.GetCurrency(), Amount.GetUnits())
	psn := &models.DynamicRemote{
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

	createErr = actor.RemoteMutate(context.Background(), d.createdID, psn, keystone.WithMutationComment("Create a default set"))
	return res.WithError(createErr)
}

func (d *Requirement) read(actor *keystone.Actor) requirements.TestResult {

	dt := &models.DynamicRemote{}
	log.Println("EID: ", d.createdID)
	getErr := actor.RemoteGetProperties(context.Background(), d.createdID, dt)
	if getErr == nil {
		if dt.String != String {
			getErr = errors.New("string mismatch")
		} else if dt.Integer != Integer {
			getErr = errors.New("integer mismatch")
		} else if dt.Time.Unix() != Time.Unix() {
			getErr = errors.New("time mismatch")
		} else if !Amount.Equals(&dt.Amount) {
			getErr = errors.New("amount mismatch")
		} else if !Amount.Equals(dt.AmountPt) {
			getErr = errors.New("amount mismatch")
		} else if dt.Amount.GetUnits() != 130 {
			getErr = errors.New("amount mismatch, expected 130")
		} else if dt.Secret.Original != Secret.Original {
			getErr = errors.New("secret original mismatch")
		} else if dt.Secret.Masked != Secret.Masked {
			getErr = errors.New("secret Masked mismatch")
		} else if dt.Boolean != Boolean {
			getErr = errors.New("boolean mismatch")
		} else if dt.MinMax != MinMax {
			getErr = errors.New("MinMax mismatch")
		} else if dt.Float != Float {
			getErr = errors.New("float mismatch")
		} else if !reflect.DeepEqual(dt.Map, Map) {
			getErr = errors.New("map mismatch")
		} else if !reflect.DeepEqual(dt.StringSlice, StringSlice) {
			getErr = errors.New("StringSlice mismatch")
		} else if !reflect.DeepEqual(dt.IntegerSlice, IntegerSlice) {
			getErr = errors.New("IntegerSlice mismatch")
		} else if len(dt.StringSet.Diff(StringSet.Values()...)) != 0 {
			getErr = errors.New("StringSet mismatch" + strings.Join(dt.StringSet.Diff(StringSet.Values()...), ","))
		} else if len(dt.IntegerSet.Diff(IntegerSet.Values()...)) != 0 {
			getErr = errors.New("IntegerSet mismatch")
		} else if !bytes.Equal(dt.RawData, RawData) {
			getErr = errors.New("RawData mismatch")
		} else if dt.EnumValue != EnumVal {
			getErr = errors.New("EnumValue mismatch")
		} else if !reflect.DeepEqual(dt.Flags, EnumVals) {
			getErr = errors.New("flags mismatch")
		} else if dt.ExternalID.String() != ExternalID.String() {
			getErr = errors.New("ExternalID mismatch")
		} else if !dt.Mixed.Matches(&MixedVal) {
			getErr = errors.New("MixedVal mismatch")
		} else if len(dt.MixedKey.Diff(MixedKey.Values())) != 0 {
			getErr = errors.New("MixedKey mismatch")
		}
	}

	return requirements.TestResult{
		Name:  "Read",
		Error: getErr,
	}
}

func (d *Requirement) append(actor *keystone.Actor) requirements.TestResult {
	psn := &models.DynamicRemote{}
	psn.IntegerSet.Add(7)
	psn.SetKeystoneID(d.createdID)

	psn.MixedKey.Append("third", keystone.NewMixed("newval"))

	updateErr := actor.RemoteMutate(context.Background(), d.createdID, psn, keystone.WithMutationComment("Update a person"))

	return requirements.TestResult{
		Name:  "Update",
		Error: updateErr,
	}
}

func (d *Requirement) readPostAppend(actor *keystone.Actor) requirements.TestResult {

	dt := &models.DynamicRemote{}
	getErr := actor.RemoteGetProperties(context.Background(), d.createdID, dt, "integer_set", "mixed_key")
	if getErr == nil {
		mixCheck := keystone.NewMixed("newval")
		if !dt.IntegerSet.Has(7) {
			getErr = errors.New("IntegerSet did not append 7")
		} else if !dt.IntegerSet.Has(1) {
			getErr = errors.New("IntegerSet did not return 1")
		} else if !dt.MixedKey.Has("third") {
			getErr = errors.New("MixedKey did not append third")
		} else if !dt.MixedKey.Get("third").Matches(&mixCheck) {
			getErr = errors.New("MixedKey did not append third")
		}
	}

	return requirements.TestResult{
		Name:  "Read After Append",
		Error: getErr,
	}
}

func (d *Requirement) reduce(actor *keystone.Actor) requirements.TestResult {
	psn := &models.DynamicRemote{}
	psn.IntegerSet.Remove(2)
	psn.MixedKey.Remove("third")
	psn.SetKeystoneID(d.createdID)
	updateErr := actor.RemoteMutate(context.Background(), d.createdID, psn, keystone.WithMutationComment("Update a person"))

	return requirements.TestResult{
		Name:  "Update",
		Error: updateErr,
	}
}

func (d *Requirement) readPostReduce(actor *keystone.Actor) requirements.TestResult {

	dt := &models.DynamicRemote{}
	getErr := actor.RemoteGetProperties(context.Background(), d.createdID, dt, "integer_set", "mixed_key")
	if getErr == nil {
		if dt.IntegerSet.Has(2) {
			getErr = errors.New("IntegerSet did not remove 2")
		} else if !dt.IntegerSet.Has(1) {
			getErr = errors.New("IntegerSet did not return 1")
		} else if dt.MixedKey.Has("third") {
			getErr = errors.New("MixedKey did not remove third")
		}
	}

	return requirements.TestResult{
		Name:  "Read After Reduce",
		Error: getErr,
	}
}
