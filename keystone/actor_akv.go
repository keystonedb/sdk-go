package keystone

import (
	"context"
	"github.com/keystonedb/sdk-go/keystone/reflector"
	"github.com/keystonedb/sdk-go/proto"
	"reflect"
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
