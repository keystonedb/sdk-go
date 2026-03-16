package keystone

import (
	"testing"

	"github.com/keystonedb/sdk-go/proto"
)

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
	propMap := make(map[string]*proto.Value, len(props))
	for _, p := range props {
		propMap[p.GetProperty()] = p.GetValue()
	}
	if _, ok := propMap["property_name"]; !ok {
		t.Fatal("Expected property 'property_name' to exist")
	}
	if _, ok := propMap["int_val"]; !ok {
		t.Fatal("Expected property 'int_val' to exist")
	}
	if _, ok := propMap["is_configured"]; !ok {
		t.Fatal("Expected property 'is_configured' to exist")
	}
	if propMap["property_name"].GetText() != "random-prop" {
		t.Fatalf("Expected property value 'random-prop', got '%s'", propMap["property_name"].GetText())
	}
	if propMap["int_val"].GetInt() != 135 {
		t.Fatalf("Expected property value 135, got %d", propMap["int_val"].GetInt())
	}
	if propMap["is_configured"].GetBool() != true {
		t.Fatalf("Expected property value true, got %v", propMap["is_configured"].GetBool())
	}
}

// TestDynamicPropertiesFromStructWithoutDefaults_SetNull tests that a forced property
// set to its zero/null value is included in the output. This simulates clearing a
// previously-set field via RemoteMutate with MutateProperties.
func TestDynamicPropertiesFromStructWithoutDefaults_SetNull(t *testing.T) {
	type Config struct {
		DynamicRemoteEntity
		Name        string
		DefaultTerm *Interval
	}

	t.Run("forced null interval is removed", func(t *testing.T) {
		config := &Config{}
		// DefaultTerm is nil - simulates clearing a previously-set interval
		forceProperties := map[string]bool{
			"default_term": true,
		}

		_, removeProps, err := DynamicPropertiesFromStructWithoutDefaults(config, forceProperties)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		found := false
		for _, p := range removeProps {
			if p == "default_term" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected 'default_term' to be in removal list when forced at zero value")
		}
	})

	t.Run("forced empty string is removed", func(t *testing.T) {
		config := &Config{}
		// Name is "" (default) - simulates clearing a name field
		forceProperties := map[string]bool{
			"name": true,
		}

		_, removeProps, err := DynamicPropertiesFromStructWithoutDefaults(config, forceProperties)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		found := false
		for _, p := range removeProps {
			if p == "name" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected 'name' to be in removal list when forced at empty value")
		}
	})
}
