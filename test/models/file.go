package models

import "github.com/keystonedb/sdk-go/keystone"

type File struct {
	keystone.BaseEntity
	FileName string
	Size     int64
	Lines    []*FileLine
}

type FileLine struct {
	keystone.Child
	LineID     string `keystone:"_child_id"`
	Data       []byte
	LineNumber int64
}

func (e *FileLine) AggregateValue() int64 {
	return e.LineNumber
}

func (e *FileLine) SetAggregateValue(value int64) {
	e.LineNumber = value
}

func (e *FileLine) KeystoneData() map[string][]byte {
	if e == nil {
		return nil
	}
	return map[string][]byte{
		"data": e.Data,
	}
}

func (e *FileLine) HydrateKeystoneData(data map[string][]byte) {
	if e != nil {
		e.Data = data["data"]
	}
}