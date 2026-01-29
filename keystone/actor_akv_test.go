package keystone

import (
	"context"
	"reflect"
	"testing"

	"github.com/keystonedb/sdk-go/proto"
)

// Tests for AKVProperty helper functions

func TestAKVProperty_toProto(t *testing.T) {
	prop := &AKVProperty{
		Property: &proto.Property{
			Name:     "test-property",
			DataType: proto.Property_Text,
		},
		Value: &proto.Value{
			Text: "test-value",
		},
	}

	result := prop.toProto()

	if result.Property.Name != "test-property" {
		t.Errorf("toProto().Property.Name = %s; want test-property", result.Property.Name)
	}
	if result.Property.DataType != proto.Property_Text {
		t.Errorf("toProto().Property.DataType = %v; want %v", result.Property.DataType, proto.Property_Text)
	}
	if result.Value.Text != "test-value" {
		t.Errorf("toProto().Value.Text = %s; want test-value", result.Value.Text)
	}
}

func TestAKVRaw(t *testing.T) {
	value := &proto.Value{
		Text: "raw-value",
	}

	prop := AKVRaw("raw-property", value)

	if prop.Property.Name != "raw-property" {
		t.Errorf("AKVRaw().Property.Name = %s; want raw-property", prop.Property.Name)
	}
	if prop.Property.DataType != proto.Property_Unmanaged {
		t.Errorf("AKVRaw().Property.DataType = %v; want %v", prop.Property.DataType, proto.Property_Unmanaged)
	}
	if prop.Value != value {
		t.Errorf("AKVRaw().Value = %v; want %v", prop.Value, value)
	}
}

func TestAKV_String(t *testing.T) {
	prop := AKV("string-property", "hello")

	if prop.Property.Name != "string-property" {
		t.Errorf("AKV().Property.Name = %s; want string-property", prop.Property.Name)
	}
	if prop.Property.DataType != proto.Property_Text {
		t.Errorf("AKV().Property.DataType = %v; want %v", prop.Property.DataType, proto.Property_Text)
	}
	if prop.Value == nil {
		t.Error("AKV().Value is nil; want non-nil")
	}
	if prop.Value.Text != "hello" {
		t.Errorf("AKV().Value.Text = %s; want hello", prop.Value.Text)
	}
}

func TestAKV_Int(t *testing.T) {
	prop := AKV("int-property", 42)

	if prop.Property.Name != "int-property" {
		t.Errorf("AKV().Property.Name = %s; want int-property", prop.Property.Name)
	}
	if prop.Property.DataType != proto.Property_Number {
		t.Errorf("AKV().Property.DataType = %v; want %v", prop.Property.DataType, proto.Property_Number)
	}
	if prop.Value == nil {
		t.Error("AKV().Value is nil; want non-nil")
	}
	if prop.Value.Int != 42 {
		t.Errorf("AKV().Value.Int = %d; want 42", prop.Value.Int)
	}
}

func TestAKV_Bool(t *testing.T) {
	prop := AKV("bool-property", true)

	if prop.Property.Name != "bool-property" {
		t.Errorf("AKV().Property.Name = %s; want bool-property", prop.Property.Name)
	}
	if prop.Property.DataType != proto.Property_Boolean {
		t.Errorf("AKV().Property.DataType = %v; want %v", prop.Property.DataType, proto.Property_Boolean)
	}
	if prop.Value == nil {
		t.Error("AKV().Value is nil; want non-nil")
	}
	if prop.Value.Bool != true {
		t.Errorf("AKV().Value.Bool = %v; want true", prop.Value.Bool)
	}
}

func TestAKV_Float(t *testing.T) {
	prop := AKV("float-property", 3.14)

	if prop.Property.Name != "float-property" {
		t.Errorf("AKV().Property.Name = %s; want float-property", prop.Property.Name)
	}
	if prop.Property.DataType != proto.Property_Float {
		t.Errorf("AKV().Property.DataType = %v; want %v", prop.Property.DataType, proto.Property_Float)
	}
	if prop.Value == nil {
		t.Error("AKV().Value is nil; want non-nil")
	}
	if prop.Value.Float != 3.14 {
		t.Errorf("AKV().Value.Float = %f; want 3.14", prop.Value.Float)
	}
}

func TestAKV_NilReflector(t *testing.T) {
	// Test with a type that doesn't have a reflector (channel)
	ch := make(chan int)
	prop := AKV("channel-property", ch)

	if prop.Property.Name != "channel-property" {
		t.Errorf("AKV().Property.Name = %s; want channel-property", prop.Property.Name)
	}
	// Value should be nil when no reflector is found
	if prop.Value != nil {
		t.Errorf("AKV().Value = %v; want nil for unsupported type", prop.Value)
	}
}

func TestAKV_Pointer(t *testing.T) {
	value := "pointer-value"
	prop := AKV("pointer-property", &value)

	if prop.Property.Name != "pointer-property" {
		t.Errorf("AKV().Property.Name = %s; want pointer-property", prop.Property.Name)
	}
	if prop.Property.DataType != proto.Property_Text {
		t.Errorf("AKV().Property.DataType = %v; want %v", prop.Property.DataType, proto.Property_Text)
	}
	if prop.Value == nil || prop.Value.Text != "pointer-value" {
		t.Errorf("AKV().Value.Text = %v; want pointer-value", prop.Value)
	}
}

// Tests for Actor AKV methods

