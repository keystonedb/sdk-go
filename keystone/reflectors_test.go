package keystone

import (
	"reflect"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGetReflector(t *testing.T) {
	tnow := time.Now()
	tests := []struct {
		name     string
		value    interface{}
		wantType string
	}{
		{"string", "test", "String"},
		{"bool", true, "Bool"},
		{"int", int(42), "Int"},
		{"int8", int8(42), "Int"},
		{"int16", int16(42), "Int"},
		{"int32", int32(42), "Int"},
		{"int64", int64(42), "Int"},
		{"uint", uint(42), "Int"},
		{"uint8", uint8(42), "Int"},
		{"uint16", uint16(42), "Int"},
		{"uint32", uint32(42), "Int"},
		{"float64", float64(42.0), "Float"},
		{"float32", float32(42.0), "Float"},
		{"string_map", map[string]string{}, "StringMap"},
		{"int_map", map[string]int{}, "IntMap"},
		{"bytes_slice", []byte{}, "Bytes"},
		{"string_slice", []string{}, "StringSlice"},
		{"int_slice", []int{}, "IntSlice"},
		{"time", time.Now(), "Time"},
		{"time", &tnow, "Time"},
		{"timestamp", &timestamppb.Timestamp{}, "Timestamp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := reflect.ValueOf(tt.value)
			ref := GetReflector(val.Type(), val)
			if ref == nil {
				t.Errorf("GetReflector() returned nil for %v", tt.name)
				return
			}
			if gotType := reflect.TypeOf(ref).Name(); gotType != tt.wantType {
				t.Errorf("GetReflector() got = %v, want %v", gotType, tt.wantType)
			}
		})
	}
}
