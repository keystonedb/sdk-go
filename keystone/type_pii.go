package keystone

import (
	"math/rand/v2"
	"net"
	"strings"

	"github.com/keystonedb/sdk-go/proto"
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

type SecurePII struct{ SecureString }

func (s *SecurePII) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_SecureText, ExtendedType: proto.Property_Personal}
}

func NewSecurePII(info, masked string) SecurePII {
	return SecurePII{NewSecureString(info, masked)}
}

type PersonName struct{ SecureString }

func NewPersonName(name string) PersonName {
	mask := BasicMask(name)
	return PersonName{NewSecureString(name, mask)}
}

func (s *PersonName) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_SecureText, ExtendedType: proto.Property_PersonName}
}

type Phone struct{ SecureString }

func NewPhone(phone string) Phone {
	mask := phone // TODO: MASK
	return Phone{NewSecureString(phone, mask)}
}

func (s *Phone) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_SecureText, ExtendedType: proto.Property_Phone}
}

type Email struct{ SecureString }

func (s *Email) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_SecureText, ExtendedType: proto.Property_Email}
}

func NewEmail(email string) Email {
	mask := email // TODO: MASK
	return Email{NewSecureString(email, mask)}
}

type SecureIP struct{ SecureString }

func (s *SecureIP) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_SecureText, ExtendedType: proto.Property_IP}
}

func NewSecureIPV4(ip string) SecureIP {
	return SecureIP{NewSecureString(ip, net.ParseIP(ip).Mask(net.IPv4Mask(0xff, 0xff, 0xff, 0)).String())}
}

func BasicMask(unmasked string) string {
	// split masked into words
	words := strings.Split(unmasked, " ")
	for i, word := range words {
		word = strings.TrimSpace(word)
		// if greater than 2 characters, put 3-7 asterisks between the first and last letter
		if l := len(word); l > 2 {
			words[i] = word[:1] + strings.Repeat("*", rand.IntN(7)+3) + word[l-1:]
		} else if l > 1 {
			// whole word becomes 3-7 asterisks
			words[i] = word[:1] + strings.Repeat("*", rand.IntN(7)+3)
		}
	}
	return strings.Join(words, " ")
}
