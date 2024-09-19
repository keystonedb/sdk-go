package keystone

import (
	"errors"
	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
)

func Test_UnmarshalNil(t *testing.T) {
	if err := Unmarshal(nil, struct{}{}); err != nil {
		t.Errorf("Unmarshal no data should not fail : %v", err)
	}

	if err := Unmarshal(make(map[Property]*proto.Value), nil); err != nil {
		t.Errorf("Unmarshal onto nil should not fail : %v", err)
	}
}

func Test_UnmarshalNonPointer(t *testing.T) {

	inputData := map[Property]*proto.Value{NewProperty("ID"): {Text: "xx"}}
	e := testEntity{}
	err := Unmarshal(inputData, e)
	if !errors.Is(err, ErrMustPassPointer) {
		t.Errorf("Unmarshal should fail with ErrMustPassPointer")
	}
}

func Test_Unmarshal(t *testing.T) {

	inputData := map[Property]*proto.Value{
		NewProperty("ID"):                          {Text: "xx"},
		NewProperty("Name"):                        {Text: "John"},
		NewProperty("Age"):                         {Int: 30},
		NewProperty("DOB"):                         {Time: timestamppb.New(time.Date(2009, 2, 13, 23, 31, 30, 12345, time.UTC))},
		NewProperty("HasGlasses"):                  {Bool: true},
		NewProperty("FraudScore"):                  {Float: 0.5},
		NewProperty("AmountPaid"):                  {Text: "USD", Int: 142},
		NewPrefixProperty("sub_struct", "SubName"): {Text: "AAA"},
	}

	e := testEntity{}
	err := Unmarshal(inputData, &e)
	if err != nil {
		t.Errorf("Marshal failed: %v", err)
	}
	if e.ID != "xx" {
		t.Errorf("ID,TEXT not set correctly")
	}
	if e.Name != "John" {
		t.Errorf("TEXT not set correctly")
	}
	if e.Age != 30 {
		t.Errorf("INT not set correctly")
	}
	if e.DOB != time.Date(2009, 2, 13, 23, 31, 30, 12345, time.UTC) {
		t.Errorf("TIME not set correctly")
	}
	if e.HasGlasses != true {
		t.Errorf("BOOL not set correctly")
	}
	if e.FraudScore != 0.5 {
		t.Errorf("FLOAT not set correctly")
	}
	if e.AmountPaid.Currency != "USD" || e.AmountPaid.Units != 142 {
		t.Errorf("AmountPaid not set correctly")
	}
	if e.SubStruct.SubName != "AAA" {
		t.Errorf("SubStruct not set correctly")
	}
}
