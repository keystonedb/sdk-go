package keystone

import (
	"github.com/keystonedb/sdk-go/proto"
)

// Amount represents money
type Amount struct {
	Currency string `json:"currency"`
	Units    int64  `json:"units"`
}

// NewAmount creates a new Amount
func NewAmount(currency string, units int64) Amount {
	return Amount{
		Currency: currency,
		Units:    units,
	}
}

func (a *Amount) MarshalValue() (*proto.Value, error) {
	return &proto.Value{
		Text: a.Currency,
		Int:  a.Units,
	}, nil
}

func (a *Amount) UnmarshalValue(value *proto.Value) error {
	if value != nil {
		a.Units = value.GetInt()
		a.Currency = value.GetText()
	}
	return nil
}

func (a *Amount) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Amount}
}

const CurrencyMixed = "mixed"

func SumAmounts(amounts ...Amount) Amount {
	switch len(amounts) {
	case 0:
		return Amount{}
	case 1:
		return amounts[0]
	}

	ret := Amount{
		Currency: amounts[0].Currency,
	}
	for _, amt := range amounts {
		ret.Units += amt.Units
		if ret.Currency != amt.Currency {
			ret.Currency = CurrencyMixed
		}
	}
	return ret
}
