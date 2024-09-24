package keystone

import (
	"errors"
	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"testing"
	"time"
)

func Test_UnmarshalPropertiesNil(t *testing.T) {
	if err := UnmarshalProperties(nil, struct{}{}); err != nil {
		t.Errorf("UnmarshalProperties no data should not fail : %v", err)
	}

	if err := UnmarshalProperties(make(map[Property]*proto.Value), nil); err != nil {
		t.Errorf("UnmarshalProperties onto nil should not fail : %v", err)
	}
}

func Test_UnmarshalNil(t *testing.T) {
	if err := Unmarshal(nil, struct{}{}); err != nil {
		t.Errorf("Unmarshal no data should not fail : %v", err)
	}

	if err := Unmarshal(nil, nil); err != nil {
		t.Errorf("Unmarshal nil to nil should not fail : %v", err)
	}

	if err := Unmarshal(&proto.EntityResponse{}, nil); err != nil {
		t.Errorf("Unmarshal onto nil should not fail : %v", err)
	}
}

func Test_UnmarshalNonPointer(t *testing.T) {

	inputData := map[Property]*proto.Value{NewProperty("ID"): {Text: "xx"}}
	e := testEntity{}
	err := UnmarshalProperties(inputData, e)
	if !errors.Is(err, ErrMustPassPointer) {
		t.Errorf("UnmarshalProperties should fail with ErrMustPassPointer")
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
	err := UnmarshalProperties(inputData, &e)
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

func Test_UnmarshalOntoUnsupported(t *testing.T) {
	sliceAny := []any{"x"}
	sliceInt := []int{1}
	sliceStruct := []testEntity{{}}
	mapInt := map[string]int{"x": 1}
	mapAny := map[string]any{"x": 1}
	mapInterface := map[string]interface{}{"x": 1}

	tests := []struct {
		name string
		dst  any
	}{
		{"nil", nil},
		{"non-pointer", sliceAny},
		{"pointer", &sliceAny},
		{"non-pointer int", sliceInt},
		{"pointer int", &sliceInt},
		{"struct slice", sliceStruct},
		{"map int", mapInt},
		{"map any", mapAny},
		{"map interface", mapInterface},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := UnmarshalToSlice(test.dst, &proto.EntityResponse{})
			if !errors.Is(err, ErrMustPointerSlice) {
				t.Errorf("UnmarshalToSlice should fail with ErrMustPointerSlice, received %v", err)
			}
		})
	}
}

func Test_UnmarshalOntoNoItems(t *testing.T) {
	sliceStruct := []testEntity{{}}
	err := UnmarshalToSlice(&sliceStruct)
	if err != nil {
		t.Errorf("UnmarshalToSlice should not fail with no items, got %v", err)
	}
}

func Test_UnmarshalToMapUnsupported(t *testing.T) {
	sliceAny := []any{"x"}
	sliceInt := []int{1}
	sliceStruct := []testEntity{{}}
	mapInt := map[string]int{"x": 1}
	mapAny := map[string]any{"x": 1}
	mapInterface := map[string]interface{}{"x": 1}

	tests := []struct {
		name string
		dst  any
	}{
		{"nil", nil},
		{"non-pointer", sliceAny},
		{"pointer", &sliceAny},
		{"non-pointer int", sliceInt},
		{"pointer int", &sliceInt},
		{"struct slice", sliceStruct},
		{"struct slice pointer", &sliceStruct},
		{"map int", mapInt},
		{"map any", mapAny},
		{"map interface", mapInterface},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := UnmarshalToMap(test.dst, &proto.EntityResponse{})
			if !errors.Is(err, ErrMustMapStringStruct) {
				t.Errorf("UnmarshalToMap should fail with ErrMustMapStringStruct, received %v", err)
			}
		})
	}
}

