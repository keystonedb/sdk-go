package models

import (
	"github.com/keystonedb/sdk-go/keystone"
)

type File struct {
	keystone.BaseEntity
	FileName string
	Size     int64
	Lines    []*FileLine
}

type FileLine struct {
	keystone.Child
	LineID     string
	Data       []byte
	LineNumber int64
}

func (e *File) Len() int           { return len(e.Lines) }
func (e *File) Less(i, j int) bool { return e.Lines[i].LineNumber < e.Lines[j].LineNumber }
func (e *File) Swap(i, j int)      { e.Lines[i], e.Lines[j] = e.Lines[j], e.Lines[i] }

func (e *FileLine) SetChildID(id string) {
	e.Child.SetChildID(id)
	e.LineID = id
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

func (e *FileLine) FromKeystoneData(data map[string][]byte) {
	if e != nil {
		e.Data = data["data"]
	}
}
