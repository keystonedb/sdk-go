package keystone

import (
	"context"
	"github.com/keystonedb/sdk-go/proto"
	"reflect"
)

func (a *Actor) SetDynamicProperties(ctx context.Context, entityID string, setProperties []*proto.EntityProperty, removeProperties []string, comment string) error {
	mutation := &proto.Mutation{
		DynamicProperties:       setProperties,
		RemoveDynamicProperties: removeProperties,
		Comment:                 comment,
	}
	mutation.Mutator = a.user

	m := &proto.MutateRequest{
		Authorization: a.Authorization(),
		EntityId:      entityID,
		Mutation:      mutation,
	}

	return mutateToError(a.connection.Mutate(ctx, m))
}

func (a *Actor) GetDynamicProperties(ctx context.Context, entityID string, properties ...string) (PropertyValueList, error) {
	m := &proto.EntityRequest{
		Authorization: a.Authorization(),
		EntityId:      entityID,
		View: &proto.EntityView{
			DynamicProperties: properties,
		},
	}

	resp, err := a.connection.Retrieve(ctx, m)
	if err != nil {
		return nil, err
	}

	res := make(PropertyValueList)
	for _, prop := range resp.GetDynamicProperties() {
		res[prop.Property] = prop.GetValue()
	}

	return res, nil
}

type PropertyValueList map[string]*proto.Value

func (p PropertyValueList) Get(key string) *proto.Value {
	return p[key]
}

func (p PropertyValueList) GetText(key, defaultValue string) string {
	if v, ok := p[key]; ok {
		return v.GetText()
	}
	return defaultValue
}

func (p PropertyValueList) GetInt(key string, defaultValue int64) int64 {
	if v, ok := p[key]; ok {
		return v.GetInt()
	}
	return defaultValue
}

func (p PropertyValueList) GetFloat(key string, defaultValue float64) float64 {
	if v, ok := p[key]; ok {
		return v.GetFloat()
	}
	return defaultValue
}

func (p PropertyValueList) GetBool(key string, defaultValue bool) bool {
	if v, ok := p[key]; ok {
		return v.GetBool()
	}
	return defaultValue
}

func NewDynamicProperties(props map[string]interface{}) []*proto.EntityProperty {
	properties := make([]*proto.EntityProperty, 0, len(props))
	for key, value := range props {
		prop := NewDynamicProperty(key, value)
		if prop != nil {
			properties = append(properties, prop)
		}
	}
	return properties
}

func NewDynamicProperty(key string, value interface{}) *proto.EntityProperty {
	v := reflect.ValueOf(value)
	ref := GetReflector(v.Type(), v)
	if ref != nil {
		val, err := ref.ToProto(v)
		if err != nil {
			return nil
		}
		return &proto.EntityProperty{
			Property: key,
			Value:    val,
		}
	}
	return nil
}