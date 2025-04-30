package keystone

import "github.com/keystonedb/sdk-go/proto"

type IntSet struct {
	values map[int64]bool

	toAdd           map[int64]bool
	toRemove        map[int64]bool
	replaceExisting bool
}

func (s *IntSet) Clear() {
	s.values = nil
	s.toAdd = nil
	s.toRemove = nil
	s.prepare()
}

func (s *IntSet) prepare() {
	if s.toAdd == nil {
		s.toAdd = make(map[int64]bool)
	}
	if s.toRemove == nil {
		s.toRemove = make(map[int64]bool)
	}
	if s.values == nil {
		s.values = make(map[int64]bool)
	}
}

func (s *IntSet) Append(value ...int64) {
	for _, v := range value {
		s.Add(v)
	}
}

func (s *IntSet) Add(value int64) {
	s.prepare()
	s.toAdd[value] = true
	delete(s.toRemove, value)
}

func (s *IntSet) Reduce(value ...int64) {
	for _, v := range value {
		s.Remove(v)
	}
}

func (s *IntSet) Remove(value int64) {
	s.prepare()
	s.toRemove[value] = true
	delete(s.toAdd, value)
}

func (s *IntSet) CurrentValues() []int64 {
	s.prepare()
	var values []int64
	for value := range s.values {
		values = append(values, value)
	}
	return values
}

func (s *IntSet) Values() []int64 {
	s.prepare()
	var values []int64

	for value := range s.values {
		if _, ok := s.toRemove[value]; !ok {
			values = append(values, value)
		}
	}

	for value := range s.toAdd {
		values = append(values, value)
	}
	return values
}

func (s *IntSet) Has(value int64) bool {
	if s.values == nil {
		return false
	}
	_, ok := s.values[value]
	return ok
}

func (s *IntSet) ReplaceWith(values ...int64) {
	s.Clear()
	s.replaceExisting = true
	s.applyValues(values...)
}

func (s *IntSet) applyValues(values ...int64) {
	s.prepare()
	for _, value := range values {
		s.values[value] = true
	}
}

func (s *IntSet) IsEmpty() bool {
	return len(s.values) == 0
}

func (s *IntSet) ToAdd() []int64 {
	s.prepare()
	var values []int64
	for value := range s.toAdd {
		values = append(values, value)
	}
	return values
}

func (s *IntSet) ToRemove() []int64 {
	s.prepare()
	var values []int64
	for value := range s.toRemove {
		values = append(values, value)
	}
	return values
}

func (s *IntSet) ReplaceExisting() bool {
	return s.replaceExisting
}

func (s *IntSet) Diff(values ...int64) []int64 {
	s.prepare()
	check := make(map[int64]bool, len(values))
	for _, x := range values {
		check[x] = s.Has(x)
	}
	var diff []int64
	for x := range s.values {
		if _, ok := check[x]; !ok {
			diff = append(diff, x)
		}
	}
	for x, matched := range check {
		if !matched {
			diff = append(diff, x)
		}
	}
	return diff
}

func (s *IntSet) merge() {
	useVals := s.CurrentValues()
	s.Clear()
	s.replaceExisting = false
	s.applyValues(useVals...)
}

func NewIntSet(values ...int64) IntSet {
	v := IntSet{}
	v.Clear()
	v.applyValues(values...)
	return v
}

func (s *IntSet) MarshalValue() (*proto.Value, error) {
	val := &proto.Value{}
	val.Array = proto.NewRepeatedValue()
	val.Array.Ints = s.Values()

	if len(s.toAdd) > 0 {
		val.ArrayAppend = proto.NewRepeatedValue()
		val.ArrayAppend.Ints = s.ToAdd()
	}

	if len(s.toRemove) > 0 {
		val.ArrayReduce = proto.NewRepeatedValue()
		val.ArrayReduce.Ints = s.ToRemove()
	}

	return val, nil
}

func (s *IntSet) UnmarshalValue(value *proto.Value) error {
	if value != nil {
		if value.Array != nil {
			s.ReplaceWith(value.Array.Ints...)
		}
		if value.ArrayAppend != nil {
			s.Append(value.ArrayAppend.Ints...)
		}
		if value.ArrayReduce != nil {
			s.Reduce(value.ArrayReduce.Ints...)
		}
	}
	return nil
}

func (s *IntSet) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_IntSet}
}

func (s *IntSet) ObserveMutation(resp *proto.MutateResponse) {
	if resp.GetSuccess() {
		s.merge()
	}
}

func (s *IntSet) IsZero() bool {
	return s == nil ||
		(len(s.values) == 0 &&
			len(s.toAdd) == 0 &&
			len(s.toRemove) == 0)
}
