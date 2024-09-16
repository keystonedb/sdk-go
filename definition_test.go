package keystone

import "testing"

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
