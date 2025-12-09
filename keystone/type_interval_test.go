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
