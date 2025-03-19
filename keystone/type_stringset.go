package keystone

import "github.com/keystonedb/sdk-go/proto"

type StringSet struct {
	values map[string]bool

	toAdd           map[string]bool
	toRemove        map[string]bool
	replaceExisting bool
}

func (s *StringSet) Clear() {
	s.values = nil
	s.toAdd = nil
	s.toRemove = nil
	s.prepare()
}

func (s *StringSet) prepare() {
	if s.toAdd == nil {
		s.toAdd = make(map[string]bool)
	}
	if s.toRemove == nil {
		s.toRemove = make(map[string]bool)
	}
	if s.values == nil {
		s.values = make(map[string]bool)
	}
}

func (s *StringSet) Append(value ...string) {
	for _, v := range value {
		s.Add(v)
	}
}

func (s *StringSet) Add(value string) {
	s.prepare()
	s.toAdd[value] = true
	delete(s.toRemove, value)
}

func (s *StringSet) Reduce(value ...string) {
	for _, v := range value {
		s.Remove(v)
	}
}

func (s *StringSet) Remove(value string) {
	s.prepare()
	s.toRemove[value] = true
	delete(s.toAdd, value)
}

func (s *StringSet) CurrentValues() []string {
	s.prepare()
	var values []string
	for value := range s.values {
		values = append(values, value)
	}
	return values
}

func (s *StringSet) Values() []string {
	s.prepare()
	var values []string

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

func (s *StringSet) Has(value string) bool {
	if s.values == nil {
		return false
	}
	_, ok := s.values[value]
	return ok
}

func (s *StringSet) ReplaceWith(values ...string) {
	s.Clear()
	s.replaceExisting = true
	s.applyValues(values...)
}

func (s *StringSet) applyValues(values ...string) {
	s.prepare()
	for _, value := range values {
		s.values[value] = true
	}
}

func (s *StringSet) IsEmpty() bool {
	return len(s.values) == 0
}

func (s *StringSet) ToAdd() []string {
	s.prepare()
	var values []string
	for value := range s.toAdd {
		values = append(values, value)
	}
	return values
}

func (s *StringSet) ToRemove() []string {
	s.prepare()
	var values []string
	for value := range s.toRemove {
		values = append(values, value)
	}
	return values
}

func (s *StringSet) ReplaceExisting() bool {
	return s.replaceExisting
}

func (s *StringSet) Diff(values ...string) []string {
	check := make(map[string]bool, len(values))
	for _, x := range values {
		check[x] = s.Has(x)
	}
	var diff []string
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

func (s *StringSet) merge() {
	useVals := s.Values()
	s.Clear()
	s.replaceExisting = false
	s.applyValues(useVals...)
}

func NewStringSet(values ...string) StringSet {
	v := StringSet{}
	v.Clear()
	v.applyValues(values...)

	return v
}

func (s *StringSet) MarshalValue() (*proto.Value, error) {
	val := &proto.Value{}
	val.Array = proto.NewRepeatedKeyValue()
	val.Array.Strings = s.Values()

	if len(s.toAdd) > 0 {
		val.ArrayAppend = proto.NewRepeatedKeyValue()
		val.ArrayAppend.Strings = s.ToAdd()
	}

	if len(s.toRemove) > 0 {
		val.ArrayReduce = proto.NewRepeatedKeyValue()
		val.ArrayReduce.Strings = s.ToRemove()
	}

	return val, nil
}

func (s *StringSet) UnmarshalValue(value *proto.Value) error {
	if value != nil {
		if value.Array != nil {
			s.ReplaceWith(value.Array.Strings...)
		}
		if value.ArrayAppend != nil {
			s.Append(value.ArrayAppend.Strings...)
		}
		if value.ArrayReduce != nil {
			s.Reduce(value.ArrayReduce.Strings...)
		}
	}
	return nil
}

func (s *StringSet) IsZero() bool {
	return s == nil ||
		(len(s.values) == 0 &&
			len(s.toAdd) == 0 &&
			len(s.toRemove) == 0)
}

func (s *StringSet) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_StringSet}
}

func (s *StringSet) MutationSuccess(resp *proto.MutateResponse) {
	s.merge()
}
