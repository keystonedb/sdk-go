package models

import (
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"time"
)

type DataTypes struct {
	keystone.BaseEntity
	String       string
	Integer      int64
	Time         time.Time
	Amount       keystone.Amount
	Secret       keystone.SecureString
	Verify       keystone.VerifyString
	Boolean      bool
	Float        float64
	Map          map[string]string
	StringSlice  []string
	IntegerSlice []int
	StringSet    keystone.StringSet
	IntegerSet   keystone.IntSet
	RawData      []byte
}

func (t DataTypes) GetKeystoneDefinition() keystone.TypeDefinition {
	return keystone.TypeDefinition{
		Options: []proto.Schema_Option{proto.Schema_StoreMutations},
	}
}