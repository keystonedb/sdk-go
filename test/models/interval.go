package models

import (
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
)

type IntervalEntity struct {
	keystone.BaseEntity
	Name      string            `keystone:",indexed"`
	Period    keystone.Interval `keystone:",indexed"`
	PeriodPtr *keystone.Interval
}

func (t *IntervalEntity) GetKeystoneDefinition() keystone.TypeDefinition {
	return keystone.TypeDefinition{Options: []proto.Schema_Option{proto.Schema_StoreMutations}}
}
