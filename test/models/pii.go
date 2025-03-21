package models

import "github.com/keystonedb/sdk-go/keystone"

type PiiPerson struct {
	keystone.BaseEntity
	Name   keystone.PersonName
	Email  keystone.Email
	Phone  keystone.Phone
	NonPii string
}
