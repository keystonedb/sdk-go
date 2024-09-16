package keystone

import "testing"

func TestWatcher_Changes(t *testing.T) {
	toTest := struct {
		entity
		Name string
		Age  int
	}{
		entity: entity{},
		Name:   "John",
		Age:    30,
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
		entity
		Name string
		Age  int
	}{
		entity: entity{},
		Name:   "John",
		Age:    30,
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
