package reflector

import (
	"bytes"
	"reflect"
	"testing"
)

func TestBytes(t *testing.T) {
	bRef := []byte("test")
	var nilByte []byte

	tests := []struct {
		name   string
		input  any
		expect []byte
	}{
		{"bytes", bRef, []byte("test")},
		{"bytes pointer", &bRef, []byte("test")},
		{"nil pointer", nilByte, nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ref := Bytes{}
			val, err := ref.ToProto(reflect.ValueOf(test.input))
			if err != nil {
				t.Errorf("Bytes.ToProto returned error: %v", err)
			}
			if !bytes.Equal(val.GetRaw(), test.expect) {
				t.Errorf("Bytes.ToProto returned %v, want %v", val.GetRaw(), test.expect)
			}

			refVal := reflect.ValueOf(new([]byte)).Elem()
			refErr := ref.SetValue(val, refVal)
			if refErr != nil {
				t.Errorf("Bytes.SetValue returned error: %v", refErr)
			}
			if !bytes.Equal(refVal.Bytes(), test.expect) {
				t.Errorf("Bytes.SetValue returned %v, want %v", refVal.Bytes(), test.expect)
			}
		})
	}
}
