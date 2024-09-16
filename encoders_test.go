package keystone

import (
	"errors"
	"github.com/keystonedb/sdk-go/sdk-go/encoders"
	"github.com/keystonedb/sdk-go/sdk-go/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"reflect"
	"testing"
	"time"
)

type basicStruct struct {
	Name string
}

type testValueMarshaler struct {
	stringValue string
	error       error
}

func (t testValueMarshaler) MarshalValue() (*proto.Value, error) {
	return &proto.Value{Text: t.stringValue}, t.error
}

func Test_NewTypeEncoder(t *testing.T) {
	tNow := time.Now()
	tsNow := timestamppb.Now()
	tests := []struct {
		name  string
		input interface{}
		want  func(reflect.Value) (*proto.Value, error)
	}{
		{"struct", testValueMarshaler{stringValue: "abc"}, addrValueUnmarshalFunc},
		{"struct ref", &testValueMarshaler{stringValue: "abc"}, valueUnmarshalFunc},
		{"string", "", encoders.String},
		{"bool", true, encoders.Bool},
		{"int", int(42), encoders.Int},
		{"int8", int8(42), encoders.Int},
		{"int16", int16(42), encoders.Int},
		{"int32", int32(42), encoders.Int},
		{"int64", int64(42), encoders.Int},
		{"float32", float32(42.42), encoders.Float32},
		{"float64", float64(42.42), encoders.Float},
		{"string slice", []string{}, encoders.StringSlice},
		{"int slice", []int{}, encoders.IntSlice},
		{"int64 slice", []int64{}, encoders.Int64Slice},
		{"int32 slice", []int32{}, encoders.Int32Slice},
		{"int16 slice", []int16{}, encoders.Int16Slice},
		{"int8 slice", []int8{}, encoders.Int8Slice},
		{"time", tNow, encoders.Time},
		{"timestamp", *tsNow, encoders.Timestamp},
		{"timestamp pointer", tsNow, encoders.Timestamp},
		{"time pointer", &tNow, encoders.Time},
		{"string map", map[string]string{}, encoders.StringMap},
		{"int map", map[string]int{}, encoders.IntMap},
		{"map", map[string][]byte{}, encoders.Map},
		{"chan", make(chan string), nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			enc := newTypeEncoder(reflect.TypeOf(test.input))
			if test.want == nil && enc == nil {
				return
			}
			expectType := reflect.ValueOf(test.want)
			if enc == nil {
				t.Errorf("newTypeEncoder returned nil, want %v for %v", expectType, test.input)
			}
			if reflect.ValueOf(enc).Pointer() != expectType.Pointer() {
				t.Errorf("newTypeEncoder returned %v, want %v for %v", enc, expectType, test.input)
			}
		})
	}
}

func Test_ValueUnmarshalFunc(t *testing.T) {
	input := &testValueMarshaler{stringValue: "abc"}
	val, err := valueUnmarshalFunc(reflect.ValueOf(input))
	if err != nil {
		t.Errorf("valueUnmarshalFunc returned error: %v", err)
	}
	if val.GetText() != "abc" {
		t.Errorf("valueUnmarshalFunc returned %v, want abc", val.GetText())
	}
}

func Test_ValueUnmarshalFuncDeRef(t *testing.T) {
	input := testValueMarshaler{stringValue: "abc"}
	val, err := valueUnmarshalFunc(reflect.ValueOf(input))
	if err != nil {
		t.Errorf("valueUnmarshalFunc returned error: %v", err)
	}
	if val.GetText() != "abc" {
		t.Errorf("valueUnmarshalFunc returned %v, want abc", val.GetText())
	}
}

func Test_ValueUnmarshalFuncNil(t *testing.T) {
	val, err := valueUnmarshalFunc(reflect.ValueOf(nil))
	if err != nil {
		t.Errorf("valueUnmarshalFunc returned error: %v", err)
	}
	if val != nil {
		t.Errorf("valueUnmarshalFunc returned %v, want nil", val.GetText())
	}
}

func Test_ValueUnmarshalFuncStruct(t *testing.T) {
	val, err := valueUnmarshalFunc(reflect.ValueOf(basicStruct{}))
	if err == nil || !errors.Is(err, InvalidValueMarshalerError) {
		t.Errorf("expected not a ValueMarshaler, got error: %v", err)
	}
	if val != nil {
		t.Errorf("valueUnmarshalFunc returned %v, want nil", val.GetText())
	}
}

func Test_AddrValueUnmarshalFunc(t *testing.T) {
	input := testValueMarshaler{stringValue: "abc"}
	val, err := addrValueUnmarshalFunc(reflect.ValueOf(input))
	if err != nil {
		t.Errorf("addrValueUnmarshalFunc returned error: %v", err)
	}
	if val.GetText() != "abc" {
		t.Errorf("addrValueUnmarshalFunc returned %v, want abc", val.GetText())
	}
}
