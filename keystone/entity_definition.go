package keystone

import (
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
	Type         string // Unique Type Name e.g. user
	Name         string // Friendly name of the entity e.g. Library User
	Description  string // Description of the entity
	Singular     string // Name for a single one of these entities e.g. User
	Plural       string // Name for a collection of these entities e.g. Users
	Options      []proto.Schema_Option
	KeystoneType proto.Schema_Type
}

func Define(input interface{}) TypeDefinition {
	definition := TypeDefinition{}

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
