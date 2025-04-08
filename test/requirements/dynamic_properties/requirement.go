package dynamic_properties

import (
	"bytes"
	"context"
	"fmt"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"github.com/kubex/k4id"
	"time"
)

var (
	ValueString = "def"
	ValueNumber = int64(123)
	ValueBool   = true
	ValueFloat  = 1.23
	ValueTime   = time.Now()
	ValueAmount = keystone.NewAmount("GBP", 1000)
	ValueMap    = map[string]string{
		"aa": "11",
		"bb": "22",
		"cc": "33",
	}
	ValueStringList = []string{"a", "b", "c"}
	ValueIntList    = []int64{0, 1, 2, 3}

	//	Type (Text, Number, Bool, Float, Time, Amount, Map, Set, Aggregate)
)

type Requirement struct {
	createdID string
}

func (d *Requirement) Name() string {
	return "Dynamic Properties"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {

	usr := &models.User{
		ExternalID: k4id.New().String(),
	}
	mutateErr := actor.Mutate(context.Background(), usr)
	if mutateErr != nil {
		return []requirements.TestResult{requirements.NewResult("Unable to prepare test", mutateErr)}
	}
	d.createdID = usr.GetKeystoneID().String()

	return []requirements.TestResult{
		d.apply(actor),
		d.read(actor),
		d.readRange(actor),
		d.delete(actor),
	}
}

func (d *Requirement) apply(actor *keystone.Actor) requirements.TestResult {

	mutateErr := actor.SetDynamicProperties(context.Background(), d.createdID,
		keystone.NewDynamicProperties(map[string]interface{}{
			"v1_string":   ValueString,
			"v1_number":   ValueNumber,
			"v1_bool":     ValueBool,
			"v1_float":    ValueFloat,
			"v1_time":     ValueTime,
			"v1_amount":   ValueAmount,
			"v1_map":      ValueMap,
			"v1_str_list": ValueStringList,
			"v1_int_list": ValueIntList,
			"to_remove":   ValueNumber,
		}), nil, "Setting Dynamic Properties")

	return requirements.TestResult{
		Name:  "Create",
		Error: mutateErr,
	}
}

func (d *Requirement) read(actor *keystone.Actor) requirements.TestResult {

	res, err := actor.GetDynamicProperties(context.Background(), d.createdID,
		"v1_string",
		"v1_number",
		"v1_bool",
		"v1_float",
		"v1_time",
		"v1_amount",
		"v1_map",
		"v1_str_list",
		"v1_int_list",
	)

	if err == nil {
		if len(res) != 9 {
			err = fmt.Errorf("not all properties were returned, got %d expected 9", len(res))
		} else if res["v1_string"].Text != ValueString {
			err = fmt.Errorf("v1_string: got %s expected %s", res["v1_string"].Text, ValueString)
		} else if res["v1_number"].Int != ValueNumber {
			err = fmt.Errorf("v1_number: got %d expected %d", res["v1_number"].Int, ValueNumber)
		} else if res["v1_bool"].Bool != ValueBool {
			err = fmt.Errorf("v1_bool: got %t expected %t", res["v1_bool"].Bool, ValueBool)
		} else if res["v1_float"].Float != ValueFloat {
			err = fmt.Errorf("v1_float: got %f expected %f", res["v1_float"].Float, ValueFloat)
		} else if res["v1_time"].Time.AsTime().Unix() != ValueTime.Unix() {
			err = fmt.Errorf("v1_time: got %s expected %s", res["v1_time"].Time, ValueTime)
		} else if res["v1_amount"].Int != ValueAmount.Units || res["v1_amount"].Text != ValueAmount.Currency {
			err = fmt.Errorf("v1_amount: got %s %d expected %s %d", res["v1_amount"].Text, res["v1_amount"].Int, ValueAmount.Currency, ValueAmount.Units)
		} else if compareSlice(res["v1_str_list"].Array.GetStrings(), ValueStringList) != nil {
			err = fmt.Errorf("v1_str_list: got %v expected %v", res["v1_str_list"].Array.GetStrings(), ValueStringList)
		} else if compareSliceInt(res["v1_int_list"].Array.GetInts(), ValueIntList) != nil {
			err = fmt.Errorf("v1_int_list: got %v expected %v", res["v1_int_list"].Array.GetInts(), ValueIntList)
		} else if !compareMapStr(res["v1_map"].Array.GetKeyValue(), ValueMap) {
			err = fmt.Errorf("v1_map: got %v expected %v", res["v1_map"].Array.GetKeyValue(), ValueMap)
		} else if res["v1_map"].Array.GetKeyValue() == nil {
			err = fmt.Errorf("v1_map: got %v expected %v", res["v1_map"].Array.GetKeyValue(), ValueMap)
		}
	}

	return requirements.TestResult{
		Name:  "Read",
		Error: err,
	}
}
func (d *Requirement) readRange(actor *keystone.Actor) requirements.TestResult {

	res, err := actor.GetDynamicProperties(context.Background(), d.createdID,
		"v1_~",
	)

	if err == nil {
		if len(res) != 9 {
			err = fmt.Errorf("not all properties were returned, got %d expected 9", len(res))
		} else if res["v1_string"].Text != ValueString {
			err = fmt.Errorf("v1_string: got %s expected %s", res["v1_string"].Text, ValueString)
		} else if res["v1_number"].Int != ValueNumber {
			err = fmt.Errorf("v1_number: got %d expected %d", res["v1_number"].Int, ValueNumber)
		} else if res["v1_bool"].Bool != ValueBool {
			err = fmt.Errorf("v1_bool: got %t expected %t", res["v1_bool"].Bool, ValueBool)
		} else if res["v1_float"].Float != ValueFloat {
			err = fmt.Errorf("v1_float: got %f expected %f", res["v1_float"].Float, ValueFloat)
		} else if res["v1_time"].Time.AsTime().Unix() != ValueTime.Unix() {
			err = fmt.Errorf("v1_time: got %s expected %s", res["v1_time"].Time, ValueTime)
		} else if res["v1_amount"].Int != ValueAmount.Units || res["v1_amount"].Text != ValueAmount.Currency {
			err = fmt.Errorf("v1_amount: got %s %d expected %s %d", res["v1_amount"].Text, res["v1_amount"].Int, ValueAmount.Currency, ValueAmount.Units)
		} else if compareSlice(res["v1_set"].GetArray().GetStrings(), ValueStringList) != nil {
			err = fmt.Errorf("v1_set: got %v expected %v", res["v1_set"].GetArray().GetStrings(), ValueStringList)
		} else if compareSliceInt(res["v1_set"].GetArray().GetInts(), ValueIntList) != nil {
			err = fmt.Errorf("v1_set: got %v expected %v", res["v1_set"].GetArray().GetInts(), ValueIntList)
		} else if !compareMapStr(res["v1_map"].GetArray().GetKeyValue(), ValueMap) {
			err = fmt.Errorf("v1_map: got %v expected %v", res["v1_map"].GetArray().GetKeyValue(), ValueMap)
		} else if res["v1_map"].GetArray().GetKeyValue() == nil {
			err = fmt.Errorf("v1_map: got %v expected %v", res["v1_map"].GetArray().GetKeyValue(), ValueMap)
		}
	}

	return requirements.TestResult{
		Name:  "Read Range",
		Error: err,
	}
}

func (d *Requirement) delete(actor *keystone.Actor) requirements.TestResult {

	mutateErr := actor.SetDynamicProperties(context.Background(), d.createdID, nil, []string{"to_remove"}, "Removing Dynamic Properties")

	return requirements.TestResult{
		Name:  "Delete",
		Error: mutateErr,
	}
}

func compareMap(got, expect map[string][]byte) bool {
	if len(got) != len(expect) {
		return false
	}
	for k, gotV := range got {
		if expectV, found := expect[k]; !found || !bytes.Equal(gotV, expectV) {
			return false
		}
	}
	return true
}

func compareMapStr(got map[string][]byte, expect map[string]string) bool {
	if len(got) != len(expect) {
		return false
	}
	for k, gotV := range got {
		if expectV, found := expect[k]; !found || !matchAnyByte(expectV, gotV) {
			return false
		}
	}
	return true
}

func matchAnyByte(got any, expect []byte) bool {
	if str, ok := got.([]byte); ok {
		return bytes.Equal(str, expect)
	}
	if str, ok := got.(string); ok {
		return bytes.Equal([]byte(str), expect)
	}
	return false
}

func compareSlice(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func compareSliceInt(a, b []int64) []int64 {
	mb := make(map[int64]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []int64
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}
