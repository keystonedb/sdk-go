package keystone

import (
	"bytes"
	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strconv"
	"time"
)

// Mixed is a mixed value
type Mixed struct {
	text  string
	int   int64
	bool  bool
	float float64
	time  time.Time
	raw   []byte
}

// NewMixed creates a new Mixed
func NewMixed(mixedInput any) Mixed {
	m := Mixed{}
	m.SetValue(mixedInput)
	return m
}

func (m *Mixed) ToString() string {
	if m == nil {
		return ""
	}
	if m.text != "" {
		return m.text
	}
	if m.int != 0 {
		return strconv.FormatInt(m.int, 10)
	}
	if m.bool {
		return strconv.FormatBool(m.bool)
	}
	if m.float != 0 {
		return strconv.FormatFloat(m.float, 'f', -1, 64)
	}
	if !m.time.IsZero() {
		return m.time.Format(time.RFC3339)
	}
	if len(m.raw) != 0 {
		return string(m.raw)
	}
	return ""
}

func (m *Mixed) String() string {
	if m == nil {
		return ""
	}
	if m.text != "" {
		return m.text
	}
	return ""
}

func (m *Mixed) Int() int64 {
	if m == nil {
		return 0
	}
	return m.int
}

func (m *Mixed) Bool() bool {
	if m == nil {
		return false
	}
	return m.bool
}

func (m *Mixed) Float() float64 {
	if m == nil {
		return 0
	}
	return m.float
}

func (m *Mixed) Time() time.Time {
	if m == nil {
		return time.Time{}
	}
	return m.time
}
func (m *Mixed) Raw() []byte {
	if m == nil {
		return nil
	}
	return m.raw
}

func (m *Mixed) SetString(text string) {
	if m == nil {
		return
	}
	m.text = text
}

func (m *Mixed) SetInt(i int64) {
	if m == nil {
		return
	}
	m.int = i
}

func (m *Mixed) SetBool(b bool) {
	if m == nil {
		return
	}
	m.bool = b
}

func (m *Mixed) SetFloat(f float64) {
	if m == nil {
		return
	}
	m.float = f
}
func (m *Mixed) SetTime(t time.Time) {
	if m == nil {
		return
	}
	m.time = t
}
func (m *Mixed) SetRaw(raw []byte) {
	if m == nil {
		return
	}
	m.raw = raw
}
func (m *Mixed) SetValue(value any) {
	if m == nil || value == nil {
		return
	}

	switch v := value.(type) {
	case string:
		m.text = v
	case int:
		m.int = int64(v)
	case int32:
		m.int = int64(v)
	case int64:
		m.int = v
	case bool:
		m.bool = v
	case float64:
		m.float = v
	case float32:
		m.float = float64(v)
	case time.Time:
		m.time = v
	case *time.Time:
		m.time = *v
	case timestamppb.Timestamp:
		m.time = v.AsTime()
	case *timestamppb.Timestamp:
		m.time = v.AsTime()
	case []byte:
		m.raw = v
	}
}

func (m *Mixed) MarshalValue() (*proto.Value, error) {
	var protoTime *timestamppb.Timestamp
	if !m.time.IsZero() {
		protoTime = timestamppb.New(m.time.UTC())
	}
	return &proto.Value{
		Text:  m.text,
		Int:   m.int,
		Bool:  m.bool,
		Float: m.float,
		Time:  protoTime,
		Raw:   m.raw,
	}, nil
}

func (m *Mixed) UnmarshalValue(value *proto.Value) error {
	if value != nil {
		m.text = value.Text
		m.int = value.Int
		m.bool = value.Bool
		m.float = value.Float
		if value.Time != nil {
			m.time = value.Time.AsTime()
		}
		m.raw = value.Raw
	}
	return nil
}

func (m *Mixed) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Mixed}
}

func (m *Mixed) IsZero() bool {
	return m == nil || (m.text == "" && m.int == 0 && m.bool == false && m.float == 0 && m.time.IsZero() && len(m.raw) == 0)
}

func (m *Mixed) Matches(with *Mixed) bool {
	if m == nil && with == nil {
		return true
	}
	if m == nil || with == nil {
		return false
	}
	if m.text != with.text {
		return false
	}
	if m.int != with.int {
		return false
	}
	if m.bool != with.bool {
		return false
	}
	if m.float != with.float {
		return false
	}
	if m.time.UnixMilli() != with.time.UnixMilli() {
		return false
	}
	if !bytes.Equal(m.raw, with.raw) {
		return false
	}
	return true
}
