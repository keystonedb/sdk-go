package keystone

import (
	"errors"
	"github.com/keystonedb/sdk-go/proto"
	"strings"
)

func NewExternalID(vendorID, appID, entityType string, id ID) *ExternalID {
	return &ExternalID{
		vendorID:   vendorID,
		appID:      appID,
		entityType: entityType,
		id:         id,
	}
}

type ExternalID struct {
	vendorID   string
	appID      string
	entityType string
	id         ID
}

func (f *ExternalID) MarshalValue() (*proto.Value, error) {
	return &proto.Value{
		Text: f.vendorID + "/" + f.appID + "/" + f.entityType + "/" + f.id.String(),
	}, nil
}

func (f *ExternalID) UnmarshalValue(value *proto.Value) error {
	if value == nil {
		return nil
	}
	if value.Text == "" {
		return nil
	}
	parts := strings.Split(value.Text, "/")
	switch len(parts) {
	case 1:
		f.id = ID(parts[0])
	case 2:
		f.entityType = parts[0]
		f.id = ID(parts[1])
	case 3:
		f.appID = parts[0]
		f.entityType = parts[1]
		f.id = ID(parts[2])
	case 4:
		f.vendorID = parts[0]
		f.appID = parts[1]
		f.entityType = parts[2]
		f.id = ID(parts[3])
	default:
		return errors.New("external id format error")
	}
	return nil
}

func (f *ExternalID) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Text, ExtendedType: proto.Property_ExternalID}
}

func (f *ExternalID) IsZero() bool {
	return f == nil || f.id.ParentID() == ""
}

func (f *ExternalID) ID() *ID {
	if f == nil {
		return nil
	}
	return &f.id
}

func (f *ExternalID) Source() *Key {
	return NewKey(f.vendorID, f.appID, f.entityType)
}
