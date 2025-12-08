package reflector

import (
	"reflect"
	"testing"
)

func TestStringMap(t *testing.T) {
	input := map[string]string{
		"foo": "bar",
	}

	p := &StringMap{}
	ref := reflect.ValueOf(input)
	pVal, err := p.ToProto(ref)
	if err != nil {
		t.Error(err)
	}

	if pVal.Array == nil {
		t.Error("Expected non-nil array")
	}

	for k, v := range input {
		if v != string(pVal.Array.KeyValue[k]) {
			t.Error("")
		}
	}

	refVal := reflect.ValueOf(new(map[string]string)).Elem()
	refErr := p.SetValue(pVal, refVal)
	if refErr != nil {
		t.Errorf("StringMap.SetValue returned error: %v", refErr)
	}
	if output, ok := refVal.Interface().(map[string]string); !ok {
		t.Errorf("StringMap.SetValue didn't return a map[string]string")
	} else {
		for k, v := range input {
			if v != output[k] {
				t.Error("")
			}
		}
	}
}

type extendedString string

func TestStringMap_Ext(t *testing.T) {
	input := map[string]extendedString{
		"foo": "bar",
	}

	p := &StringMap{}
	pVal, err := p.ToProto(reflect.ValueOf(input))
	if err != nil {
		t.Error(err)
	}

	if pVal.Array == nil {
		t.Error("Expected non-nil array")
	}

	for k, v := range input {
		if string(v) != string(pVal.Array.KeyValue[k]) {
			t.Error("")
		}
	}

	refVal := reflect.ValueOf(new(map[string]extendedString)).Elem()
	refErr := p.SetValue(pVal, refVal)
	if refErr != nil {
		t.Errorf("StringMap.SetValue returned error: %v", refErr)
	}
	if output, ok := refVal.Interface().(map[string]extendedString); !ok {
		t.Errorf("StringMap.SetValue didn't return a map[string]extendedString")
	} else {
		for k, v := range input {
			if v != output[k] {
				t.Error("")
			}
		}
	}
}
