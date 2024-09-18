package reflector

import (
	"fmt"
	"reflect"
	"testing"
)

func TestFloat(t *testing.T) {
	ref64 := float64(12.34)
	ref32 := float32(34.12)
	var nilFloat64 float64
	var nilFloat32 float32

	tests := []struct {
		name    string
		against Float
		input   any
		expect  any
	}{
		{"float64", Float{}, ref64, float64(12.34)},
		{"float32", Float{Is32: true}, ref32, float32(34.12)},
		{"nil float64", Float{}, nilFloat64, float64(0)},
		{"nil float32", Float{Is32: true}, nilFloat32, float32(0)},
		{"int", Float{}, 123, float64(123)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ref := test.against
			val, err := ref.ToProto(reflect.ValueOf(test.input))
			if err != nil {
				t.Errorf("Float.ToProto returned error: %v", err)
			}

			if fmt.Sprintf("%f", val.GetFloat()) != fmt.Sprintf("%f", test.expect) {
				t.Errorf("Float.ToProto returned %v, want %v", val.GetFloat(), test.expect)
			}

			var refVal reflect.Value
			if test.against.Is32 {
				refVal = reflect.ValueOf(new(float32)).Elem()
			} else {
				refVal = reflect.ValueOf(new(float64)).Elem()
			}
			refErr := ref.SetValue(val, refVal)
			if refErr != nil {
				t.Errorf("Float.SetValue returned error: %v", refErr)
			}

			if fmt.Sprintf("%f", refVal.Float()) != fmt.Sprintf("%f", test.expect) {
				t.Errorf("Float.SetValue returned %v, want %v", refVal.Float(), test.expect)
			}
		})
	}
}
