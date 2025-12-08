package keystone

import (
	"fmt"

	"github.com/keystonedb/sdk-go/proto"
)

type IntervalType string

const (
	IntervalSecond IntervalType = "sec"
	IntervalMinute IntervalType = "min"
	IntervalHour   IntervalType = "hour"
	IntervalDay    IntervalType = "day"
	IntervalWeek   IntervalType = "week"
	IntervalMonth  IntervalType = "month"
	IntervalYear   IntervalType = "year"
)

// Interval represents a duration like "1 month", "5 days", etc.
type Interval struct {
	Type  IntervalType `json:"type"`
	Count int64        `json:"count"`
}

// NewInterval creates a new Interval
func NewInterval(intervalType IntervalType, count int64) *Interval {
	return &Interval{
		Type:  intervalType,
		Count: count,
	}
}

func (i *Interval) GetType() IntervalType {
	if i == nil {
		return ""
	}
	return i.Type
}

func (i *Interval) GetCount() int64 {
	if i == nil {
		return 0
	}
	return i.Count
}

func (i *Interval) String() string {
	if i == nil {
		return ""
	}
	t := i.Type
	c := i.Count
	// naive pluralization: add 's' when count != 1
	if c == 1 || c == -1 {
		return fmt.Sprintf("%d %s", c, t)
	}
	// add trailing 's' if not already pluralized
	if t != "" && t[len(t)-1] != 's' {
		t = t + "s"
	}
	return fmt.Sprintf("%d %s", c, t)
}

func (i *Interval) IsZero() bool {
	return i == nil || (i.Count == 0 && i.Type == "")
}

func (i *Interval) Equals(other *Interval) bool {
	if i == nil && other == nil {
		return true
	}
	if i == nil || other == nil {
		return false
	}
	return i.GetType() == other.GetType() && i.GetCount() == other.GetCount()
}

func (i *Interval) GreaterThan(other *Interval) bool {
	if i == nil {
		return false
	}
	return i.GetCount() > other.GetCount()
}

func (i *Interval) LessThan(other *Interval) bool {
	if i == nil {
		return other != nil
	}
	return i.GetCount() < other.GetCount()
}

func (i *Interval) Diff(with *Interval) *Interval {
	if i == nil || with == nil {
		return nil
	}
	if i.GetType() != with.GetType() {
		return nil
	}
	return &Interval{
		Type:  i.GetType(),
		Count: i.GetCount() - with.GetCount(),
	}
}

func (i *Interval) MarshalValue() (*proto.Value, error) {
	if i.IsZero() {
		return nil, nil
	}
	return &proto.Value{
		Text: string(i.Type),
		Int:  i.Count,
	}, nil
}

func (i *Interval) UnmarshalValue(value *proto.Value) error {
	if value != nil {
		i.Count = value.GetInt()
		i.Type = IntervalType(value.GetText())
	}
	return nil
}

// PropertyDefinition returns a generic definition; we store as Text with auxiliary Int count.
// There is no dedicated Interval data type in proto, so Text is the closest fit.
func (i *Interval) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Amount, ExtendedType: proto.Property_Interval}
}
