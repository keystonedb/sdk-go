package models

import "github.com/keystonedb/sdk-go/keystone"

// AggregateSample is the source entity for the standalone aggregate requirements.
// Aggregate outputs use keystone.AggregateRow, so no summary models are needed.
type AggregateSample struct {
	keystone.BaseEntity
	Run      string  `keystone:",indexed"`
	Category string  `keystone:",indexed"`
	Position int64   `keystone:",indexed"`
	Amount   int64   `keystone:",indexed"`
	Weight   float64 `keystone:",indexed"`
}
