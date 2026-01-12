package keystone

import (
	"encoding/json"

	"github.com/keystonedb/sdk-go/proto"
)

type Keyed[T any] struct {
	values          map[string]T
	toAdd           map[string]T
	toRemove        map[string]bool
	replaceExisting bool
}

func (s *Keyed[T]) Clear() {
	s.values = nil
	s.toAdd = nil
	s.toRemove = nil
	s.prepare()
}

func (s *Keyed[T]) prepare() {
	if s.toAdd == nil {
		s.toAdd = make(map[string]T)
	}
	if s.toRemove == nil {
		s.toRemove = make(map[string]bool)
	}
	if s.values == nil {
		s.values = make(map[string]T)
	}
}

func (s *Keyed[T]) Set(key string, value T) {
	s.prepare()
	s.values[key] = value
	delete(s.toRemove, key)
}

func (s *Keyed[T]) Append(key string, value T) {
	s.prepare()
	s.toAdd[key] = value
	delete(s.toRemove, key)
}

func (s *Keyed[T]) Remove(key string) {
	s.prepare()
	s.toRemove[key] = true
	delete(s.toAdd, key)
}

func (s *Keyed[T]) Values() map[string]T {
	s.prepare()
	values := make(map[string]T)
	for key, value := range s.values {
		if _, ok := s.toRemove[key]; !ok {
			values[key] = value
		}
	}
	for key, value := range s.toAdd {
		values[key] = value
	}
	return values
}

func (s *Keyed[T]) Has(key string) bool {
	if s.values == nil {
		return false
	}
	_, ok := s.values[key]
	return ok
}

func (s *Keyed[T]) Get(key string) *T {
	s.prepare()
	if value, ok := s.toAdd[key]; ok {
		return &value
	}
	if value, ok := s.values[key]; ok {
		return &value
	}
	return nil
}

func (s *Keyed[T]) IsEmpty() bool {
	return len(s.values) == 0
}

func (s *Keyed[T]) applyValues(with map[string]T) {
	s.values = with
}

func (s *Keyed[T]) merge() {
	useVals := s.Values()
	s.Clear()
	s.replaceExisting = false
	s.values = useVals
}

func NewKeyed[T any](with map[string]T) Keyed[T] {
	v := Keyed[T]{}
	v.Clear()
	if with != nil {
		v.applyValues(with)
	}
	return v
}

func (s *Keyed[T]) MarshalValue() (*proto.Value, error) {
	val := &proto.Value{}
	val.Array = proto.NewRepeatedValue()
	val.KnownType = proto.Property_KeyValue

	for k, v := range s.values {
		data, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		val.Array.KeyValue[k] = data
	}

	if len(s.toAdd) > 0 {
		val.ArrayAppend = proto.NewRepeatedValue()
		for k, v := range s.toAdd {
			data, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			val.ArrayAppend.KeyValue[k] = data
		}
	}

	if len(s.toRemove) > 0 {
		val.ArrayReduce = proto.NewRepeatedValue()
		for key := range s.toRemove {
			val.ArrayReduce.KeyValue[key] = nil
		}
	}

	return val, nil
}

func (s *Keyed[T]) UnmarshalValue(value *proto.Value) error {
	if value != nil {
		if value.Array != nil && value.Array.GetKeyValue() != nil {
			newVal := make(map[string]T)
			for k, v := range value.Array.GetKeyValue() {
				var target T
				if err := json.Unmarshal(v, &target); err == nil {
					newVal[k] = target
				}
			}
			s.applyValues(newVal)
		}
		if value.ArrayAppend != nil && value.ArrayAppend.GetKeyValue() != nil {
			for k, v := range value.ArrayAppend.GetKeyValue() {
				var target T
				if err := json.Unmarshal(v, &target); err == nil {
					s.Set(k, target)
				}
			}
		}
		if value.ArrayReduce != nil && value.ArrayReduce.GetKeyValue() != nil {
			for key := range value.ArrayReduce.GetKeyValue() {
				s.Remove(key)
			}
		}
	}
	return nil
}

func (s *Keyed[T]) IsZero() bool {
	return s == nil ||
		(len(s.values) == 0 &&
			len(s.toAdd) == 0 &&
			len(s.toRemove) == 0)
}

func (s *Keyed[T]) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_KeyValue}
}

func (s *Keyed[T]) ObserveMutation(resp *proto.MutateResponse) {
	if resp.GetSuccess() {
		s.merge()
	}
}
