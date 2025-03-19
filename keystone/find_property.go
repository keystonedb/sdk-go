package keystone

import (
	"github.com/keystonedb/sdk-go/proto"
	"reflect"
)

// WhereEquals is a find option that filters entities by a Property equaling a Value
func WhereEquals(key string, value any) FindOption {
	return propertyFilter{key: key, values: valuesFromAny(value), operator: proto.Operator_Equal}
}

// WhereNotEquals is a find option that filters entities by a Property not equaling a Value
func WhereNotEquals(key string, value any) FindOption {
	return propertyFilter{key: key, values: valuesFromAny(value), operator: proto.Operator_NotEqual}
}

// WhereGreaterThan is a find option that filters entities by a Property being greater than a Value
func WhereGreaterThan(key string, value any) FindOption {
	return propertyFilter{key: key, values: valuesFromAny(value), operator: proto.Operator_GreaterThan}
}

// WhereGreaterThanOrEquals is a find option that filters entities by a Property being greater than or equal to a Value
func WhereGreaterThanOrEquals(key string, value any) FindOption {
	return propertyFilter{key: key, values: valuesFromAny(value), operator: proto.Operator_GreaterThanOrEqual}
}

// WhereLessThan is a find option that filters entities by a Property being less than a Value
func WhereLessThan(key string, value any) FindOption {
	return propertyFilter{key: key, values: valuesFromAny(value), operator: proto.Operator_LessThan}
}

// WhereLessThanOrEquals is a find option that filters entities by a Property being less than or equal to a Value
func WhereLessThanOrEquals(key string, value any) FindOption {
	return propertyFilter{key: key, values: valuesFromAny(value), operator: proto.Operator_LessThanOrEqual}
}

// WhereContains is a find option that filters entities by a Property containing a Value
func WhereContains(key string, value any) FindOption {
	return propertyFilter{key: key, values: valuesFromAny(value), operator: proto.Operator_Contains}
}

// WhereNotContains is a find option that filters entities by a Property not containing a Value
func WhereNotContains(key string, value any) FindOption {
	return propertyFilter{key: key, values: valuesFromAny(value), operator: proto.Operator_NotContains}
}

// WhereStartsWith is a find option that filters entities by a Property starting with a Value
func WhereStartsWith(key string, value any) FindOption {
	return propertyFilter{key: key, values: valuesFromAny(value), operator: proto.Operator_StartsWith}
}

// WhereEndsWith is a find option that filters entities by a Property ending with a Value
func WhereEndsWith(key string, value string) FindOption {
	return propertyFilter{key: key, values: valuesFromAny(value), operator: proto.Operator_EndsWith}
}

// WhereIn is a find option that filters entities by a Property being in a list of values
func WhereIn(key string, value ...any) FindOption {
	return propertyFilter{key: key, values: valuesFromAny(value...), operator: proto.Operator_In}
}

// WhereNotIn is a find option that filters entities by a Property not being in a list of values
func WhereNotIn(key string, value ...any) FindOption {
	return propertyFilter{key: key, values: valuesFromAny(value...), operator: proto.Operator_NotEqual}
}

// WhereBetween is a find option that filters entities by a Property being between two values
func WhereBetween(key string, value1, value2 any) FindOption {
	return propertyFilter{key: key, values: valuesFromAny(value1, value2), operator: proto.Operator_Between}
}

// IsNull is a find option that filters entities by a Property being null
func IsNull(key string) FindOption {
	return propertyFilter{key: key, operator: proto.Operator_IsNull}
}

// IsNotNull is a find option that filters entities by a Property not being null
func IsNotNull(key string) FindOption {
	return propertyFilter{key: key, operator: proto.Operator_IsNotNull}
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

func valueFromString(value string) *proto.Value {
	return &proto.Value{Text: value}
}

func valueFromInt(value int64) *proto.Value {
	return &proto.Value{Int: value}
}

func valueFromAny(value any) *proto.Value {
	if value == nil {
		return nil
	}
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
