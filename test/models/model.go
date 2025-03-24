package models

import (
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"time"
)

type Person struct {
	keystone.BaseEntity
	Name         string
	HeightInCm   int64
	DOB          time.Time
	BankBalance  keystone.Amount       `json:",omitempty"`
	PaymentCount int64                 `keystone:"_count_relation:payment"`
	FullName     keystone.SecureString `keystone:",indexed" json:",omitempty"`
	AccountPin   keystone.VerifyString `keystone:",verify"`
	SecretAnswer keystone.SecureString `keystone:",indexed,secure" json:",omitempty"`
}

type Transaction struct {
	keystone.BaseEntity
	Amount      keystone.Amount
	ID          string `keystone:",lookup"`
	PaymentType string
}

func (t Transaction) GetKeystoneDefinition() keystone.TypeDefinition {
	return keystone.TypeDefinition{
		Options: []proto.Schema_Option{proto.Schema_Immutable},
	}
}
