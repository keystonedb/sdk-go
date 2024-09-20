package keystone

import (
	"github.com/keystonedb/sdk-go/proto"
	"reflect"
)

// WhereEquals is a find option that filters entities by a property equaling a value
func WhereEquals(key string, value any) FindOption {
	return propertyFilter{key: key, values: valuesFromAny(value), operator: proto.Operator_Equal}
}

// WhereNotEquals is a find option that filters entities by a property not equaling a value
func WhereNotEquals(key string, value any) FindOption {
	return propertyFilter{key: key, values: valuesFromAny(value), operator: proto.Operator_NotEqual}
}

// WhereGreaterThan is a find option that filters entities by a property being greater than a value
func WhereGreaterThan(key string, value any) FindOption {
	return propertyFilter{key: key, values: valuesFromAny(value), operator: proto.Operator_GreaterThan}
}

// WhereGreaterThanOrEquals is a find option that filters entities by a property being greater than or equal to a value
func WhereGreaterThanOrEquals(key string, value any) FindOption {
	return propertyFilter{key: key, values: valuesFromAny(value), operator: proto.Operator_GreaterThanOrEqual}
}

// WhereLessThan is a find option that filters entities by a property being less than a value
func WhereLessThan(key string, value any) FindOption {
	return propertyFilter{key: key, values: valuesFromAny(value), operator: proto.Operator_LessThan}
}

// WhereLessThanOrEquals is a find option that filters entities by a property being less than or equal to a value
func WhereLessThanOrEquals(key string, value any) FindOption {
	return propertyFilter{key: key, values: valuesFromAny(value), operator: proto.Operator_LessThanOrEqual}
}

// WhereContains is a find option that filters entities by a property containing a value
func WhereContains(key string, value any) FindOption {
	return propertyFilter{key: key, values: valuesFromAny(value), operator: proto.Operator_Contains}
}

// WhereNotContains is a find option that filters entities by a property not containing a value
func WhereNotContains(key string, value any) FindOption {
	return propertyFilter{key: key, values: valuesFromAny(value), operator: proto.Operator_NotContains}
}

// WhereStartsWith is a find option that filters entities by a property starting with a value
func WhereStartsWith(key string, value any) FindOption {
	return propertyFilter{key: key, values: valuesFromAny(value), operator: proto.Operator_StartsWith}
}

// WhereEndsWith is a find option that filters entities by a property ending with a value
func WhereEndsWith(key string, value string) FindOption {
	return propertyFilter{key: key, values: valuesFromAny(value), operator: proto.Operator_EndsWith}
}

// WhereIn is a find option that filters entities by a property being in a list of values
func WhereIn(key string, value ...any) FindOption {
	return propertyFilter{key: key, values: valuesFromAny(value...), operator: proto.Operator_In}
}

// WhereNotIn is a find option that filters entities by a property not being in a list of values
func WhereNotIn(key string, value ...any) FindOption {
	return propertyFilter{key: key, values: valuesFromAny(value...), operator: proto.Operator_Equal}
}

// WhereBetween is a find option that filters entities by a property being between two values
func WhereBetween(key string, value1, value2 any) FindOption {
	return propertyFilter{key: key, values: valuesFromAny(value1, value2), operator: proto.Operator_Equal}
}

type propertyFilter struct {
	key      string
	values   []*proto.Value
	operator proto.Operator
}

func (f propertyFilter) Apply(config *filterRequest) {
	if config.Filters == nil {
		config.Filters = make([]*proto.PropertyFilter, 0)
	}

	config.Filters = append(config.Filters, &proto.PropertyFilter{
		Property: f.key,
		Operator: f.operator,
		Values:   f.values,
	})
}

func valueFromAny(value any) *proto.Value {
	v := reflect.ValueOf(value)
	ref := GetReflector(v.Type(), v)
	if ref != nil {
		if val, err := ref.ToProto(v); err == nil {
			return val
		}
	}
	return nil
}

func valuesFromAny(values ...any) []*proto.Value {
	var result []*proto.Value
	for _, v := range values {
		result = append(result, valueFromAny(v))
	}
	return result
}
