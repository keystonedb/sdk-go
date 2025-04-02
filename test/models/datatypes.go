package models

import (
	"time"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
)

type EnumValue int32

const (
	ENUM_VALUE0 EnumValue = 0
	ENUM_VALUE1 EnumValue = 1
)

type DataTypes struct {
	keystone.BaseEntity
	String       string
	Integer      int64
	Time         time.Time
	Amount       keystone.Amount
	AmountPt     *keystone.Amount
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
	MinMax       keystone.MinMax
	Flags        []EnumValue
	EnumValue    EnumValue
}

func (t *DataTypes) GetKeystoneDefinition() keystone.TypeDefinition {
	return keystone.TypeDefinition{
		Options: []proto.Schema_Option{proto.Schema_StoreMutations},
	}
}
