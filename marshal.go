package keystone

import (
	"errors"
	"github.com/keystonedb/sdk-go/sdk-go/proto"
	"log"
	"reflect"
)

var CannotMarshalPrimitives = errors.New("cannot marshal primitive type")
var CannotMarshalNil = errors.New("cannot marshal nil")

func Marshal(v interface{}) (map[Property]*proto.Value, error) {

	if v == nil {
		return nil, CannotMarshalNil
	}

	if m, ok := v.(Marshaler); ok {
		return m.MarshalKeystone()
	}

	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	if reflect.PointerTo(val.Type()).Implements(marshalerType) {
		// Support structs with Marshal pointer receiver
		vp := reflect.New(val.Type())
		vp.Elem().Set(val)
		x := vp.Interface()

		if m, ok := x.(Marshaler); ok {
			return m.MarshalKeystone()
		}
	}

	if val.Kind() != reflect.Struct {
		return nil, CannotMarshalPrimitives
	}

	properties := make(map[Property]*proto.Value)

	for _, field := range reflect.VisibleFields(val.Type()) {
		if field.Anonymous || !field.IsExported() {
			continue
		}

		currentProp := NewProperty(field.Name)
		enc := newTypeEncoder(field.Type)
		if enc == nil {
			if field.Type.Kind() == reflect.Struct {
				subStruct, err := Marshal(val.FieldByIndex(field.Index).Interface())
				if err != nil {
					return nil, err
				} else {
					prefix := currentProp.Name()
					for k, subV := range subStruct {
						k.SetPrefix(prefix)
						properties[k] = subV
					}
				}
				continue
			}

			log.Println("Unsupported type", field.Type, field.Name)
			// Skip unsupported types
			continue
		}

		protoVal, err := enc(val.FieldByIndex(field.Index))
		if err != nil {
			return nil, err
		}

		properties[currentProp] = protoVal
	}
	return properties, nil
}

var CannotMarshalValueError = errors.New("cannot marshal value")

func MarshalValue(v interface{}) (*proto.Value, error) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	enc := newTypeEncoder(val.Type())
	if enc == nil {
		return nil, CannotMarshalValueError
	}

	return enc(val)
}

func NewMarshaledEntity() MarshaledEntity {
	return MarshaledEntity{
		Properties: make(map[Property]*proto.Value),
	}
}

type MarshaledEntity struct {
	Properties map[Property]*proto.Value
}

func (m *MarshaledEntity) Append(name string, val interface{}) error {
	prop, err := MarshalValue(val)
	if err != nil {
		return err
	}

	m.Properties[NewProperty(name)] = prop
	return nil
}
