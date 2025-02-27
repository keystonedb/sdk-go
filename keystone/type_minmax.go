package keystone

import (
	"github.com/keystonedb/sdk-go/proto"
)

type MinMax struct {
	min int64
	max int64
}

func (s *MinMax) Min() int64 {
	return s.min
}

func (s *MinMax) Max() int64 {
	return s.max
}

func (s *MinMax) SetMin(value int64) {
	s.min = value
}

func (s *MinMax) SetMax(value int64) {
	s.max = value
}

func (s *MinMax) Update(min, max int64) {
	s.SetMin(min)
	s.SetMax(max)
}

func NewMinMax(min, max int64) MinMax {
	return MinMax{min: min, max: max}
}

func (s *MinMax) MarshalValue() (*proto.Value, error) {
	val := &proto.Value{}
	val.Array = proto.NewRepeatedKeyValue()
	val.Array.Ints = []int64{s.min, s.max}
	return val, nil
}

func (s *MinMax) UnmarshalValue(value *proto.Value) error {
	if value != nil {
		if value.Array != nil && len(value.Array.Ints) > 0 {
			// get the min and max values from value.Array.Ints
			value.Array.SortInts()
			s.min = value.Array.Ints[0]
			s.max = value.Array.Ints[len(value.Array.Ints)-1]
		}
	}
	return nil
}

func (s *MinMax) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Ints}
}
