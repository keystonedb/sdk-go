package models

import (
	"github.com/keystonedb/sdk-go/keystone"
	"time"
)

type FileData struct {
	keystone.BaseEntity

	UserID          string `keystone:",indexed"`
	Submitted       time.Time
	State           int    `keystone:",indexed"`
	ConnectorID     string `keystone:",indexed"`
	IsPending       bool   `keystone:",indexed"`
	CheckKey        string `keystone:",indexed"`
	LineInformation string
}

func (f FileData) GetKeystoneDefinition() keystone.TypeDefinition {
	return keystone.TypeDefinition{}
}
