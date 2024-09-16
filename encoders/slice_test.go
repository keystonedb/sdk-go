package encoders

import (
	"errors"
	"github.com/keystonedb/sdk-go/sdk-go/proto"
	"reflect"
	"testing"
)

func Test_StringSlice(t *testing.T) {
	input := []string{"Hello", "World"}
	for _, test := range []reflect.Value{
		reflect.ValueOf(input), reflect.ValueOf(&input),
	} {
		str, err := StringSlice(test)
		if err != nil {
			t.Errorf("stringSliceEncoder returned error: %v", err)
		}
		if matchErr := proto.MatchRepeatedValue(str.GetArray(), &proto.RepeatedValue{Strings: input}); matchErr != nil {
			t.Error(matchErr)
		}
	}

	result, err := StringSlice(reflect.ValueOf([]int{42}))
	if result != nil {
		t.Errorf("stringSliceEncoder returned %v, want nil", result)
	}
	if !errors.Is(err, InvalidStringSliceError) {
		t.Errorf("stringSliceEncoder returned no error, want InvalidStringSliceError")
	}
}

func Test_IntSliceEncoders(t *testing.T) {

	tests := []struct {
		name      string
		input     interface{}
		compare   []int64
		useFunc   func(reflect.Value) (*proto.Value, error)
		expectErr error
	}{
		{"int64", []int64{42, 42}, []int64{42, 42}, Int64Slice, InvalidInt64SliceError},
		{"int", []int{42, 42}, []int64{42, 42}, IntSlice, InvalidIntSliceError},
		{"int32", []int32{42, 42}, []int64{42, 42}, Int32Slice, InvalidInt32SliceError},
		{"int16", []int16{42, 42}, []int64{42, 42}, Int16Slice, InvalidInt16SliceError},
		{"int8", []int8{42, 42}, []int64{42, 42}, Int8Slice, InvalidInt8SliceError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, testVals := range []reflect.Value{
				reflect.ValueOf(test.input), reflect.ValueOf(&test.input),
			} {
				str, err := test.useFunc(testVals)
				if err != nil {
					t.Errorf("%s returned error: %v", test.name, err)
				}
				if matchErr := proto.MatchRepeatedValue(str.GetArray(), &proto.RepeatedValue{Ints: test.compare}); matchErr != nil {
					t.Error(matchErr)
				}
			}

			result, err := test.useFunc(reflect.ValueOf("string"))
			if result != nil {
				t.Errorf(" returned %v, want nil", result)
			}
			if !errors.Is(err, test.expectErr) {
				t.Errorf("got err %v, want %v", err, test.expectErr)
			}
		})
	}
}
