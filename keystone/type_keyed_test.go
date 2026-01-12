package keystone

import (
	"reflect"
	"testing"
)

func TestKeyed(t *testing.T) {
	k := NewKeyed[string](nil)
	k.Set("a", "1")
	k.Set("b", "2")

	if k.Values()["a"] != "1" {
		t.Errorf("Expected 1, got %v", k.Values()["a"])
	}

	k.Append("c", "3")
	if k.Values()["c"] != "3" {
		t.Errorf("Expected 3, got %v", k.Values()["c"])
	}

	k.Remove("a")
	if _, ok := k.Values()["a"]; ok {
		t.Errorf("Expected a to be removed")
	}

	mv, err := k.MarshalValue()
	if err != nil {
		t.Fatal(err)
	}

	k2 := NewKeyed[string](nil)
	err = k2.UnmarshalValue(mv)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(k.Values(), k2.Values()) {
		t.Errorf("Expected %v, got %v", k.Values(), k2.Values())
	}
}

func TestKeyedStruct(t *testing.T) {
	type testStruct struct {
		Name string
		Age  int
	}

	k := NewKeyed[testStruct](nil)
	k.Set("user1", testStruct{Name: "Alice", Age: 30})

	mv, err := k.MarshalValue()
	if err != nil {
		t.Fatal(err)
	}

	k2 := NewKeyed[testStruct](nil)
	err = k2.UnmarshalValue(mv)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(k.Values(), k2.Values()) {
		t.Errorf("Expected %v, got %v", k.Values(), k2.Values())
	}
}
