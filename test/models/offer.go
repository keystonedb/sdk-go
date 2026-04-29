package models

import "github.com/keystonedb/sdk-go/keystone"

type Offer struct {
	keystone.BaseEntity
	DisplayName string
	TestRun     string   `keystone:",indexed"`
	IsGlobal    bool     `keystone:",indexed"`
	ProductIDs  []string `keystone:",indexed"`
}
