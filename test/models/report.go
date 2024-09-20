package models

import "github.com/keystonedb/sdk-go/keystone"

type ReportedEvent struct {
	keystone.TimeSeriesEntity
	keystone.BaseEntity
	Library         string `keystone:",metricFilter"`
	Converted       bool   `keystone:",metric"`
	ResultCode      string `keystone:",metric"`
	LibraryID       string `keystone:",metricFilter"`
	ResponseMessage string `keystone:",metric"`
	BinData         string
}
