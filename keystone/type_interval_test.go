package keystone

import (
	"testing"

	"github.com/keystonedb/sdk-go/proto"
)

func Test_Interval_Basics(t *testing.T) {
	i := NewInterval(IntervalMonth, 1)
	if i.GetType() != IntervalMonth {
		t.Fatalf("type mismatch: got %s", i.GetType())
	}
	if i.GetCount() != 1 {
		t.Fatalf("count mismatch: got %d", i.GetCount())
	}
	if i.String() != "1 month" {
		t.Fatalf("string mismatch: got %s", i.String())
	}

	i2 := NewInterval(IntervalMonth, 2)
	if i2.String() != "2 months" {
		t.Fatalf("plural string mismatch: got %s", i2.String())
	}

	if i.Equals(i2) {
		t.Fatalf("equals should be false")
	}
	if !i2.GreaterThan(i) || !i.LessThan(i2) {
		t.Fatalf("comparison helpers failed")
	}

	d := i2.Diff(i)
	if d == nil || d.GetType() != IntervalMonth || d.GetCount() != 1 {
		t.Fatalf("diff unexpected: %+v", d)
	}
}

func Test_Interval_Marshal_Unmarshal(t *testing.T) {
	src := NewInterval(IntervalWeek, 5)
	pv, err := src.MarshalValue()
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}
	if pv == nil {
		t.Fatalf("marshal produced nil value")
	}

	var dst Interval
	if err := dst.UnmarshalValue(pv); err != nil {
		t.Fatalf("unmarshal err: %v", err)
	}

	if !src.Equals(&dst) {
		t.Fatalf("round-trip mismatch: src=%+v dst=%+v", src, dst)
	}
}

func Test_Interval_PropertyDefinition_IsZero(t *testing.T) {
	var z Interval
	if !z.IsZero() {
		t.Fatalf("expected zero interval")
	}

	nz := NewInterval(IntervalDay, 0)
	if nz.IsZero() {
		t.Fatalf("unexpected zero when type set")
	}

	pd := nz.PropertyDefinition()
	if pd.DataType != proto.Property_Amount { // matches current implementation
		t.Fatalf("unexpected property datatype: %v", pd.DataType)
	}
}

func Test_Interval_NoneAndIndefinite_CountZero(t *testing.T) {
	tests := []struct {
		name     string
		iType    IntervalType
		inCount  int64
		outCount int64
	}{
		{"None with 10", IntervalNone, 10, 0},
		{"Indefinite with 5", IntervalIndefinite, 5, 0},
		{"Second with 10", IntervalSecond, 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := NewInterval(tt.iType, tt.inCount)
			if got := i.GetCount(); got != tt.outCount {
				t.Fatalf("NewInterval(%s, %d) count = %d; want %d", tt.iType, tt.inCount, got, tt.outCount)
			}
		})
	}
}

func Test_Interval_Unmarshal_None(t *testing.T) {
	pv := &proto.Value{
		Text: string(IntervalNone),
		Int:  123,
	}
	var i Interval
	if err := i.UnmarshalValue(pv); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if i.Count != 0 {
		t.Fatalf("UnmarshalValue(None, 123) count = %d; want 0", i.Count)
	}
}

func Test_Interval_Unmarshal_Indefinite(t *testing.T) {
	pv := &proto.Value{
		Text: string(IntervalIndefinite),
		Int:  123,
	}
	var i Interval
	if err := i.UnmarshalValue(pv); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if i.Count != 0 {
		t.Fatalf("UnmarshalValue(Indefinite, 123) count = %d; want 0", i.Count)
	}
}

func Test_Interval_Comparisons_CrossTypes(t *testing.T) {
	tests := []struct {
		name        string
		a           *Interval
		b           *Interval
		wantGreater bool
		wantLess    bool
	}{
		{
			name:        "1 hour vs 30 minutes",
			a:           NewInterval(IntervalHour, 1),
			b:           NewInterval(IntervalMinute, 30),
			wantGreater: true,
			wantLess:    false,
		},
		{
			name:        "1 hour vs 60 minutes (equal)",
			a:           NewInterval(IntervalHour, 1),
			b:           NewInterval(IntervalMinute, 60),
			wantGreater: false,
			wantLess:    false,
		},
		{
			name:        "61 seconds vs 1 minute",
			a:           NewInterval(IntervalSecond, 61),
			b:           NewInterval(IntervalMinute, 1),
			wantGreater: true,
			wantLess:    false,
		},
		{
			name:        "1 day vs 25 hours",
			a:           NewInterval(IntervalDay, 1),
			b:           NewInterval(IntervalHour, 25),
			wantGreater: false,
			wantLess:    true,
		},
		{
			name:        "Indefinite vs 100 years",
			a:           NewInterval(IntervalIndefinite, 0),
			b:           NewInterval(IntervalYear, 100),
			wantGreater: true,
			wantLess:    false,
		},
		{
			name:        "None vs 1 second",
			a:           NewInterval(IntervalNone, 0),
			b:           NewInterval(IntervalSecond, 1),
			wantGreater: false,
			wantLess:    true,
		},
		{
			name:        "1 Month vs 29 Days",
			a:           NewInterval(IntervalMonth, 1),
			b:           NewInterval(IntervalDay, 29),
			wantGreater: true,
			wantLess:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.GreaterThan(tt.b); got != tt.wantGreater {
				t.Errorf("GreaterThan() = %v, want %v", got, tt.wantGreater)
			}
			if got := tt.a.LessThan(tt.b); got != tt.wantLess {
				t.Errorf("LessThan() = %v, want %v", got, tt.wantLess)
			}
		})
	}
}
