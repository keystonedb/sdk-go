package keystone

import "testing"

func TestDynamicPropertiesFromStruct(t *testing.T) {
	type TestingStruct struct {
		PropertyName string `keystone:"property_name"`
		IntVal       int
		IsConfigured bool
	}

	props, err := DynamicPropertiesFromStruct(&TestingStruct{
		PropertyName: "random-prop",
		IntVal:       135,
		IsConfigured: true,
	})
	if err != nil {
		t.Fatalf("Failed to create dynamic properties: %v", err)
	}
	if len(props) != 3 {
		t.Fatalf("Expected 3 properties, got %d", len(props))
	}
	if props[0].GetProperty() != "property_name" {
		t.Fatalf("Expected property name 'property_name', got '%s'", props[0].GetProperty())
	}
	if props[1].GetProperty() != "int_val" {
		t.Fatalf("Expected property name 'int_val', got '%s'", props[1].GetProperty())
	}
	if props[2].GetProperty() != "is_configured" {
		t.Fatalf("Expected property name 'is_configured', got '%s'", props[2].GetProperty())
	}
	if props[0].GetValue().GetText() != "random-prop" {
		t.Fatalf("Expected property value 'random-prop', got '%s'", props[0].GetValue().GetText())
	}
	if props[1].GetValue().GetInt() != 135 {
		t.Fatalf("Expected property value 135, got %d", props[1].GetValue().GetInt())
	}
	if props[2].GetValue().GetBool() != true {
		t.Fatalf("Expected property value true, got %v", props[2].GetValue().GetBool())
	}
}
