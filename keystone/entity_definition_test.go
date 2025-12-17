package keystone

import (
	"errors"
	"testing"
	"time"

	"github.com/keystonedb/sdk-go/proto"
)

type definedEntity struct {
	Name string
}

type parentDefinedEntity struct {
	Child    definedEntity
	HasSaved bool
	Numeric  int
}

type definedTimeSeriesEntity struct{}

func (d definedTimeSeriesEntity) GetTimeSeriesInputTime() time.Time { return time.Now() }

type definedTimeSeriesEntityPoint struct{}

func (d *definedTimeSeriesEntityPoint) GetTimeSeriesInputTime() time.Time { return time.Now() }

func (e *definedEntity) GetKeystoneDefinition() TypeDefinition {
	d := NewTypeDefinition()
	d.Type = "entity"
	d.Name = "Entity"
	d.Description = "An entity"
	d.Singular = "Entity"
	d.Plural = "Entities"
	return d
}

func Test_QuickDefine(t *testing.T) {
	tests := []struct {
		name              string
		with              interface{}
		expectType        string
		expectName        string
		expectDescription string
		expectSingular    string
		expectPlural      string
		expectedType      proto.Schema_Type
	}{
		{"MarshaledEntity", MarshaledEntity{}, "marshaled-entity", "Marshaled Entity", "", "", "", proto.Schema_Entity},
		{"MarshaledEntity pointer", &MarshaledEntity{}, "marshaled-entity", "Marshaled Entity", "", "", "", proto.Schema_Entity},
		{"definedEntity pointer", &definedEntity{}, "entity", "Entity", "An entity", "Entity", "Entities", proto.Schema_Entity},
		{"definedEntity", definedEntity{}, "entity", "Entity", "An entity", "Entity", "Entities", proto.Schema_Entity},
		{"definedTimeSeriesEntity", definedTimeSeriesEntity{}, "defined-time-series-entity", "Defined Time Series Entity", "", "", "", proto.Schema_TimeSeries},
		{"definedTimeSeriesEntity", &definedTimeSeriesEntity{}, "defined-time-series-entity", "Defined Time Series Entity", "", "", "", proto.Schema_TimeSeries},
		{"definedTimeSeriesEntityPoint", definedTimeSeriesEntityPoint{}, "defined-time-series-entity-point", "Defined Time Series Entity Point", "", "", "", proto.Schema_TimeSeries},
		{"definedTimeSeriesEntityPoint", &definedTimeSeriesEntityPoint{}, "defined-time-series-entity-point", "Defined Time Series Entity Point", "", "", "", proto.Schema_TimeSeries},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := QuickDefine(test.with)
			if res.Type != test.expectType {
				t.Errorf("QuickDefine().Type = %s; want %s", res.Type, test.expectType)
			}
			if res.Name != test.expectName {
				t.Errorf("QuickDefine().Name = %s; want %s", res.Name, test.expectName)
			}
			if res.Description != test.expectDescription {
				t.Errorf("QuickDefine().Description = %s; want %s", res.Description, test.expectDescription)
			}
			if res.Singular != test.expectSingular {
				t.Errorf("QuickDefine().Singular = %s; want %s", res.Singular, test.expectSingular)
			}
			if res.Plural != test.expectPlural {
				t.Errorf("QuickDefine().Plural = %s; want %s", res.Plural, test.expectPlural)
			}
		})
	}
}
func Test_MapProperties_Unsupported(t *testing.T) {
	tests := []struct {
		name      string
		with      interface{}
		expectErr error
	}{
		{"nil", nil, CannotMapNil},
		{"string", "string", CannotMapPrimitives},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := MapProperties(test.with)
			if !errors.Is(err, test.expectErr) {
				t.Errorf("MapProperties() = %v; want %v", err, test)
			}
		})
	}
}

func Test_MapProperties_Nested(t *testing.T) {

	props, err := MapProperties(parentDefinedEntity{})
	if err != nil {
		t.Errorf("MapProperties() = %v; want nil", err)
	}
	if len(props) != 3 {
		t.Errorf("MapProperties() = %d; want 3", len(props))
	}

	child, hasChild := props[NewPrefixProperty("child", "name")]
	if !hasChild {
		t.Errorf("MapProperties()[child.name] was not returned")
	} else if child.DataType != proto.Property_Text {
		t.Errorf("MapProperties()[child.name] type was %s, want %s", child.DataType, proto.Property_Text)
	}

	numeric, hasNumeric := props[NewProperty("numeric")]
	if !hasNumeric {
		t.Errorf("MapProperties()[numeric] was not returned")
	} else if numeric.DataType != proto.Property_Number {
		t.Errorf("MapProperties()[numeric] type was %s, want %s", numeric.DataType, proto.Property_Number)
	}

	saved, hasHasSaved := props[NewProperty("has_saved")]
	if !hasHasSaved {
		t.Errorf("MapProperties()[has_saved] was not returned")
	} else if saved.DataType != proto.Property_Boolean {
		t.Errorf("MapProperties()[has_saved] type was %s, want %s", saved.DataType, proto.Property_Boolean)
	}
}

func Test_MapProperties_viaDefined(t *testing.T) {
	def := Define(definedEntity{})
	props := def.Properties

	if len(props) != 1 {
		t.Errorf("MapProperties() = %d; want 1", len(props))
	}

	name, hasName := props[NewProperty("name")]
	if !hasName {
		t.Errorf("MapProperties()[name] was not returned")
	} else if name.DataType != proto.Property_Text {
		t.Errorf("MapProperties()[name] type was %s, want %s", name.DataType, proto.Property_Text)
	}
}
