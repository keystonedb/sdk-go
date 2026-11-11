package keystone

import (
	"hash/fnv"
	"net"
	"regexp"
	"strings"
	"unicode/utf8"

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
	if name == "" {
		return PersonName{}
	}
	mask := BasicMask(name)
	return PersonName{NewSecureString(name, mask)}
}

func (s *PersonName) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_SecureText, ExtendedType: proto.Property_PersonName}
}

type Phone struct{ SecureString }

// keep last 3, mask 4 preceeding, keep rest
var phoneMask = regexp.MustCompile(`^(.*?)(\d{0,4})(\d{3})$`)

func NewPhone(phone string) Phone {
	if phone == "" {
		return Phone{}
	}
	mask := phoneMask.ReplaceAllString(phone, "${1}****${3}")
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
	var mask string
	if email == "" {
		return Email{}
	}
	split := strings.SplitN(email, "@", 2)
	mask = BasicMask(split[0])
	if len(split) > 1 {
		if strings.HasSuffix(split[1], ".me") {
			mask += "@" + BasicMask(strings.TrimSuffix(split[1], ".me")) + ".me"
		} else {
			mask += "@" + split[1]
		}
	}
	return Email{NewSecureString(email, mask)}
}

type SecureIP struct{ SecureString }

func (s *SecureIP) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_SecureText, ExtendedType: proto.Property_IP}
}

func NewSecureIPV4(ip string) SecureIP {
	if parsed := net.ParseIP(ip); parsed != nil {
		if maskedIP := parsed.Mask(net.IPv4Mask(0xff, 0xff, 0xff, 0)); maskedIP != nil {
			return SecureIP{NewSecureString(ip, maskedIP.String())}
		}
	}
	return SecureIP{}
}

func NewSecureIPV6(ip string) SecureIP {
	if parsed := net.ParseIP(ip); parsed != nil {
		mask := net.CIDRMask(48, 128)
		if maskedIP := parsed.Mask(mask); maskedIP != nil {
			return SecureIP{NewSecureString(ip, maskedIP.String())}
		}
	}
	return SecureIP{}
}

func NewSecureIP(ip string) SecureIP {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return SecureIP{}
	}
	if parsed.To4() != nil {
		return NewSecureIPV4(ip)
	}
	return NewSecureIPV6(ip)
}

// BasicMask masks each word, keeping its first and last letter. The number of asterisks
// (3-9) is derived from the word, so the same input always produces the same mask.
func BasicMask(unmasked string) string {
	if unmasked == "" {
		return ""
	}
	// split masked into words
	words := strings.Split(unmasked, " ")
	for i, word := range words {
		word = strings.TrimSpace(word)
		// if greater than 2 characters, put 3-7 asterisks between the first and last letter
		if l := utf8.RuneCountInString(word); l > 2 {
			first, _ := utf8.DecodeRuneInString(word)
			last, _ := utf8.DecodeLastRuneInString(word)
			words[i] = string(first) + strings.Repeat("*", maskLength(word)) + string(last)
		} else if l > 1 {
			// whole word becomes 3-7 asterisks
			first, _ := utf8.DecodeRuneInString(word)
			words[i] = string(first) + strings.Repeat("*", maskLength(word))
		} else {
			// single character, mask it
			words[i] = strings.Repeat("*", maskLength(word))
		}
	}
	return strings.Join(words, " ")
}

// maskLength returns an asterisk count of 3-9 derived from the word.
func maskLength(word string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(word))
	return int(h.Sum32()%7) + 3
}
