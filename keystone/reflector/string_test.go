package reflector

import (
	"reflect"
	"testing"
)

func TestString(t *testing.T) {
	hello := "Hello World!"
	tests := []struct {
		name   string
		input  any
		expect string
	}{
		{"value", hello, hello},
		{"pointer", &hello, hello},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ref := String{}
			val, err := ref.ToProto(reflect.ValueOf(test.input))
			if err != nil {
				t.Errorf("String.ToProto returned error: %v", err)
			}
			if val.GetText() != test.expect {
				t.Errorf("String.ToProto returned %v, want %v", val.GetText(), test.expect)
			}

			refVal := reflect.ValueOf(new(string)).Elem()
			refErr := ref.SetValue(val, refVal)
			if refErr != nil {
				t.Errorf("String.SetValue returned error: %v", refErr)
			}
			if refVal.String() != test.expect {
				t.Errorf("String.SetValue returned %v, want %v", refVal.String(), test.expect)
			}
		})
	}
}
