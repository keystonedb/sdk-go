package keystone

import "github.com/keystonedb/sdk-go/proto"

// GroupBy adds group keys. Aggregate queries return Rows instead of Entities.
func GroupBy(properties ...string) QueryOption {
	properties = append([]string(nil), properties...)
	return queryOption(func(req *proto.QueryRequest) { req.GroupBy = appendQueryProperties(req.GroupBy, properties) })
}

// Aggregate adds an aggregate under alias. An empty property means COUNT(*).
// Set distinct to true to aggregate distinct property values.
// The server validates property types, aliases and supported combinations.
func Aggregate(function proto.QueryAggregate_Function, property, alias string, distinct ...bool) QueryOption {
	d := len(distinct) > 0 && distinct[0]
	return queryOption(func(req *proto.QueryRequest) {
		req.Aggregates = append(req.Aggregates, &proto.QueryAggregate{Function: function, Property: property, As: alias, Distinct: d})
	})
}

// Count counts all matching entities in each group.
func Count(alias string) QueryOption { return Aggregate(proto.QueryAggregate_COUNT, "", alias) }

// CountProperty counts non-null values of property.
func CountProperty(property, alias string) QueryOption {
	return Aggregate(proto.QueryAggregate_COUNT, property, alias)
}

// CountDistinct counts distinct non-null property values.
func CountDistinct(property, alias string) QueryOption {
	return Aggregate(proto.QueryAggregate_COUNT, property, alias, true)
}

// Sum sums numeric property values.
func Sum(property, alias string) QueryOption {
	return Aggregate(proto.QueryAggregate_SUM, property, alias)
}

// Min returns the minimum property value.
func Min(property, alias string) QueryOption {
	return Aggregate(proto.QueryAggregate_MIN, property, alias)
}

// Max returns the maximum property value.
func Max(property, alias string) QueryOption {
	return Aggregate(proto.QueryAggregate_MAX, property, alias)
}

// Avg averages numeric property values.
func Avg(property, alias string) QueryOption {
	return Aggregate(proto.QueryAggregate_AVG, property, alias)
}

// Having filters aggregate aliases or group keys using Where, And, Or and
// null checks. Only property filters are accepted; invalid options fail Query.
func Having(filters ...FindOption) QueryOption {
	owned := make([]*proto.PropertyFilter, 0, len(filters))
	var valid func(FindOption) bool
	valid = func(option FindOption) bool {
		f, ok := option.(propertyFilter)
		if !ok {
			return false
		}
		for _, child := range f.nested {
			if !valid(child) {
				return false
			}
		}
		return true
	}
	for _, filter := range filters {
		if !valid(filter) {
			return invalidQueryOption("query: Having accepts only property filters")
		}
		owned = append(owned, filter.(propertyFilter).toProto())
	}
	return queryOption(func(req *proto.QueryRequest) { req.Having = append(req.Having, owned...) })
}

type invalidQueryOption string

func (o invalidQueryOption) Apply(config *filterRequest) { config.queryError = string(o) }