func Test_UnmarshalToMapNoItems(t *testing.T) {
	mapStruct := make(map[string]testEntity)
	err := UnmarshalToMap(mapStruct)
	if err != nil {
		t.Errorf("UnmarshalToMap should not fail with no items, got %v", err)
	}
}
func Test_UnmarshalToMapNilMap(t *testing.T) {
	var nilMpStruct map[string]testEntity
	err := UnmarshalToMap(nilMpStruct, &proto.EntityResponse{})
	if !errors.Is(err, ErrNilMapGiven) {
		t.Errorf("UnmarshalToMap should fail with ErrNilMapGiven, received %v", err)
	}
}

func Test_UnmarshalToMapBadEntity(t *testing.T) {

	mapStruct := make(map[string]testEntity, 0)
	err := UnmarshalToMap(mapStruct, &proto.EntityResponse{})
	if !errors.Is(err, ErrEntityIDMissing) {
		t.Errorf("UnmarshalToMap should fail with ErrEntityIDMissing, received %v", err)
	}
}

func Test_UnmarshalToMap(t *testing.T) {
	responses := []*proto.EntityResponse{
		{Entity: &proto.Entity{EntityId: "abc"}, Properties: []*proto.EntityProperty{{Property: "name", Value: &proto.Value{Text: "nma"}}}},
		{Entity: &proto.Entity{EntityId: "123"}, Properties: []*proto.EntityProperty{{Property: "name", Value: &proto.Value{Text: "nm1"}}}},
		{Entity: &proto.Entity{EntityId: "x-y"}, Properties: []*proto.EntityProperty{{Property: "name", Value: &proto.Value{Text: "nmx"}}}},
	}

	mapStruct := make(map[string]testEntity)
	err := UnmarshalToMap(mapStruct, responses...)
	if err != nil {
		t.Errorf("UnmarshalToMap failed: %v", err)
	}
	if len(mapStruct) != 3 {
		t.Fatalf("UnmarshalToMap did not set all items")
	}

	if mapStruct["abc"].Name != "nma" {
		t.Errorf("UnmarshalToMap failed to set name for abc")
	}
	if mapStruct["123"].Name != "nm1" {
		t.Errorf("UnmarshalToMap failed to set name for 123")
	}
	if mapStruct["x-y"].Name != "nmx" {
		t.Errorf("UnmarshalToMap failed to set name for x-y")
	}
	log.Println(mapStruct)
}

func Test_UnmarshalToMapPointer(t *testing.T) {
	responses := []*proto.EntityResponse{
		{Entity: &proto.Entity{EntityId: "abc"}, Properties: []*proto.EntityProperty{{Property: "name", Value: &proto.Value{Text: "nma"}}}},
		{Entity: &proto.Entity{EntityId: "123"}, Properties: []*proto.EntityProperty{{Property: "name", Value: &proto.Value{Text: "nm1"}}}},
		{Entity: &proto.Entity{EntityId: "x-y"}, Properties: []*proto.EntityProperty{{Property: "name", Value: &proto.Value{Text: "nmx"}}}},
	}

	mapStruct := make(map[string]*testEntity)
	err := UnmarshalToMap(mapStruct, responses...)
	if err != nil {
		t.Errorf("UnmarshalToMap failed: %v", err)
	}
	if len(mapStruct) != 3 {
		t.Fatalf("UnmarshalToMap did not set all items")
	}

	if mapStruct["abc"].Name != "nma" {
		t.Errorf("UnmarshalToMap failed to set name for abc")
	}
	if mapStruct["123"].Name != "nm1" {
		t.Errorf("UnmarshalToMap failed to set name for 123")
	}
	if mapStruct["x-y"].Name != "nmx" {
		t.Errorf("UnmarshalToMap failed to set name for x-y")
	}
}

