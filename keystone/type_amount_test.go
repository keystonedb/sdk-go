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
		expect  *Amount
		amounts []*Amount
	}{
		{nil, []*Amount{}},
		{&Amount{"USD", 100}, []*Amount{{"USD", 100}}},
		{&Amount{"USD", 200}, []*Amount{{"USD", 100}, {"USD", 100}}},
		{&Amount{"USD", 300}, []*Amount{{"USD", 100}, {"USD", 100}, {"USD", 100}}},
		{&Amount{CurrencyMixed, 200}, []*Amount{{"GBP", 100}, {"USD", 100}}},
	}

	for _, test := range tests {
		amount := Amounts(test.amounts).Sum()
		if test.expect == nil {
			if amount != nil {
				t.Error("Amount should be nil")
			}
			continue
		}
		if amount.Currency != test.expect.Currency {
			t.Error("Currency should be", test.expect.Currency)
		}
		if amount.Units != test.expect.Units {
			t.Error("Units should be", test.expect.Units)
		}
	}
}

func TestAmountChecks(t *testing.T) {
	amt1 := NewAmount("USD", 100)
	amt1a := NewAmount("USD", 100)
	amt2 := NewAmount("USD", 200)
	amt3 := NewAmount("USD", 300)

	if !amt1.Equals(amt1a) {
		t.Error("100 should equal 100")
	}

	if amt1.GreaterThan(amt2) {
		t.Error("100 should not be greater than 200")
	}

	if !amt2.GreaterThan(amt1) {
		t.Error("200 should be greater than 100")
	}

	if amt3.LessThan(amt2) {
		t.Error("300 should not be less than 200")
	}

	if !amt2.LessThan(amt3) {
		t.Error("200 should be less than 300")
	}
}

func TestAmountDiff(t *testing.T) {
	total := NewAmount("USD", 100)
	preTax := NewAmount("USD", 80)
	diff := total.Diff(preTax)
	if diff.GetUnits() != 20 {
		t.Error("Difference should be 20")
	}
	if diff.GetCurrency() != "USD" {
		t.Error("Currency should be USD")
	}
}
