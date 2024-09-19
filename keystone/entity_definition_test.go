package keystone

import (
	"github.com/keystonedb/sdk-go/proto"
	"testing"
	"time"
)

type definedEntity struct {
}

type definedTimeSeriesEntity struct{}

func (d definedTimeSeriesEntity) GetTimeSeriesInputTime() time.Time { return time.Now() }

type definedTimeSeriesEntityPoint struct{}

func (d *definedTimeSeriesEntityPoint) GetTimeSeriesInputTime() time.Time { return time.Now() }

func (e *definedEntity) GetKeystoneDefinition() TypeDefinition {
	return TypeDefinition{
		Type:        "entity",
		Name:        "Entity",
		Description: "An entity",
		Singular:    "Entity",
		Plural:      "Entities",
	}
}

func Test_Define(t *testing.T) {
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
			res := Define(test.with)
			if res.Type != test.expectType {
				t.Errorf("Define().Type = %s; want %s", Define(test.with).Type, test.expectType)
			}
			if res.Name != test.expectName {
				t.Errorf("Define().Name = %s; want %s", Define(test.with).Name, test.expectName)
			}
			if res.Description != test.expectDescription {
				t.Errorf("Define().Description = %s; want %s", Define(test.with).Description, test.expectDescription)
			}
			if res.Singular != test.expectSingular {
				t.Errorf("Define().Singular = %s; want %s", Define(test.with).Singular, test.expectSingular)
			}
			if res.Plural != test.expectPlural {
				t.Errorf("Define().Plural = %s; want %s", Define(test.with).Plural, test.expectPlural)
			}
		})
	}
}