func TestActor_AKVPut_NilActor(t *testing.T) {
	var actor *Actor
	_, err := actor.AKVPut(context.Background(), AKV("test", "value"))
	if err == nil {
		t.Error("expected error for nil actor, got nil")
	}
	if err.Error() != "actor or connection is nil" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestActor_AKVPut_NilConnection(t *testing.T) {
	actor := &Actor{}
	_, err := actor.AKVPut(context.Background(), AKV("test", "value"))
	if err == nil {
		t.Error("expected error for nil connection, got nil")
	}
	if err.Error() != "actor or connection is nil" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestActor_AKVGet_NilActor(t *testing.T) {
	var actor *Actor
	_, err := actor.AKVGet(context.Background(), "test")
	if err == nil {
		t.Error("expected error for nil actor, got nil")
	}
	if err.Error() != "actor or connection is nil" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestActor_AKVGet_NilConnection(t *testing.T) {
	actor := &Actor{}
	_, err := actor.AKVGet(context.Background(), "test")
	if err == nil {
		t.Error("expected error for nil connection, got nil")
	}
	if err.Error() != "actor or connection is nil" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestActor_AKVDel_NilActor(t *testing.T) {
	var actor *Actor
	_, err := actor.AKVDel(context.Background(), "test")
	if err == nil {
		t.Error("expected error for nil actor, got nil")
	}
	if err.Error() != "actor or connection is nil" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestActor_AKVDel_NilConnection(t *testing.T) {
	actor := &Actor{}
	_, err := actor.AKVDel(context.Background(), "test")
	if err == nil {
		t.Error("expected error for nil connection, got nil")
	}
	if err.Error() != "actor or connection is nil" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

// Tests for AKVProperty with different types

func TestAKVProperty_NilProperty(t *testing.T) {
	prop := &AKVProperty{
		Property: nil,
		Value:    &proto.Value{Text: "test"},
	}

	result := prop.toProto()

	if result.Property != nil {
		t.Errorf("toProto().Property = %v; want nil", result.Property)
	}
	if result.Value.Text != "test" {
		t.Errorf("toProto().Value.Text = %s; want test", result.Value.Text)
	}
}

func TestAKVProperty_NilValue(t *testing.T) {
	prop := &AKVProperty{
		Property: &proto.Property{Name: "test"},
		Value:    nil,
	}

	result := prop.toProto()

	if result.Property.Name != "test" {
		t.Errorf("toProto().Property.Name = %s; want test", result.Property.Name)
	}
	if result.Value != nil {
		t.Errorf("toProto().Value = %v; want nil", result.Value)
	}
}

func TestAKV_Bytes(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03}
	prop := AKV("bytes-property", data)

	if prop.Property.Name != "bytes-property" {
		t.Errorf("AKV().Property.Name = %s; want bytes-property", prop.Property.Name)
	}
	if prop.Property.DataType != proto.Property_Bytes {
		t.Errorf("AKV().Property.DataType = %v; want %v", prop.Property.DataType, proto.Property_Bytes)
	}
	if prop.Value == nil {
		t.Error("AKV().Value is nil; want non-nil")
	}
	if !reflect.DeepEqual(prop.Value.Raw, data) {
		t.Errorf("AKV().Value.Raw = %v; want %v", prop.Value.Raw, data)
	}
}

func TestAKV_Int64(t *testing.T) {
	var value int64 = 9223372036854775807
	prop := AKV("int64-property", value)

	if prop.Property.Name != "int64-property" {
		t.Errorf("AKV().Property.Name = %s; want int64-property", prop.Property.Name)
	}
	if prop.Property.DataType != proto.Property_Number {
		t.Errorf("AKV().Property.DataType = %v; want %v", prop.Property.DataType, proto.Property_Number)
	}
	if prop.Value == nil {
		t.Error("AKV().Value is nil; want non-nil")
	}
	if prop.Value.Int != value {
		t.Errorf("AKV().Value.Int = %d; want %d", prop.Value.Int, value)
	}
}

func TestAKV_Int32(t *testing.T) {
	var value int32 = 2147483647
	prop := AKV("int32-property", value)

	if prop.Property.Name != "int32-property" {
		t.Errorf("AKV().Property.Name = %s; want int32-property", prop.Property.Name)
	}
	if prop.Property.DataType != proto.Property_Number {
		t.Errorf("AKV().Property.DataType = %v; want %v", prop.Property.DataType, proto.Property_Number)
	}
	if prop.Value == nil {
		t.Error("AKV().Value is nil; want non-nil")
	}
}

func TestAKV_Float32(t *testing.T) {
	var value float32 = 3.14
	prop := AKV("float32-property", value)

	if prop.Property.Name != "float32-property" {
		t.Errorf("AKV().Property.Name = %s; want float32-property", prop.Property.Name)
	}
	if prop.Property.DataType != proto.Property_Float {
		t.Errorf("AKV().Property.DataType = %v; want %v", prop.Property.DataType, proto.Property_Float)
	}
	if prop.Value == nil {
		t.Error("AKV().Value is nil; want non-nil")
	}
}

func TestAKV_Uint(t *testing.T) {
	var value uint = 42
	prop := AKV("uint-property", value)

	if prop.Property.Name != "uint-property" {
		t.Errorf("AKV().Property.Name = %s; want uint-property", prop.Property.Name)
	}
	// uint maps to Number (stored as int64)
	if prop.Property.DataType != proto.Property_Number {
		t.Errorf("AKV().Property.DataType = %v; want %v", prop.Property.DataType, proto.Property_Number)
	}
}
