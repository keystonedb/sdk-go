package keystone

import (
	"github.com/keystonedb/sdk-go/sdk-go/proto"
	"testing"
)

func TestAmount_MarshalValue(t *testing.T) {
	amount := NewAmount("USD", 100)
	value, err := amount.MarshalValue()
	if err != nil {
		t.Error(err)
	}
	if value.Text != "USD" {
		t.Error("Text should be USD")
	}
	if value.Int != 100 {
		t.Error("Int should be 100")
	}
}

func TestAmount_UnmarshalValue(t *testing.T) {
	amount := Amount{}
	value := &proto.Value{
		Text: "USD",
		Int:  100,
	}
	err := amount.UnmarshalValue(value)
	if err != nil {
		t.Error(err)
	}
	if amount.Currency != "USD" {
		t.Error("Currency should be USD")
	}
	if amount.Units != 100 {
		t.Error("Units should be 100")
	}
}

func TestNewAmount(t *testing.T) {
	amount := NewAmount("USD", 100)
	if amount.Currency != "USD" {
		t.Error("Currency should be USD")
	}
	if amount.Units != 100 {
		t.Error("Units should be 100")
	}
}
