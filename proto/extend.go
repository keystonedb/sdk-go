package proto

import "time"

type EntityResponseIDSort []*EntityResponse

func (a EntityResponseIDSort) Len() int      { return len(a) }
func (a EntityResponseIDSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a EntityResponseIDSort) Less(i, j int) bool {
	return a[i].GetEntity().GetEntityId() < a[j].GetEntity().GetEntityId()
}

func CreateDate(time time.Time) *Date {
	return &Date{
		Year:  int32(time.Year()),
		Month: int32(time.Month()),
		Day:   int32(time.Day()),
	}
}

func NewRepeatedValue() *RepeatedValue {
	return &RepeatedValue{KeyValue: make(map[string][]byte), Mixed: make(map[string]*Value)}
}

func (x *RepeatedValue) IsZero() bool {
	return x == nil || (len(x.KeyValue) == 0 && len(x.Strings) == 0 && len(x.Ints) == 0 && len(x.Mixed) == 0)
}
