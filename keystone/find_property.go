package keystone

import (
	"github.com/keystonedb/sdk-go/proto"
	"reflect"
)

// Where allows for filtering entities by a Property, with a string operator
func Where(key, operator string, values ...any) FindOption {
	if len(values) == 0 {
		return nil
	}

	value := values[0]
	switch operator {
	case "eq", "=":
		return WhereEquals(key, value)
	case "neq", "!=":
		return WhereNotEquals(key, value)
	case "gt", ">":
		return WhereGreaterThan(key, value)
	case "gte", ">=":
		return WhereGreaterThanOrEquals(key, value)
	case "lt", "<":
		return WhereLessThan(key, value)
	case "lte", "<=":
		return WhereLessThanOrEquals(key, value)
	case "contains", "c":
		return WhereContains(key, value)
	case "notcontains", "nc":
		return WhereNotContains(key, value)
	case "startswith", "sw":
		return WhereStartsWith(key, value)
	case "endswith", "ew":
		return WhereEndsWith(key, value)
	case "in":
		return WhereIn(key, values...)
	case "notin":
		return WhereNotIn(key, values...)
	case "between", "btw", "><":
		if len(values) < 2 {
			return nil
		}
		return WhereBetween(key, values[0], values[1])
	}
	return nil
}

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
func WhereEndsWith(key string, value any) FindOption {
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

func Or(filters ...FindOption) FindOption {
	return propertyFilter{or: true, nested: filters}
}

func And(filters ...FindOption) FindOption {
	return propertyFilter{nested: filters}
}

type propertyFilter struct {
	key      string
	values   []*proto.Value
	operator proto.Operator
	or       bool
	nested   []FindOption
}

func (f propertyFilter) Apply(config *filterRequest) {
	if config.Filters == nil {
		config.Filters = make([]*proto.PropertyFilter, 0)
	}
	config.Filters = append(config.Filters, f.toProto())
}

func (f propertyFilter) toProto() *proto.PropertyFilter {
	ret := &proto.PropertyFilter{
		Property: f.key,
		Operator: f.operator,
		Values:   f.values,
		Or:       f.or,
	}

	if f.nested != nil {
		for _, filter := range f.nested {
			if fil, ok := filter.(propertyFilter); ok {
				ret.Nested = append(ret.Nested, fil.toProto())
			}
		}
	}

	return ret
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
