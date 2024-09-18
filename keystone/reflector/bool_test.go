package reflector

import (
	"reflect"
	"testing"
)

func TestBool(t *testing.T) {
	tr := true
	fr := false
	tests := []struct {
		name   string
		input  any
		expect bool
	}{
		{"true", true, true},
		{"false", false, false},
		{"true pointer", &tr, true},
		{"false pointer", &fr, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ref := Bool{}
			val, err := ref.ToProto(reflect.ValueOf(test.input))
			if err != nil {
				t.Errorf("bool.ToProto returned error: %v", err)
			}
			if val.GetBool() != test.expect {
				t.Errorf("bool.ToProto returned %v, want %v", val.GetBool(), test.expect)
			}

			refVal := reflect.ValueOf(new(bool)).Elem()
			refErr := ref.SetValue(val, refVal)
			if refErr != nil {
				t.Errorf("bool.SetValue returned error: %v", refErr)
			}
			if refVal.Bool() != test.expect {
				t.Errorf("bool.SetValue returned %v, want %v", refVal.Bool(), test.expect)
			}
		})
	}
}
