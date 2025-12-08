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
	StringPtr    *string
	Integer      int64
	IntegerPtr   *int64
	Time         time.Time
	TimePt       *time.Time
	Amount       keystone.Amount
	AmountPt     *keystone.Amount
	Interval     keystone.Interval
	IntervalPt   *keystone.Interval
	Secret       keystone.SecureString `keystone:",no-snapshot"`
	Verify       keystone.VerifyString
	Boolean      bool
	BooleanPtr   *bool
	Float        float64
	FloatPtr     *float64
	Map          map[string]string
	StringSlice  []string
	IntegerSlice []int
	StringSet    keystone.StringSet
	IntegerSet   keystone.IntSet
	RawData      []byte
	MinMax       keystone.MinMax
	Flags        []EnumValue
	EnumValue    EnumValue
	ExternalID   keystone.ExternalID
	Mixed        keystone.Mixed
	MixedKey     keystone.KeyMixed
	IDSlice      []keystone.ID
}

func (t *DataTypes) GetKeystoneDefinition() keystone.TypeDefinition {
	return keystone.TypeDefinition{
		Options: []proto.Schema_Option{proto.Schema_StoreMutations},
	}
}
