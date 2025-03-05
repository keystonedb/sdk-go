package keystone

import (
	"github.com/keystonedb/sdk-go/proto"
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

func TestSumAmounts(t *testing.T) {
	tests := []struct {
		expect  Amount
		amounts []Amount
	}{
		{Amount{"", 0}, []Amount{}},
		{Amount{"USD", 100}, []Amount{{"USD", 100}}},
		{Amount{"USD", 200}, []Amount{{"USD", 100}, {"USD", 100}}},
		{Amount{"USD", 300}, []Amount{{"USD", 100}, {"USD", 100}, {"USD", 100}}},
		{Amount{CurrencyMixed, 200}, []Amount{{"GBP", 100}, {"USD", 100}}},
	}

	for _, test := range tests {
		amount := SumAmounts(test.amounts...)
		if amount.Currency != test.expect.Currency {
			t.Error("Currency should be", test.expect.Currency)
		}
		if amount.Units != test.expect.Units {
			t.Error("Units should be", test.expect.Units)
		}
	}
}
