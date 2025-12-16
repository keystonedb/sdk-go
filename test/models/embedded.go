package models

import "github.com/keystonedb/sdk-go/keystone"

type Embedded struct {
	keystone.BaseEntity
	Name        string
	Extended    ExtendedData
	ExtendedRef *ExtendedData
}

type ExtendedData struct {
	LookupValue string `keystone:",lookup,query"`
	UniqueID    string `keystone:",unique"`
	Price       keystone.Amount
	BoolValue   bool
	StringValue string
}
