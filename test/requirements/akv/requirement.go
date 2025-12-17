package akv

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/keystonedb/sdk-go/test/requirements"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var rawVal = &proto.Value{
	Text:       "random text value",
	Int:        1342,
	Bool:       true,
	Float:      12.54,
	Time:       timestamppb.New(time.Now()),
	SecureText: "secure text",
	Raw:        []byte("abc123"),
	Array: &proto.RepeatedValue{
		KeyValue: map[string][]byte{
			"val1": []byte("123"),
			"val2": []byte("abc"),
		},
		Strings: []string{"val1", "v2", "v3", "v3"},
		Ints:    []int64{1, 2, 3, 3},
	},
}

type Requirement struct{}

func (d *Requirement) Name() string {
	return "App Key Value"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.put(actor),
		d.get(actor),
		d.putRaw(actor),
		d.getRaw(actor),
	}
}

func (d *Requirement) put(actor *keystone.Actor) requirements.TestResult {

	putResp, putErr := actor.AKVPut(context.Background(), keystone.AKV("val1", 123), keystone.AKV("val2", "abc"))

	if putErr == nil {
		if !putResp.Success {
			putErr = fmt.Errorf("%d - %s", putResp.GetErrorCode(), putResp.ErrorMessage)
		}
	}

	return requirements.TestResult{
		Name:  "Put",
		Error: putErr,
	}
}
func (d *Requirement) get(actor *keystone.Actor) requirements.TestResult {

	resp, getErr := actor.AKVGet(context.Background(), "val1", "val2", "val3")

	if getErr == nil {
		if val, hasVal := resp["val1"]; !hasVal {
			getErr = errors.New("val1 not found")
		} else if val.GetInt() != 123 {
			getErr = errors.New("val1 has wrong value")
		}

		if val, hasVal := resp["val2"]; !hasVal {
			getErr = errors.New("val2 not found")
		} else if val.GetText() != "abc" {
			getErr = errors.New("val2 has wrong value")
		}

		if _, hasVal := resp["val3"]; hasVal {
			getErr = errors.New("val3 should not be found")
		}
	}

	return requirements.TestResult{
		Name:  "Get",
		Error: getErr,
	}
}

func (d *Requirement) putRaw(actor *keystone.Actor) requirements.TestResult {

	putResp, putErr := actor.AKVPut(context.Background(), keystone.AKVRaw("rawval", rawVal))

	if putErr == nil {
		if !putResp.Success {
			putErr = fmt.Errorf("%d - %s", putResp.GetErrorCode(), putResp.ErrorMessage)
		}
	}

	return requirements.TestResult{
		Name:  "Put Raw",
		Error: putErr,
	}
}

func (d *Requirement) getRaw(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "GetRaw"}
	resp, getErr := actor.AKVGet(context.Background(), "rawval")
	if getErr != nil {
		return res.WithError(getErr)
	}

	retRaw, ok := resp["rawval"]
	if !ok {
		return res.WithError(errors.New("rawval not found"))
	}
	if retRaw.Text != rawVal.Text {
		return res.WithError(errors.New("rawval has wrong text value"))
	}
	if retRaw.Int != rawVal.Int {
		return res.WithError(errors.New("rawval has wrong int value"))
	}
	if retRaw.Bool != rawVal.Bool {
		return res.WithError(errors.New("rawval has wrong bool value"))
	}
	if retRaw.Float != rawVal.Float {
		return res.WithError(errors.New("rawval has wrong float value"))
	}
	if retRaw.Time.AsTime().Unix() != rawVal.Time.AsTime().Unix() {
		return res.WithError(errors.New("rawval has wrong time value"))
	}
	if retRaw.SecureText != rawVal.SecureText {
		return res.WithError(errors.New("rawval has wrong secure text value"))
	}
	if !bytes.Equal(retRaw.Raw, rawVal.Raw) {
		return res.WithError(errors.New("rawval has wrong raw value"))
	}

	if retRaw.Array == nil {
		return res.WithError(errors.New("rawval has no array value"))
	}

	if len(retRaw.Array.KeyValue) != len(rawVal.Array.KeyValue) {
		return res.WithError(errors.New("rawval has wrong array key value length"))
	}
	if len(retRaw.Array.Strings) != len(rawVal.Array.Strings) {
		return res.WithError(errors.New("rawval has wrong array strings length"))
	}
	if len(retRaw.Array.Ints) != len(rawVal.Array.Ints) {
		return res.WithError(errors.New("rawval has wrong array ints length"))
	}

	return res
}
