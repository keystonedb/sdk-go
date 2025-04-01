package models

import (
	"crypto/sha1"
	"fmt"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
)

type WithPrimary struct {
	keystone.BaseEntity
	NameHash  string `keystone:",primary"`
	FirstName string
	LastName  string
}

func (w *WithPrimary) SetHash() {
	toHash := w.FirstName + " " + w.LastName
	w.NameHash = fmt.Sprintf("%x", sha1.Sum([]byte(toHash)))[:12]
}

func (w WithPrimary) GetKeystoneDefinition() keystone.TypeDefinition {
	return keystone.TypeDefinition{
		Options: []proto.Schema_Option{proto.Schema_Immutable},
	}
}
