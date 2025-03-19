package keystone

import (
	"github.com/keystonedb/sdk-go/proto"
	"net"
)

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

type URL String

func (s *URL) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Text, ExtendedType: proto.Property_URL}
}

type Country String

func (s *Country) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Text, ExtendedType: proto.Property_Country}
}

type SecurePII SecureString

func (s *SecurePII) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_SecureText, ExtendedType: proto.Property_Personal}
}

func NewSecurePII(info, masked string) SecurePII {
	return SecurePII(NewSecureString(info, masked))
}

type PersonName SecureString

func NewPersonName(name string) PersonName {
	mask := name // TODO: MASK
	return PersonName(NewSecureString(name, mask))
}

func (s *PersonName) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_SecureText, ExtendedType: proto.Property_PersonName}
}

type Phone SecureString

func NewPhone(phone string) Phone {
	mask := phone // TODO: MASK
	return Phone(NewSecureString(phone, mask))
}

func (s *Phone) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_SecureText, ExtendedType: proto.Property_Phone}
}

type Email SecureString

func (s *Email) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_SecureText, ExtendedType: proto.Property_Email}
}
func NewEmail(email string) Email {
	mask := email // TODO: MASK
	return Email(NewSecureString(email, mask))
}

type SecureIP SecureString

func (s *SecureIP) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_SecureText, ExtendedType: proto.Property_IP}
}

func NewSecureIPV4(ip string) SecureIP {
	return SecureIP(NewSecureString(ip, net.ParseIP(ip).Mask(net.IPv4Mask(0xff, 0xff, 0xff, 0)).String()))
}
