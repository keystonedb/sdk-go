package models

import (
	"crypto/sha256"
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
	hash := sha256.Sum256([]byte(toHash))
	w.NameHash = fmt.Sprintf("%x", hash)[:12]
}

func (w *WithPrimary) GetKeystoneDefinition() keystone.TypeDefinition {
	return keystone.TypeDefinition{
		Options: []proto.Schema_Option{proto.Schema_Immutable},
	}
}
