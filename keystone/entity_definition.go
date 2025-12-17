package keystone

import (
	"errors"
	"reflect"
	"strings"

	"github.com/keystonedb/sdk-go/keystone/reflector"
	"github.com/keystonedb/sdk-go/proto"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
	isChild      bool

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

	_, definition.isChild = input.(ChildEntity)

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
	return MapPropertiesWithPrefix(v, "")
}

func MapPropertiesWithPrefix(v interface{}, prefix string) (map[Property]proto.PropertyDefinition, error) {

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

		currentProp, def := ReflectProperty(field, prefix)
		if currentProp.HydrateOnly() {
			continue
		}
		currentVal := val.FieldByIndex(field.Index)
		ref := GetReflector(field.Type, currentVal)
		if ref != nil {
			properties[currentProp] = mergeDefinitions(def, ref.PropertyDefinition())
		} else {
			var subProps map[Property]proto.PropertyDefinition
			var err error
			if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
				subProps, err = MapPropertiesWithPrefix(reflect.New(field.Type.Elem()).Interface(), currentProp.Name())
			} else if field.Type.Kind() == reflect.Struct {
				subProps, err = MapPropertiesWithPrefix(currentVal.Interface(), currentProp.Name())
			}
			if err != nil {
				return nil, err
			}
			if subProps != nil {
				for k, subV := range subProps {
					properties[k] = subV
				}
			}
		}
	}
	return properties, nil
}

func (t TypeDefinition) Schema() *proto.Schema {
	sch := &proto.Schema{
		Name:        t.Name,
		Description: t.Description,
		Type:        t.Type,
		Singular:    t.Singular,
		Plural:      t.Plural,
		Options:     t.Options,
		IsChild:     t.isChild,
		KsType:      t.KeystoneType,
	}

	if len(t.Properties) > 0 {
		for k, v := range t.Properties {
			if k.Name() == "" || k.Name()[:1] == "_" {
				// Skip over internal and empty properties
				continue
			}
			sch.Properties = append(sch.Properties, &proto.Property{
				Name:         k.Name(),
				DataType:     v.DataType,
				ExtendedType: v.ExtendedType,
				Options:      v.Options,
			})
		}
	}

	return sch
}

func (t TypeDefinition) HasOption(option proto.Schema_Option) bool {
	for _, o := range t.Options {
		if o == option {
			return true
		}
	}
	return false
}
