package reflector

import (
	"errors"
	"reflect"
	"testing"
)

func Test_Int(t *testing.T) {
	tests := []struct {
		reflector Int
		input     any
		expect    int64
	}{
		{Int{Kind: reflect.Int}, int(42), int64(42)},
		{Int{Kind: reflect.Int8}, int8(42), int64(42)},
		{Int{Kind: reflect.Int16}, int16(42), int64(42)},
		{Int{Kind: reflect.Int32}, int32(42), int64(42)},
		{Int{Kind: reflect.Int64}, int64(42), int64(42)},
	}

	for _, test := range tests {
		t.Run(test.reflector.Kind.String(), func(t *testing.T) {
			for _, tstVal := range []reflect.Value{
				reflect.ValueOf(test.input), reflect.ValueOf(&test.input),
			} {
				str, err := test.reflector.ToProto(tstVal)
				if err != nil {
					t.Errorf("Int.ToProto returned error: %v", err)
				}
				if str.GetInt() != test.expect {
					t.Errorf("Int.ToProto returned %v, want %v", str.GetText(), test.input)
				}

				var refVal reflect.Value
				switch test.reflector.Kind {
				case reflect.Int:
					refVal = reflect.ValueOf(new(int)).Elem()
				case reflect.Int8:
					refVal = reflect.ValueOf(new(int8)).Elem()
				case reflect.Int16:
					refVal = reflect.ValueOf(new(int16)).Elem()
				case reflect.Int32:
					refVal = reflect.ValueOf(new(int32)).Elem()
				case reflect.Int64:
					refVal = reflect.ValueOf(new(int64)).Elem()
				}
				refErr := test.reflector.SetValue(str, refVal)
				if refErr != nil {
					t.Errorf("Int.SetValue returned error: %v", refErr)
				}
				if refVal.Int() != test.expect {
					t.Errorf("Int.SetValue returned %v, want %v", refVal.Int(), test.expect)
				}
			}
		})
	}
}

func Test_Int_Unsupported(t *testing.T) {
	tests := []struct {
		name    string
		attempt any
	}{
		{"string", ""},
		{"pointer", new(string)},
		{"bool false", false},
		{"bool true", true},
		{"float32", float32(42.12)},
		{"float64", float64(42.12)},
	}

	refs := []Int{
		{Kind: reflect.Int},
		{Kind: reflect.Int8},
		{Kind: reflect.Int16},
		{Kind: reflect.Int32},
		{Kind: reflect.Int64},
	}

	ref := Int{}
	if err := ref.SetValue(nil, reflect.ValueOf(0)); !errors.Is(err, UnsupportedTypeError) {
		t.Errorf("unconfigured Int.SetValue returned error: %v, want %v", err, UnsupportedTypeError)
	}

	for _, test := range tests {
		for _, ref := range refs {
			t.Run(test.name+" - "+ref.Kind.String(), func(t *testing.T) {
				val, err := ref.ToProto(reflect.ValueOf(test.attempt))
				if val != nil {
					t.Errorf("Int.ToProto returned %v, want nil", val)
				}
				if !errors.Is(err, UnsupportedTypeError) {
					t.Errorf("Int.ToProto returned error: %v, want %v", err, UnsupportedTypeError)
				}
			})
		}
	}
}
