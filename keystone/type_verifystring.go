package keystone

import (
	"github.com/keystonedb/sdk-go/proto"
)

// VerifyString is a string that can be verified
type VerifyString struct {
	Original string `json:"original,omitempty"`
	verified bool
}

// String returns the original string if it exists, otherwise the masked string
func (e *VerifyString) String() string {
	return e.Original
}

// NewVerifyString creates a new VerifyString
func NewVerifyString(original string) VerifyString {
	return VerifyString{Original: original}
}

func (e *VerifyString) MarshalValue() (*proto.Value, error) {
	return &proto.Value{
		SecureText: e.Original,
	}, nil
}

func (e *VerifyString) UnmarshalValue(value *proto.Value) error {
	if value != nil {
		e.Original = value.GetSecureText()
		e.verified = value.GetBool()
	}
	return nil
}

func (e *VerifyString) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_VerifyText}
}
