package keystone

import (
	"encoding/json"
	"github.com/keystonedb/sdk-go/proto"
)

// SecureString is a string that represents sensitive Data
type SecureString struct {
	Masked   string `json:"masked,omitempty"`
	Original string `json:"-"`
}

func (e *SecureString) MarshalJSON() ([]byte, error) { return json.Marshal(e.Masked) }

func (e *SecureString) UnmarshalJSON(data []byte) error {
	e.Masked = string(data)
	return nil
}

// String returns the original string if it exists, otherwise the masked string
func (e *SecureString) String() string {
	if e.Original != "" {
		return e.Original
	}
	return e.Masked
}

// NewSecureString creates a new SecureString
func NewSecureString(original, masked string) SecureString {
	return SecureString{
		Masked:   masked,
		Original: original,
	}
}

func (e *SecureString) MarshalValue() (*proto.Value, error) {
	return &proto.Value{
		Text:       e.Masked,
		SecureText: e.Original,
	}, nil
}

func (e *SecureString) UnmarshalValue(value *proto.Value) error {
	if value != nil {
		e.Original = value.GetSecureText()
		e.Masked = value.GetText()
	}
	return nil
}

func (e *SecureString) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_SecureText}
}

func (e *SecureString) IsZero() bool {
	return e == nil || e.Original == ""
}
