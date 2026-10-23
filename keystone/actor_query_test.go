package keystone

import (
	"context"
	"testing"

	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	gproto "google.golang.org/protobuf/proto"
)

func TestActorQueryRequestAndResponse(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	total := int64(3)
	request := &proto.QueryRequest{
		Authorization: &proto.Authorization{Token: "ignored"}, From: &proto.Key{Key: "ignored"},
		Select: []string{"secret", "name"}, Decrypt: []string{"secret"},
		Where:     []*proto.PropertyFilter{{Property: "name", Operator: proto.Operator_Equal, Values: []*proto.Value{{Text: "test"}}}},
		Predicate: &proto.QueryPredicate{Condition: &proto.QueryPredicate_Not{Not: &proto.QueryPredicate{Condition: &proto.QueryPredicate_Label{Label: &proto.EntityLabel{Name: "hidden"}}}}},
		OrderBy:   []*proto.PropertySort{{Property: "name", NullsFirst: true}},
		Limit:     2, EntityIds: &proto.EntityIDList{Values: []string{"entity"}}, ParentEntityId: "parent",
		GroupBy: []string{"name"}, Aggregates: []*proto.QueryAggregate{{Function: proto.QueryAggregate_COUNT, As: "count", Property: "name", Distinct: true}},
		Having: []*proto.PropertyFilter{{Property: "count", Operator: proto.Operator_GreaterThan, Values: []*proto.Value{{Int: 1}}}},
		Cursor: "cursor", IncludeTotal: true, Explain: true,
		Include: &proto.QueryInclude{Labels: true, Relationships: []*proto.Key{{Key: "owns"}}, RelationshipCount: true, RelationshipCountTypes: []*proto.Key{{Key: "owns"}}},
	}
	original := gproto.Clone(request)
	want := gproto.Clone(request).(*proto.QueryRequest)
	want.Authorization = actor.Authorization()
	want.From = &proto.Key{Key: "offer", Source: actor.Authorization().Source}
	response := &proto.QueryResponse{Entities: []*proto.EntityResponse{{Entity: &proto.Entity{EntityId: "entity"}}}, Rows: []*proto.QueryRow{{Values: map[string]*proto.Value{"count": {Int: 3}}}}, NextCursor: "next", HasMore: true, TotalResults: &total, Execution: &proto.QueryExecution{Strategy: "index", ReturnedRows: 1}}
	mock.QueryFunc = func(_ context.Context, req *proto.QueryRequest) (*proto.QueryResponse, error) {
		if !gproto.Equal(req, want) {
			t.Errorf("request mismatch: got %v, want %v", req, want)
		}
		return response, nil
	}
	got, err := actor.QueryWithRequest(context.Background(), "offer", request)
	if err != nil || !gproto.Equal(got, response) {
		t.Fatalf("response = %v, error = %v", got, err)
	}
	if !gproto.Equal(request, original) {
		t.Fatal("caller request was modified")
	}
}

func TestActorQueryFindOptions(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	options := []FindOption{
		WhereEquals("name", "test"), WhereNotEquals("name", "other"),
		WhereGreaterThan("age", 1), WhereGreaterThanOrEquals("age", 2),
		WhereLessThan("age", 9), WhereLessThanOrEquals("age", 8),
		WhereContains("product_ids", "prod-101"), WhereIn("product_ids", "a", "b"), WhereNotIn("product_ids", "c"),
		Or(WhereEquals("is_global", true), And(WhereContains("product_ids", "prod-102"), WhereEquals("state", 1))),
		WhereNotContains("name", "bad"), WhereStartsWith("name", "te"), WhereEndsWith("name", "st"), WhereBetween("age", 1, 9), IsNull("deleted"), IsNotNull("name"),
		OnlyActive(), SortDesc("modified"), SortAsc("name"), SortByNullFirst("age", true),
		Limit(25, 0), WithLabel("category", "featured"), ChildOf("parent"), WithEntityIDs([]string{"a", "b"}),
	}
	expected := &filterRequest{}
	for _, opt := range options {
		opt.Apply(expected)
	}
	mock.QueryFunc = func(_ context.Context, req *proto.QueryRequest) (*proto.QueryResponse, error) {
		if !gproto.Equal(&proto.QueryRequest{Where: req.Where, OrderBy: req.OrderBy}, &proto.QueryRequest{Where: expected.Filters, OrderBy: expected.sortBy}) {
			t.Error("filters or ordering lost")
		}
		if req.Limit != 25 || req.ParentEntityId != "parent" || len(req.GetEntityIds().GetValues()) != 2 {
			t.Errorf("constraints lost: %v", req)
		}
		if req.GetPredicate().GetLabel().GetName() != "category" {
			t.Error("label lost")
		}
		return &proto.QueryResponse{}, nil
	}
	if _, err := actor.Query(context.Background(), "offer", options...); err != nil {
		t.Fatal(err)
	}
}

func TestActorQueryRelationshipDirections(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	for _, tc := range []struct {
		option    FindOption
		direction proto.QueryRelationship_Direction
	}{
		{RelationOf("related", "owns", "vendor", "app"), proto.QueryRelationship_INCOMING},
		{RelationTo(ID("related"), "owns", "vendor", "app"), proto.QueryRelationship_OUTGOING},
		{RelationOfSibling("related", "owns"), proto.QueryRelationship_INCOMING},
		{RelationToSibling(ID("related"), "owns"), proto.QueryRelationship_OUTGOING},
	} {
		mock.QueryFunc = func(_ context.Context, req *proto.QueryRequest) (*proto.QueryResponse, error) {
			all := req.GetPredicate().GetAll().GetPredicates()
			if len(all) != 2 {
				t.Fatalf("expected original predicate AND relationship, got %v", req.Predicate)
			}
			rel := all[1].GetRelationship()
			if rel.GetDirection() != tc.direction || rel.GetType().GetKey() != "owns" || len(rel.EntityIds) != 1 || rel.EntityIds[0] != "related" {
				t.Errorf("unexpected relationship %v", rel)
			}
			return &proto.QueryResponse{}, nil
		}
		_, err := actor.QueryWithRequest(context.Background(), "offer", &proto.QueryRequest{Predicate: &proto.QueryPredicate{Condition: &proto.QueryPredicate_Label{Label: &proto.EntityLabel{Name: "test"}}}}, tc.option)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestActorQueryEntityIDPresence(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	for _, tc := range []struct {
		name    string
		options []FindOption
		present bool
	}{
		{"unrestricted", nil, false}, {"empty", []FindOption{WithEntityIDs([]string{})}, true}, {"nil IDs", []FindOption{WithEntityIDs(nil)}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mock.QueryFunc = func(_ context.Context, req *proto.QueryRequest) (*proto.QueryResponse, error) {
				if (req.EntityIds != nil) != tc.present {
					t.Errorf("entity ID presence lost: %v", req.EntityIds)
				}
				if len(req.Select) != 0 || len(req.Where) != 0 {
					t.Error("unexpected implicit selection or state filter")
				}
				return &proto.QueryResponse{}, nil
			}
			if _, err := actor.Query(context.Background(), "offer", tc.options...); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestActorQueryErrors(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	mock.QueryFunc = func(context.Context, *proto.QueryRequest) (*proto.QueryResponse, error) {
		return nil, status.Error(codes.PermissionDenied, "denied")
	}
	if _, err := actor.Query(context.Background(), "offer"); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("error = %v", err)
	}
	for _, opt := range []FindOption{nil, Limit(10, 1)} {
		if _, err := actor.Query(context.Background(), "offer", opt); err == nil {
			t.Fatal("expected invalid option error")
		}
	}
}