func Test_UnmarshalSlice(t *testing.T) {

	var sliceStruct []testEntity
	responses := []*proto.EntityResponse{
		{Entity: &proto.Entity{EntityId: "abc"}, Properties: []*proto.EntityProperty{{Property: "name", Value: &proto.Value{Text: "nma"}}}},
		{Entity: &proto.Entity{EntityId: "123"}, Properties: []*proto.EntityProperty{{Property: "name", Value: &proto.Value{Text: "nm1"}}}},
		{Entity: &proto.Entity{EntityId: "x-y"}, Properties: []*proto.EntityProperty{{Property: "name", Value: &proto.Value{Text: "nmx"}}}},
	}

	err := UnmarshalToSlice(&sliceStruct, responses...)
	if err != nil {
		t.Errorf("UnmarshalToSlice failed: %v", err)
	}
	if len(sliceStruct) != 3 {
		t.Fatalf("UnmarshalToSlice did not set all items")
	}

	seenABC := false
	seen123 := false
	seenXY := false

	for _, v := range sliceStruct {
		if v.Name == "nma" {
			seenABC = true
		}
		if v.Name == "nm1" {
			seen123 = true
		}
		if v.Name == "nmx" {
			seenXY = true
		}
	}
	if !seenABC || !seen123 || !seenXY {
		t.Errorf("UnmarshalToSlice did not set all items")
	}
}
func Test_UnmarshalSlicePointer(t *testing.T) {

	var sliceStruct []*testEntity
	responses := []*proto.EntityResponse{
		{Entity: &proto.Entity{EntityId: "abc"}, Properties: []*proto.EntityProperty{{Property: "name", Value: &proto.Value{Text: "nma"}}}},
		{Entity: &proto.Entity{EntityId: "123"}, Properties: []*proto.EntityProperty{{Property: "name", Value: &proto.Value{Text: "nm1"}}}},
		{Entity: &proto.Entity{EntityId: "x-y"}, Properties: []*proto.EntityProperty{{Property: "name", Value: &proto.Value{Text: "nmx"}}}},
	}

	err := UnmarshalToSlice(&sliceStruct, responses...)
	if err != nil {
		t.Errorf("UnmarshalToSlice failed: %v", err)
	}
	if len(sliceStruct) != 3 {
		t.Fatalf("UnmarshalToSlice did not set all items")
	}

	seenABC := false
	seen123 := false
	seenXY := false

	for _, v := range sliceStruct {
		if v.Name == "nma" {
			seenABC = true
		}
		if v.Name == "nm1" {
			seen123 = true
		}
		if v.Name == "nmx" {
			seenXY = true
		}
	}
	if !seenABC || !seen123 || !seenXY {
		t.Errorf("UnmarshalToSlice did not set all items")
	}
}

func Test_AsSlice(t *testing.T) {

	responses := []*proto.EntityResponse{
		{Entity: &proto.Entity{EntityId: "abc"}, Properties: []*proto.EntityProperty{{Property: "name", Value: &proto.Value{Text: "nma"}}}},
		{Entity: &proto.Entity{EntityId: "123"}, Properties: []*proto.EntityProperty{{Property: "name", Value: &proto.Value{Text: "nm1"}}}},
		{Entity: &proto.Entity{EntityId: "x-y"}, Properties: []*proto.EntityProperty{{Property: "name", Value: &proto.Value{Text: "nmx"}}}},
	}

	sliceStruct, err := AsSlice[testEntity](responses...)
	if err != nil {
		t.Errorf("AsSlice failed: %v", err)
	}
	if len(sliceStruct) != 3 {
		t.Fatalf("AsSlice did not set all items")
	}

	seenABC := false
	seen123 := false
	seenXY := false

	for _, v := range sliceStruct {
		if v.Name == "nma" {
			seenABC = true
		}
		if v.Name == "nm1" {
			seen123 = true
		}
		if v.Name == "nmx" {
			seenXY = true
		}
	}
	if !seenABC || !seen123 || !seenXY {
		t.Errorf("AsSlice did not set all items")
	}
}

func Test_New(t *testing.T) {
	entity := &proto.EntityResponse{
		Entity: &proto.Entity{EntityId: "abc"}, Properties: []*proto.EntityProperty{{Property: "name", Value: &proto.Value{Text: "nma"}}},
	}

	res, err := New[testEntity](entity)
	if err != nil {
		t.Errorf("New failed: %v", err)
	}
	if res.Name != "nma" {
		t.Errorf("New did not set name")
	}
}
