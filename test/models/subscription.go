package models

import (
	"github.com/keystonedb/sdk-go/keystone"
	"time"
)

type Subscription struct {
	keystone.BaseEntity
	StartDate        time.Time
	NumberOfRenewals int `keystone:"_count_descendant:renewal"`
}

type Renewal struct {
	keystone.BaseChildEntity
	StartDate    time.Time
	EndDate      time.Time
	CreationDate time.Time
	PaymentDate  time.Time
}
