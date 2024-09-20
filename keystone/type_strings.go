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

func (s *String) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Text}
}

type IPAddress String

func (s *IPAddress) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Text, ExtendedType: proto.Property_IP}
}

type UserInput String

func (s *UserInput) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Text, ExtendedType: proto.Property_UserInput}
}

type PII String

func (s *PII) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Text, ExtendedType: proto.Property_Personal}
}

type PersonName String

func (s *PersonName) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Text, ExtendedType: proto.Property_PersonName}
}

type Email String

func (s *Email) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Text, ExtendedType: proto.Property_Email}
}

type Phone String

func (s *Phone) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Text, ExtendedType: proto.Property_Phone}
}

type Country String

func (s *Country) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Text, ExtendedType: proto.Property_Country}
}

type URL String

func (s *URL) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Text, ExtendedType: proto.Property_URL}
}
