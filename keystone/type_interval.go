package keystone

import (
	"fmt"
	"math"
	"time"

	"github.com/keystonedb/sdk-go/proto"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type IntervalType string

const (
	IntervalNone       IntervalType = "none"
	IntervalSecond     IntervalType = "sec"
	IntervalMinute     IntervalType = "min"
	IntervalHour       IntervalType = "hour"
	IntervalDay        IntervalType = "day"
	IntervalWeek       IntervalType = "week"
	IntervalMonth      IntervalType = "month"
	IntervalYear       IntervalType = "year"
	IntervalIndefinite IntervalType = "indefinite"
)

// Interval represents a duration like "1 month", "5 days", etc.
type Interval struct {
	Type  IntervalType `json:"type"`
	Count int64        `json:"count"`
}

// NewInterval creates a new Interval
func NewInterval(intervalType IntervalType, count int64) *Interval {
	if intervalType == IntervalNone || intervalType == IntervalIndefinite {
		count = 0
	}
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
	// none and indefinite are singular and don't need pluralization
	if t == IntervalNone || t == IntervalIndefinite {
		return cases.Title(language.English).String(string(t))
	}
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
	if other == nil {
		return true
	}
	if i.Type == other.Type {
		return i.GetCount() > other.GetCount()
	}
	return i.approximateSeconds() > other.approximateSeconds()
}

func (i *Interval) LessThan(other *Interval) bool {
	if i == nil {
		return other != nil
	}
	if other == nil {
		return false
	}
	if i.Type == other.Type {
		return i.GetCount() < other.GetCount()
	}
	return i.approximateSeconds() < other.approximateSeconds()
}

// secondsPerInterval returns the number of seconds for one unit of the given interval type.
// Returns 0 for none, math.MaxInt64 for indefinite, or the approximate seconds otherwise.
func secondsPerInterval(t IntervalType) int64 {
	if t == IntervalNone {
		return 0
	}
	if t == IntervalIndefinite {
		return math.MaxInt64
	}

	multipliers := map[IntervalType]int64{
		IntervalSecond: 1,
		IntervalMinute: 60,
		IntervalHour:   3600,
		IntervalDay:    86400,
		IntervalWeek:   604800,
		IntervalMonth:  2592000,  // 30 days approx.
		IntervalYear:   31536000, // 365 days approx.
	}

	return multipliers[t]
}

func (i *Interval) approximateSeconds() int64 {
	if i == nil {
		return 0
	}
	seconds := secondsPerInterval(i.Type)
	if seconds == math.MaxInt64 {
		return seconds
	}
	return i.Count * seconds
}

// ToDuration converts the interval into a time.Duration.
func (i *Interval) ToDuration() time.Duration {
	if i == nil {
		return 0
	}
	seconds := secondsPerInterval(i.Type)
	if seconds == math.MaxInt64 {
		return time.Duration(math.MaxInt64)
	}
	return time.Duration(i.Count) * time.Second * time.Duration(seconds)
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
		i.Type = IntervalType(value.GetText())
		if i.Type == IntervalNone || i.Type == IntervalIndefinite {
			i.Count = 0
		} else {
			i.Count = value.GetInt()
		}
	}
	return nil
}

// PropertyDefinition returns a generic definition; we store as Text with auxiliary Int count.
// There is no dedicated Interval data type in proto, so Text is the closest fit.
func (i *Interval) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Amount, ExtendedType: proto.Property_Interval}
}
