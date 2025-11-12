package reflector

import (
	"reflect"
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	currentTime := time.Now()
	tests := []struct {
		name   string
		input  any
		expect time.Time
	}{
		{"value", currentTime, currentTime},
		{"pointer", &currentTime, currentTime},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ref := Time{}
			val, err := ref.ToProto(reflect.ValueOf(test.input))
			if err != nil {
				t.Errorf("Time.ToProto returned error: %v", err)
			}
			if !val.GetTime().AsTime().Equal(currentTime) {
				t.Errorf("Time.ToProto returned %v, want %v", val.GetTime(), currentTime)
			}

			refVal := reflect.ValueOf(new(time.Time)).Elem()
			refErr := ref.SetValue(val, refVal)
			if refErr != nil {
				t.Errorf("Time.SetValue returned error: %v", refErr)
			}

			asTime := refVal.Interface().(time.Time)
			if !test.expect.Equal(asTime) {
				t.Errorf("Time.SetValue returned %v, want %v", asTime, test.expect)
			}
		})
	}
}
