package models

import (
	"time"

	"github.com/keystonedb/sdk-go/keystone"
)

type DynamicRemote struct {
	keystone.DynamicRemoteEntity
	String       string
	Integer      int64
	Time         time.Time
	Amount       keystone.Amount
	AmountPt     *keystone.Amount
	Secret       keystone.SecureString `keystone:",no-snapshot"`
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
	ExternalID   keystone.ExternalID
	Mixed        keystone.Mixed
	MixedKey     keystone.KeyMixed
}
