package keystone

import (
	"context"
	"errors"
	"reflect"

	"github.com/keystonedb/sdk-go/keystone/reflector"
	"github.com/keystonedb/sdk-go/proto"
)

type AKVProperty struct {
	Property *proto.Property
	Value    *proto.Value
}

func (p *AKVProperty) toProto() *proto.AKVProperty {
	return &proto.AKVProperty{
		Property: p.Property,
		Value:    p.Value,
	}
}

func AKVRaw(property string, value *proto.Value) AKVProperty {
	prop := AKVProperty{
		Property: &proto.Property{
			Name:     property,
			DataType: proto.Property_Unmanaged,
		},
		Value: value,
	}
	return prop
}

func AKV(property string, value any) AKVProperty {
	prop := AKVProperty{
		Property: &proto.Property{
			Name: property,
		},
	}

	val := reflector.Deref(reflect.ValueOf(value))
	ref := GetReflector(val.Type(), val)
	if ref == nil {
		return prop
	}

	prop.Property.DataType = ref.PropertyDefinition().DataType
	prop.Property.ExtendedType = ref.PropertyDefinition().ExtendedType
	prop.Property.Options = ref.PropertyDefinition().Options

	pVal, err := ref.ToProto(val)
	if err != nil {
		return prop
	}
	prop.Value = pVal

	return prop
}

func (a *Actor) AKVPut(ctx context.Context, properties ...AKVProperty) (*proto.GenericResponse, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}

	putRequest := &proto.AKVPutRequest{
		Authorization: a.Authorization(),
	}

	for _, prop := range properties {
		putRequest.Properties = append(putRequest.Properties, prop.toProto())
	}

	resp, err := a.connection.AKVPut(ctx, putRequest)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (a *Actor) AKVGet(ctx context.Context, properties ...string) (map[string]*proto.Value, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}

	getRequest := &proto.AKVGetRequest{
		Authorization: a.Authorization(),
		Properties:    properties,
	}
	resp, err := a.connection.AKVGet(ctx, getRequest)
	if err != nil {
		return nil, err
	}

	return resp.GetProperties(), nil
}

// AKVDel deletes entries from the application key-value store
func (a *Actor) AKVDel(ctx context.Context, properties ...string) (*proto.GenericResponse, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}

	delRequest := &proto.AKVDelRequest{
		Authorization: a.Authorization(),
		Properties:    properties,
	}

	resp, err := a.connection.AKVDel(ctx, delRequest)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
