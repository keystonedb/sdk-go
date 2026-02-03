package reflector

import (
	"reflect"
	"testing"
)

func TestIntMap(t *testing.T) {
	input := map[string]int{
		"foo": 123,
		"bar": -456,
	}

	p := &IntMap{}
	ref := reflect.ValueOf(input)
	pVal, err := p.ToProto(ref)
	if err != nil {
		t.Fatal(err)
	}

	if pVal.Array == nil {
		t.Fatal("Expected non-nil array")
	}

	if string(pVal.Array.KeyValue["foo"]) != "123" {
		t.Errorf("expected 123, got %s", pVal.Array.KeyValue["foo"])
	}
	if string(pVal.Array.KeyValue["bar"]) != "-456" {
		t.Errorf("expected -456, got %s", pVal.Array.KeyValue["bar"])
	}

	refVal := reflect.ValueOf(new(map[string]int)).Elem()
	err = p.SetValue(pVal, refVal)
	if err != nil {
		t.Fatalf("IntMap.SetValue returned error: %v", err)
	}
	output, ok := refVal.Interface().(map[string]int)
	if !ok {
		t.Fatal("IntMap.SetValue didn't return a map[string]int")
	}
	if output["foo"] != 123 || output["bar"] != -456 {
		t.Errorf("output mismatch: %v", output)
	}
}

type extendedIntMap int32

func TestIntMap_Uint(t *testing.T) {
	input := map[string]uint64{
		"foo": 123,
	}

	p := &IntMap{}
	pVal, err := p.ToProto(reflect.ValueOf(input))
	if err != nil {
		t.Fatal(err)
	}

	refVal := reflect.ValueOf(new(map[string]uint64)).Elem()
	err = p.SetValue(pVal, refVal)
	if err != nil {
		t.Fatalf("IntMap.SetValue returned error: %v", err)
	}
	output := refVal.Interface().(map[string]uint64)
	if output["foo"] != 123 {
		t.Errorf("expected 123, got %d", output["foo"])
	}
}

func TestIntMap_Ext(t *testing.T) {
	input := map[string]extendedIntMap{
		"foo": 789,
	}

	p := &IntMap{}
	pVal, err := p.ToProto(reflect.ValueOf(input))
	if err != nil {
		t.Fatal(err)
	}

	refVal := reflect.ValueOf(new(map[string]extendedIntMap)).Elem()
	err = p.SetValue(pVal, refVal)
	if err != nil {
		t.Fatalf("IntMap.SetValue returned error: %v", err)
	}
	output := refVal.Interface().(map[string]extendedIntMap)
	if output["foo"] != 789 {
		t.Errorf("expected 789, got %v", output["foo"])
	}
}
