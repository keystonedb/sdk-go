package models

import "github.com/keystonedb/sdk-go/keystone"

type User struct {
	keystone.BaseEntity
	ExternalID string `keystone:",unique"`
	Validate   string
}
