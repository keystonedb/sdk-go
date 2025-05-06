package keystone

import "github.com/keystonedb/sdk-go/proto"

// KeyMixed is a map of mixed values
type KeyMixed struct {
	values          map[string]Mixed
	toAdd           map[string]Mixed
	toRemove        map[string]bool
	replaceExisting bool
}

func (s *KeyMixed) Clear() {
	s.values = nil
	s.toAdd = nil
	s.toRemove = nil
	s.prepare()
}

func (s *KeyMixed) prepare() {
	if s.toAdd == nil {
		s.toAdd = make(map[string]Mixed)
	}
	if s.toRemove == nil {
		s.toRemove = make(map[string]bool)
	}
	if s.values == nil {
		s.values = make(map[string]Mixed)
	}
}

func (s *KeyMixed) Set(key string, value Mixed) {
	s.prepare()
	s.values[key] = value
	delete(s.toRemove, key)
}

func (s *KeyMixed) Append(key string, value Mixed) {
	s.prepare()
	s.toAdd[key] = value
	delete(s.toRemove, key)
}

func (s *KeyMixed) Remove(key string) {
	s.prepare()
	s.toRemove[key] = true
	delete(s.toAdd, key)
}

func (s *KeyMixed) Values() map[string]Mixed {
	s.prepare()
	values := make(map[string]Mixed)
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

func (s *KeyMixed) Has(value string) bool {
	if s.values == nil {
		return false
	}
	_, ok := s.values[value]
	return ok
}

func (s *KeyMixed) Get(key string) *Mixed {
	s.prepare()
	if value, ok := s.toAdd[key]; ok {
		return &value
	}
	if value, ok := s.values[key]; ok {
		return &value
	}
	return nil
}

func (s *KeyMixed) Diff(with map[string]Mixed) map[string]Mixed {
	s.prepare()
	diff := s.values
	for key, value := range with {
		if checkVal, ok := diff[key]; !ok {
			diff[key] = value
		} else if value.Matches(&checkVal) {
			delete(diff, key)
		}
	}
	return diff
}

func (s *KeyMixed) IsEmpty() bool {
	return len(s.values) == 0
}

func (s *KeyMixed) ReplaceExisting() bool {
	return s.replaceExisting
}

func (s *KeyMixed) applyValues(with map[string]Mixed) {
	s.values = with
}
func (s *KeyMixed) merge() {
	useVals := s.Values()
	s.Clear()
	s.replaceExisting = false
	s.values = useVals
}

func NewKeyMixed(with map[string]Mixed) KeyMixed {
	v := KeyMixed{}
	v.Clear()
	if with != nil {
		v.applyValues(with)
	}
	return v
}

func (s *KeyMixed) MarshalValue() (*proto.Value, error) {
	val := &proto.Value{}
	val.Array = proto.NewRepeatedValue()
	if err := s.mapToProto(s.values, val.Array.Mixed); err != nil {
		return nil, err
	}

	if len(s.toAdd) > 0 {
		val.ArrayAppend = proto.NewRepeatedValue()
		if err := s.mapToProto(s.toAdd, val.ArrayAppend.Mixed); err != nil {
			return nil, err
		}
	}

	if len(s.toRemove) > 0 {
		val.ArrayReduce = proto.NewRepeatedValue()
		for key := range s.toRemove {
			val.ArrayReduce.Mixed[key] = nil
		}
	}

	return val, nil
}

func (s *KeyMixed) mapToProto(with map[string]Mixed, onto map[string]*proto.Value) error {
	if with == nil {
		return nil
	}
	for key, value := range with {
		mVal, err := value.MarshalValue()
		if err != nil {
			return nil
		}
		onto[key] = mVal
	}
	return nil
}

func (s *KeyMixed) protoToMap(with map[string]*proto.Value, onto map[string]Mixed) error {
	if with == nil {
		return nil
	}
	for key, value := range with {
		mVal := Mixed{}
		err := mVal.UnmarshalValue(value)
		if err != nil {
			return nil
		}
		onto[key] = mVal
	}
	return nil
}

func (s *KeyMixed) UnmarshalValue(value *proto.Value) error {
	if value != nil {
		if value.Array != nil && value.Array.GetMixed() != nil {
			newVal := make(map[string]Mixed)
			if err := s.protoToMap(value.Array.GetMixed(), newVal); err != nil {
				return err
			}
			s.applyValues(newVal)
		}
		if value.ArrayAppend != nil && value.ArrayAppend.GetMixed() != nil {
			newVal := make(map[string]Mixed)
			if err := s.protoToMap(value.ArrayAppend.GetMixed(), newVal); err != nil {
				return err
			}
			for key, val := range newVal {
				s.Set(key, val)
			}
		}
		if value.ArrayReduce != nil && value.ArrayReduce.GetMixed() != nil {
			for key := range value.ArrayReduce.GetMixed() {
				s.Remove(key)
			}
		}
	}
	return nil
}

func (s *KeyMixed) IsZero() bool {
	return s == nil ||
		(len(s.values) == 0 &&
			len(s.toAdd) == 0 &&
			len(s.toRemove) == 0)
}

func (s *KeyMixed) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_KeyMixed}
}

func (s *KeyMixed) ObserveMutation(resp *proto.MutateResponse) {
	if resp.GetSuccess() {
		s.merge()
	}
}
