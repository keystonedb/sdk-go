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
func NewAmount(currency string, units int64) *Amount {
	return &Amount{
		Currency: currency,
		Units:    units,
	}
}

func (a *Amount) GetUnits() int64 {
	if a == nil {
		return 0
	}
	return a.Units
}

func (a *Amount) GetCurrency() string {
	if a == nil {
		return ""
	}
	return a.Currency
}

func (a *Amount) IsZero() bool {
	return a == nil || a.Units == 0 && a.Currency == ""
}

func (a *Amount) Equals(other *Amount) bool {
	return a.GetCurrency() == other.GetCurrency() && a.GetUnits() == other.GetUnits()
}

func (a *Amount) GreaterThan(other *Amount) bool {
	return a.GetUnits() > other.GetUnits()
}

func (a *Amount) LessThan(other *Amount) bool {
	return a.GetUnits() < other.GetUnits()
}

/*
#Not sure if these would be best returning a new amount, or updating the existing
#Leaving unavailable until a decision is made here

func (a *Amount) Add(u ...*Amount) *Amount {
	for _, unit := range u {
		a.Units += unit.GetUnits()
	}
	return a
}

func (a *Amount) Sub(u ...*Amount) *Amount {
	for _, unit := range u {
		a.Units -= unit.GetUnits()
	}
	return a
}*/

func (a *Amount) MarshalValue() (*proto.Value, error) {
	if a.IsZero() {
		return nil, nil
	}
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

type Amounts []*Amount

func (a Amounts) Sum() *Amount {
	switch len(a) {
	case 0:
		return nil
	case 1:
		return a[0]
	}
	ret := &Amount{}
	for _, amt := range a {
		if amt == nil {
			continue
		}
		ret.Units += amt.Units
		if ret.Currency == "" {
			ret.Currency = amt.Currency
		} else if ret.Currency != amt.Currency {
			ret.Currency = "mixed"
		}
	}

	if ret.IsZero() {
		return nil
	}

	return ret
}

func (a Amounts) Max() *Amount {
	var res *Amount
	for _, amt := range a {
		if res == nil || amt.Units > res.Units {
			res = amt
		}
	}
	return res
}

func (a Amounts) Min() *Amount {
	var res *Amount
	for _, amt := range a {
		if res == nil || amt.Units < res.Units {
			res = amt
		}
	}
	return res
}
