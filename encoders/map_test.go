package encoders

import (
	"errors"
	"github.com/keystonedb/sdk-go/sdk-go/proto"
	"reflect"
	"testing"
)

func Test_MapEncoders(t *testing.T) {

	tests := []struct {
		name      string
		input     interface{}
		compare   map[string][]byte
		useFunc   func(reflect.Value) (*proto.Value, error)
		expectErr error
	}{
		{"bytes map", map[string][]byte{"a": []byte("b")}, map[string][]byte{"a": []byte("b")}, Map, InvalidMapError},
		{"string map", map[string]string{"a": "b"}, map[string][]byte{"a": []byte("b")}, StringMap, InvalidStringMapError},
		{"int map", map[string]int{"a": 42}, map[string][]byte{"a": []byte("42")}, IntMap, InvalidIntMapError},
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
				if matchErr := proto.MatchRepeatedValue(str.GetArray(), &proto.RepeatedValue{KeyValue: test.compare}); matchErr != nil {
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
