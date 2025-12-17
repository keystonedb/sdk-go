package keystone

import (
	"testing"

	"github.com/keystonedb/sdk-go/proto"
)

func TestMinMax_UnmarshalValue(t *testing.T) {
	tests := []struct {
		name  string
		min   int64
		max   int64
		value []int64
	}{
		{"Basic", 0, 1, []int64{0, 1}},
		{"Flipped", 0, 1, []int64{1, 0}},
		{"Single", 1, 1, []int64{1}},
		{"Empty", 0, 0, []int64{}},
		{"Negative", -1, 1, []int64{-1, 1}},
		{"Negative Flipped", -1, 1, []int64{1, -1}},
		{"Negative Single", -1, -1, []int64{-1}},
		{"Simple", 10, 40, []int64{10, 40}},
		{"Problem", 10, 50, []int64{10, 20, 30, 40, 50}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MinMax{}
			val := &proto.Value{
				Array: &proto.RepeatedValue{
					Ints: tt.value,
				},
			}
			if err := s.UnmarshalValue(val); err != nil {
				t.Errorf("MinMax.UnmarshalValue() error = %v", err)
			} else {
				if s.Min() != tt.min {
					t.Errorf("Expected Min to be %d, got %d", tt.min, s.Min())
				}
				if s.Max() != tt.max {
					t.Errorf("Expected Max to be %d, got %d", tt.max, s.Max())
				}
			}
		})
	}
}
