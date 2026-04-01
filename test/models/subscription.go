package models

import (
	"time"

	"github.com/keystonedb/sdk-go/keystone"
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
	CreationDate time.Time `keystone:"_created"`
	PaymentDate  time.Time `keystone:",indexed"`
	Notes        []*RenewalNote

	NumNotes int64 `keystone:"_child_count:renewal-note"`
}

type RenewalNote struct {
	keystone.Child
	Date time.Time
	Note string
}
