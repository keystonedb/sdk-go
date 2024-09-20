package keystone

import (
	"errors"
	"github.com/keystonedb/sdk-go/keystone/reflector"
	"github.com/keystonedb/sdk-go/proto"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"reflect"
	"strings"
)

// EntityDefinition is an interface that defines the keystone entity
type EntityDefinition interface {
	GetKeystoneDefinition() TypeDefinition
}

// TypeDefinition is a definition of a keystone type
type TypeDefinition struct {
	id           string // Generated ID from keystone server
	Type         string // Unique Type Name e.g. user
	Name         string // Friendly name of the entity e.g. Library User
	Description  string // Description of the entity
	Singular     string // Name for a single one of these entities e.g. User
	Plural       string // Name for a collection of these entities e.g. Users
	Options      []proto.Schema_Option
	KeystoneType proto.Schema_Type

	Properties map[Property]proto.PropertyDefinition
}

func NewTypeDefinition() TypeDefinition {
	return TypeDefinition{
		Properties: map[Property]proto.PropertyDefinition{},
	}
}

func QuickDefine(input interface{}) TypeDefinition {
	definition := NewTypeDefinition()

	if definer, ok := input.(EntityDefinition); ok {
		definition = definer.GetKeystoneDefinition()
	} else if t := reflect.ValueOf(input).Type(); t.Kind() == reflect.Struct {
		if d, o := reflect.New(t).Interface().(EntityDefinition); o {
			definition = d.GetKeystoneDefinition()
		}
	}

	if definition.Type == "" {
		definition.Type = Type(input)
	}

	if definition.Name == "" {
		definition.Name = cases.Title(language.English).String(strings.ReplaceAll(definition.Type, "-", " "))
	}

	if _, ok := input.(TSEntity); ok {
		definition.KeystoneType = proto.Schema_TimeSeries
	}

	return definition
}

func Define(input interface{}) TypeDefinition {
	definition := QuickDefine(input)
	if props, err := MapProperties(input); err == nil {
		definition.Properties = props
	}
	return definition
}

var CannotMapPrimitives = errors.New("cannot map primitive type")
var CannotMapNil = errors.New("cannot map nil")

func MapProperties(v interface{}) (map[Property]proto.PropertyDefinition, error) {

	if v == nil {
		return nil, CannotMapNil
	}
	val := reflector.Deref(reflect.ValueOf(v))
	if val.Kind() != reflect.Struct {
		return nil, CannotMapPrimitives
	}

	properties := make(map[Property]proto.PropertyDefinition)

	for _, field := range reflect.VisibleFields(val.Type()) {
		if field.Anonymous || !field.IsExported() {
			continue
		}

		currentProp := NewProperty(field.Name)
		currentVal := val.FieldByIndex(field.Index)
		ref := GetReflector(field.Type, currentVal)
		if ref != nil {
			properties[currentProp] = ref.PropertyDefinition()
		} else if field.Type.Kind() == reflect.Struct {
			subProps, err := MapProperties(currentVal.Interface())
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

func (t TypeDefinition) Schema() *proto.Schema {
	return nil
}
