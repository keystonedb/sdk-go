package keystone

import (
	"testing"
)

func TestCastRaw_FromStruct(t *testing.T) {
	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
		Zip    string `json:"zip"`
	}

	input := Address{
		Street: "123 Main St",
		City:   "Springfield",
		Zip:    "62701",
	}

	m := NewMixed(input)

	if len(m.Raw()) == 0 {
		t.Fatal("expected raw to be populated when creating Mixed from a struct")
	}

	var got Address
	if err := m.CastRaw(&got); err != nil {
		t.Fatalf("CastRaw returned error: %v", err)
	}

	if got != input {
		t.Errorf("CastRaw result = %+v, want %+v", got, input)
	}
}

func TestCastRaw_FromStructPointer(t *testing.T) {
	type Config struct {
		Enabled bool   `json:"enabled"`
		Name    string `json:"name"`
		Count   int    `json:"count"`
	}

	input := &Config{
		Enabled: true,
		Name:    "test",
		Count:   42,
	}

	m := NewMixed(input)

	var got Config
	if err := m.CastRaw(&got); err != nil {
		t.Fatalf("CastRaw returned error: %v", err)
	}

	if got != *input {
		t.Errorf("CastRaw result = %+v, want %+v", got, *input)
	}
}

func TestCastRaw_EmptyRaw(t *testing.T) {
	m := NewMixed("hello")

	var got struct{}
	err := m.CastRaw(&got)
	if err == nil {
		t.Error("expected error when calling CastRaw on Mixed with no raw data")
	}
}
