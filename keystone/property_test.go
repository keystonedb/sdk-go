package keystone

import "testing"

func Test_PropertyNils(t *testing.T) {
	var prop *Property
	if prop.Name() != "" {
		t.Errorf("Property.Name() = %s; want empty string", prop.Name())
	}

	// Should not panic
	prop.SetPrefix("")
}

func Test_Property_Name(t *testing.T) {
	tests := []struct {
		name     string
		property string
		expect   string
	}{
		{"snake_case", "snake_case", "snake_case"},
		{"camelCase", "camelCase", "camel_case"},
		{"PascalCase", "PascalCase", "pascal_case"},
		{"kebab-case", "kebab-case", "kebab_case"},
		{"UPPERCASE", "UPPERCASE", "uppercase"},
		{"lowercase", "lowercase", "lowercase"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			prop := NewProperty(test.property)
			if prop.Name() != test.expect {
				t.Errorf("Property.Name() = %s; want %s", prop.Name(), test.expect)
			}
		})
	}
}

func Test_Property_SetPrefix(t *testing.T) {
	tests := []struct {
		name     string
		property string
		prefix   string
		expect   string
	}{
		{"snake_case", "snake_case", "prefix", "prefix.snake_case"},
		{"camelCase", "camelCase", "prefix", "prefix.camel_case"},
		{"PascalCase", "PascalCase", "prefix", "prefix.pascal_case"},
		{"kebab-case", "kebab-case", "prefix", "prefix.kebab_case"},
		{"UPPERCASE", "UPPERCASE", "prefix", "prefix.uppercase"},
		{"lowercase", "lowercase", "prefix", "prefix.lowercase"},
		{"secondary prefix", "lowercase", "prefix.secondary", "prefix.secondary.lowercase"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			prop := NewProperty(test.property)
			prop.SetPrefix(test.prefix)
			if prop.Name() != test.expect {
				t.Errorf("Property.Name() = %s; want %s", prop.Name(), test.expect)
			}
		})
	}
}

func Test_Type(t *testing.T) {
	tests := []struct {
		name   string
		with   interface{}
		expect string
	}{
		{"MarshaledEntity", MarshaledEntity{}, "marshaled-entity"},
		{"MarshaledEntity", &MarshaledEntity{}, "marshaled-entity"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if Type(test.with) != test.expect {
				t.Errorf("Type() = %s; want %s", Type(test.with), test.expect)
			}
		})
	}
}
