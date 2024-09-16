package encoders

import (
	"bytes"
	"reflect"
	"testing"
)

func Test_BytesEncoder(t *testing.T) {
	input := []byte("Hello, World!")
	for _, test := range []reflect.Value{
		reflect.ValueOf(input), reflect.ValueOf(&input),
	} {
		str, err := Bytes(test)
		if err != nil {
			t.Errorf("bytesEncoder returned error: %v", err)
		}
		if bytes.Compare(str.GetRaw(), input) != 0 {
			t.Errorf("bytesEncoder returned %v, want %v", str.GetRaw(), input)
		}
	}
}

func Test_String(t *testing.T) {
	input := "Hello, World!"
	for _, test := range []reflect.Value{
		reflect.ValueOf(input), reflect.ValueOf(&input),
	} {
		str, err := String(test)
		if err != nil {
			t.Errorf("stringEncoder returned error: %v", err)
		}
		if str.GetText() != input {
			t.Errorf("stringEncoder returned %v, want %v", str.GetText(), input)
		}
	}
}

func Test_Bool(t *testing.T) {
	input := true
	for _, test := range []reflect.Value{
		reflect.ValueOf(input), reflect.ValueOf(&input),
	} {
		bl, err := Bool(test)
		if err != nil {
			t.Errorf("boolEncoder returned error: %v", err)
		}
		if bl.GetBool() != input {
			t.Errorf("boolEncoder returned %v, want %v", bl.GetBool(), input)
		}
	}
}

func Test_Int(t *testing.T) {
	intVal := int(42)
	int8Val := int8(42)
	int16Val := int16(4233)
	int32Val := int32(422332)
	int64Val := int64(4223973223)
	tests := []struct {
		input reflect.Value
		want  int64
	}{
		{reflect.ValueOf(intVal), 42},
		{reflect.ValueOf(&intVal), 42},
		{reflect.ValueOf(int8Val), 42},
		{reflect.ValueOf(&int8Val), 42},
		{reflect.ValueOf(int16Val), 4233},
		{reflect.ValueOf(&int16Val), 4233},
		{reflect.ValueOf(int32Val), 422332},
		{reflect.ValueOf(&int32Val), 422332},
		{reflect.ValueOf(int64Val), 4223973223},
		{reflect.ValueOf(&int64Val), 4223973223},
	}
	for _, test := range tests {
		in, err := Int(test.input)
		if err != nil {
			t.Errorf("intEncoder returned error: %v", err)
		}
		if in.GetInt() != test.want {
			t.Errorf("intEncoder returned %v, want %v", in.GetInt(), test.want)
		}
	}
}

func Test_Float(t *testing.T) {
	floatVal := float64(42.42)
	tests := []struct {
		input reflect.Value
		want  float64
	}{
		{reflect.ValueOf(floatVal), 42.42},
		{reflect.ValueOf(&floatVal), 42.42},
	}
	for _, test := range tests {
		in, err := Float(test.input)
		if err != nil {
			t.Errorf("floatEncoder returned error: %v", err)
		}
		if in.GetFloat() != test.want {
			t.Errorf("floatEncoder returned %v, want %v", in.GetFloat(), test.want)
		}
	}
}

func Test_Float32(t *testing.T) {
	input := float32(42.42)
	for _, test := range []reflect.Value{
		reflect.ValueOf(input), reflect.ValueOf(&input),
	} {
		fl, err := Float32(test)
		if err != nil {
			t.Errorf("float32Encoder returned error: %v", err)
		}
		if float32(fl.GetFloat()) != input {
			t.Errorf("float32Encoder returned %v, want %v", fl.GetFloat(), float64(input))
		}
	}

	for _, test := range []reflect.Value{
		reflect.ValueOf(float64(42.42)), reflect.ValueOf("string"), reflect.ValueOf(true),
	} {
		_, err := Float32(test)
		if err == nil {
			t.Errorf("float32Encoder returned no error, want error")
		}
	}
}
