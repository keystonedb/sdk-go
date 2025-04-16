package models

import (
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
)

type Hashi struct {
	keystone.BaseEntity
	FirstName string
	LastName  string
}

func (w *Hashi) GetKeystoneDefinition() keystone.TypeDefinition {
	return keystone.TypeDefinition{
		Options: []proto.Schema_Option{proto.Schema_HashedID},
	}
}
