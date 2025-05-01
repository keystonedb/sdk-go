package keystone

import (
	"errors"
	"testing"
)

func TestWatcher_Changes(t *testing.T) {
	toTest := struct {
		Name string
		Age  int
	}{
		Name: "John",
		Age:  30,
	}

	nameProp := NewProperty("Name")
	ageProp := NewProperty("Age")

	watched, err := NewWatcher(toTest)
	if err != nil {
		t.Errorf("NewWatcher() returned error: %v", err)
	}

	toTest.Name = "Smith"
	toTest.Age = 21

	changes, err := watched.Changes(toTest, true)
	if err != nil {
		t.Errorf("Changes() returned error: %v", err)
	}

	if len(changes) != 2 {
		t.Fatalf("Changes() returned %d changes, want 2", len(changes))
	}
	if changes[nameProp].Text != "Smith" {
		t.Errorf("Changes() returned %s, want Smith", changes[nameProp].Text)
	}
	if changes[ageProp].Int != 21 {
		t.Errorf("Changes() returned %d, want 21", changes[ageProp].Int)
	}
}

func TestWatcher_ChangesFromNil(t *testing.T) {
	w, err := NewWatcher(struct{}{})
	if err != nil {
		t.Errorf("NewWatcher() returned error: %v", err)
	}

	nameProp := NewProperty("Name")
	ageProp := NewProperty("Age")

	toTest := struct {
		Name string
		Age  int
	}{
		Name: "John",
		Age:  30,
	}
	changes, err := w.Changes(toTest, true)
	if len(changes) != 2 {
		t.Fatalf("Changes() returned %d changes, want 2", len(changes))
	}
	if changes[nameProp].Text != "John" {
		t.Errorf("Changes() returned %s, want John", changes[nameProp].Text)
	}
	if changes[ageProp].Int != 30 {
		t.Errorf("Changes() returned %d, want 30", changes[ageProp].Int)
	}
}

func TestWatcher_DefaultWatcher(t *testing.T) {
	toTest := struct {
		Name string
		Age  int
	}{
		Age: 30,
	}

	w, err := NewDefaultsWatcher(toTest)
	if err != nil {
		t.Errorf("NewDefaultsWatcher() returned error: %v", err)
	}

	ageProp := NewProperty("Age")

	// Name is unchanged from defaults, but age should alter the default values

	changes, err := w.Changes(toTest, true)
	if len(changes) != 1 {
		t.Fatalf("Changes() returned %d changes, want 2", len(changes))
	}
	if changes[ageProp].Int != 30 {
		t.Errorf("Changes() returned %d, want 30", changes[ageProp].Int)
	}
}

func TestWatcher_CatchErrors(t *testing.T) {
	xErr := errors.New("test error")
	safe := struct{ Apply testValueMarshaler }{testValueMarshaler{stringValue: "test"}}
	failing := struct{ Apply testValueMarshaler }{testValueMarshaler{error: xErr}}
	_, err := NewWatcher(failing)
	if !errors.Is(err, xErr) {
		t.Errorf("NewWatcher() returned %v, want %v", err, xErr)
	}

	w, err := NewWatcher(safe)
	if err != nil {
		t.Errorf("NewWatcher() returned error: %v", err)
	}

	_, err = w.Changes(failing, true)
	if !errors.Is(err, xErr) {
		t.Errorf("Changes() returned %v, want %v", err, xErr)
	}
}

func TestWatcher_DataTypeChanges(t *testing.T) {
	toTest := struct {
		Name Mixed
		Keys KeyMixed
	}{
		Name: NewMixed("John"),
		Keys: NewKeyMixed(nil),
	}

	toTest.Keys.Set("key1", NewMixed("value1"))

	nameProp := NewProperty("Name")
	keysProp := NewProperty("Keys")

	watched, err := NewWatcher(toTest)
	if err != nil {
		t.Errorf("NewWatcher() returned error: %v", err)
	}

	toTest.Name.SetString("Smith")
	toTest.Keys.Append("key2", NewMixed("value1"))

	changes, err := watched.Changes(toTest, false)
	if err != nil {
		t.Errorf("Changes() returned error: %v", err)
	}

	if len(changes) != 2 {
		t.Fatalf("Changes() returned %d changes, want 2", len(changes))
	}
	if changes[nameProp].Text != "Smith" {
		t.Errorf("Changes() returned %s, want Smith", changes[nameProp].Text)
	}
	if changes[keysProp].ArrayAppend.GetMixed()["key2"].GetText() != "value1" {
		t.Errorf("Changes() returned %s, want value1", changes[keysProp].ArrayAppend.GetMixed()["key2"].GetText())
	}
}
