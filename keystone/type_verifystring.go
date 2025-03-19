package keystone

import (
	"github.com/keystonedb/sdk-go/proto"
)

// VerifyString is a string that can be verified
type VerifyString struct {
	Original string `json:"original,omitempty"`
	verified *bool
}

// String returns the original string if it exists, otherwise the masked string
func (e *VerifyString) String() string {
	return e.Original
}

// Verified returns true if the string is confirmed to be the same as remote
func (e *VerifyString) Verified() bool { return e.verified != nil && *e.verified == true }

// WasChecked returns true if the string was checked
func (e *VerifyString) WasChecked() bool { return e.verified != nil }

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
		verified := value.GetBool()
		e.verified = &verified
	}
	return nil
}

func (e *VerifyString) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_VerifyText}
}

func (e *VerifyString) IsZero() bool {
	return e == nil || e.Original == ""
}
