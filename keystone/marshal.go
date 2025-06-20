package keystone

import (
	"errors"
	"github.com/keystonedb/sdk-go/keystone/reflector"
	"github.com/keystonedb/sdk-go/proto"
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

	val := reflector.Deref(reflect.ValueOf(v))

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

		currentProp, propDef := ReflectProperty(field, "")

		if currentProp.HydrateOnly() {
			// Skip properties that are only for hydration
			continue
		}

		currentVal := val.FieldByIndex(field.Index)
		ref := GetReflector(field.Type, currentVal)
		if ref != nil {
			if vRef, vRefOk := ref.(valueMarshalReflector); vRefOk && vRef.IsZero() {
				continue
			}
			protoVal, err := ref.ToProto(currentVal)
			protoVal.KnownType = mergeDefinitions(propDef, ref.PropertyDefinition()).DataType
			if err != nil {
				return nil, err
			} else {
				properties[currentProp] = protoVal
			}
		} else if !currentVal.IsZero() {
			var subProps map[Property]*proto.Value
			var err error

			if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
				subProps, err = Marshal(reflector.Deref(currentVal).Interface())
			} else if field.Type.Kind() == reflect.Struct {
				subProps, err = Marshal(currentVal.Interface())
			}

			if err != nil {
				return nil, err
			} else {
				prefix := currentProp.Name()
				for k, subV := range subProps {
					k.SetPrefix(prefix)
					properties[k] = subV
				}
			}
		}
	}
	return properties, nil
}

var CannotMarshalValueError = errors.New("cannot marshal Value")

func MarshalValue(v interface{}) (*proto.Value, error) {
	val := reflector.Deref(reflect.ValueOf(v))
	ref := GetReflector(val.Type(), val)
	if ref != nil {
		return ref.ToProto(val)
	}
	return nil, CannotMarshalValueError
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
