package keystone

import "slices"

import "github.com/keystonedb/sdk-go/proto"

// QueryOption configures an entity query. Existing FindOptions can also be used.
type QueryOption = FindOption

type queryOption func(*proto.QueryRequest)

func (o queryOption) Apply(config *filterRequest) {
	config.queryOptions = append(config.queryOptions, o)
}

// Select adds properties to the result projection. Repeated selections are merged.
func Select(properties ...string) QueryOption {
	properties = append([]string(nil), properties...)
	return queryOption(func(req *proto.QueryRequest) {
		req.Select = appendQueryProperties(req.Select, properties)
	})
}

// SelectSecure selects properties and requests their decrypted values.
func SelectSecure(properties ...string) QueryOption {
	properties = append([]string(nil), properties...)
	return queryOption(func(req *proto.QueryRequest) {
		req.Select = appendQueryProperties(req.Select, properties)
		req.Decrypt = appendQueryProperties(req.Decrypt, properties)
	})
}

func appendQueryProperties(selected, properties []string) []string {
	for _, property := range properties {
		found := slices.Contains(selected, property)
		if !found {
			selected = append(selected, property)
		}
	}
	return selected
}

// HasParent restricts results to children of the given entity, like ChildOf.
func HasParent(entityID string) QueryOption {
	return ChildOf(entityID)
}

// HasRelation requires an outgoing relationship of the given type to any of the
// listed entities. With no entity IDs it matches any destination. The relationship
// type belongs to the actor's vendor/app; use RelationOf or RelationTo for an
// explicit vendor/app or an incoming relationship. Repeated calls are ANDed.
func HasRelation(relationshipType string, entityIDs ...string) QueryOption {
	entityIDs = append([]string(nil), entityIDs...)
	return queryOption(func(req *proto.QueryRequest) {
		predicate := &proto.QueryPredicate{Condition: &proto.QueryPredicate_Relationship{
			Relationship: &proto.QueryRelationship{
				Type:      &proto.Key{Key: relationshipType},
				Direction: proto.QueryRelationship_OUTGOING,
				EntityIds: append([]string(nil), entityIDs...),
			},
		}}
		if req.Predicate == nil {
			req.Predicate = predicate
		} else {
			req.Predicate = &proto.QueryPredicate{Condition: &proto.QueryPredicate_All{
				All: &proto.QueryPredicateGroup{Predicates: []*proto.QueryPredicate{req.Predicate, predicate}},
			}}
		}
	})
}

// Cursor continues a query using the previous response's NextCursor.
func Cursor(cursor string) QueryOption {
	return queryOption(func(req *proto.QueryRequest) { req.Cursor = cursor })
}

// IncludeTotal requests an exact count before the cursor and limit are applied.
func IncludeTotal() QueryOption {
	return queryOption(func(req *proto.QueryRequest) { req.IncludeTotal = true })
}
