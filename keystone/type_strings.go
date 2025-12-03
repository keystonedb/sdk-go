package keystone

import "github.com/keystonedb/sdk-go/proto"

type String string

func (s *String) MarshalValue() (*proto.Value, error) {
	return &proto.Value{
		Text: string(*s),
	}, nil
}

func (s *String) UnmarshalValue(value *proto.Value) error {
	if value != nil {
		str := String(value.GetText())
		*s = str
	}
	return nil
}

func (s *String) IsZero() bool {
	return s == nil || *s == ""
}

func (s *String) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Text}
}

func (s *String) String() string {
	if s == nil {
		return ""
	}
	return string(*s)
}
