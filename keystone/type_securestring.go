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

// MarshalJSON encodes the masked value as a JSON string; the original is never encoded.
func (e SecureString) MarshalJSON() ([]byte, error) { return json.Marshal(e.Masked) }

// UnmarshalJSON reads the masked value from a JSON string, or from the {"masked": ...}
// object written by older SDK versions.
func (e *SecureString) UnmarshalJSON(data []byte) error {
	var masked string
	if err := json.Unmarshal(data, &masked); err != nil {
		var obj struct {
			Masked string `json:"masked"`
		}
		if objErr := json.Unmarshal(data, &obj); objErr != nil {
			return err
		}
		masked = obj.Masked
	}
	e.Masked = masked
	e.Original = ""
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

// IsZero reports whether there is no original value to send.
func (e *SecureString) IsZero() bool {
	return e == nil || e.Original == ""
}

// IsCleared reports whether both the original and masked values are empty, which marshals
// as an empty secure value so the stored value is removed.
func (e *SecureString) IsCleared() bool {
	return e == nil || (e.Original == "" && e.Masked == "")
}
