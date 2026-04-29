package keystone

import (
	"context"
	"testing"

	"github.com/keystonedb/sdk-go/proto"
)

type queryIndexOffer struct {
	BaseEntity
	ProductIDs []string `keystone:",indexed"`
	IsGlobal   bool     `keystone:",indexed"`
}

func newQueryIndexTestActor(t *testing.T) (*Actor, *MockServer, func()) {
	t.Helper()

	conn, mock, _, server := MockConnection()
	go func() { _ = server.Serve(mockListener) }()
	actor := conn.Actor("query-1", "127.0.0.1", "user-1", "go-test")

	return &actor, mock, func() {
		server.Stop()
		_ = mockListener.Close()
	}
}

func TestActorQueryIndex_WhereContainsOnStringSlice(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()

	mock.QueryIndexFunc = func(_ context.Context, req *proto.QueryIndexRequest) (*proto.QueryIndexResponse, error) {
		if req.GetAuthorization() == nil {
			t.Fatal("expected authorization to be forwarded")
		}
		if req.GetSchema().GetKey() != "query-index-offer" {
			t.Fatalf("expected schema query-index-offer, got %q", req.GetSchema().GetKey())
		}
		if len(req.GetFilters()) != 1 {
			t.Fatalf("expected 1 filter, got %d", len(req.GetFilters()))
		}

		filter := req.GetFilters()[0]
		if filter.GetProperty() != "product_ids" {
			t.Fatalf("expected property product_ids, got %q", filter.GetProperty())
		}
		if filter.GetOperator() != proto.Operator_Contains {
			t.Fatalf("expected contains operator, got %v", filter.GetOperator())
		}
		if len(filter.GetValues()) != 1 {
			t.Fatalf("expected 1 value, got %d", len(filter.GetValues()))
		}
		if filter.GetValues()[0].GetText() != "prod-101" {
			t.Fatalf("expected contains value prod-101, got %q", filter.GetValues()[0].GetText())
		}

		return &proto.QueryIndexResponse{}, nil
	}

	if _, err := actor.QueryIndex(context.Background(), Type(queryIndexOffer{}), []string{"product_ids"},
		WhereContains("product_ids", "prod-101")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestActorQueryIndex_WhereContainsOnStringSliceWithNestedOr(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()

	mock.QueryIndexFunc = func(_ context.Context, req *proto.QueryIndexRequest) (*proto.QueryIndexResponse, error) {
		if len(req.GetFilters()) != 1 {
			t.Fatalf("expected 1 filter, got %d", len(req.GetFilters()))
		}

		filter := req.GetFilters()[0]
		if !filter.GetOr() {
			t.Fatal("expected top-level OR filter")
		}
		if len(filter.GetNested()) != 2 {
			t.Fatalf("expected 2 nested filters, got %d", len(filter.GetNested()))
		}

		global := filter.GetNested()[0]
		if global.GetProperty() != "is_global" {
			t.Fatalf("expected first nested property is_global, got %q", global.GetProperty())
		}
		if global.GetOperator() != proto.Operator_Equal {
			t.Fatalf("expected equal operator, got %v", global.GetOperator())
		}
		if len(global.GetValues()) != 1 || !global.GetValues()[0].GetBool() {
			t.Fatalf("expected global=true predicate, got %+v", global.GetValues())
		}

		contains := filter.GetNested()[1]
		if contains.GetProperty() != "product_ids" {
			t.Fatalf("expected second nested property product_ids, got %q", contains.GetProperty())
		}
		if contains.GetOperator() != proto.Operator_Contains {
			t.Fatalf("expected contains operator, got %v", contains.GetOperator())
		}
		if len(contains.GetValues()) != 1 || contains.GetValues()[0].GetText() != "prod-102" {
			t.Fatalf("expected contains value prod-102, got %+v", contains.GetValues())
		}

		return &proto.QueryIndexResponse{}, nil
	}

	if _, err := actor.QueryIndex(context.Background(), Type(queryIndexOffer{}), []string{"is_global", "product_ids"},
		Or(
			WhereEquals("is_global", true),
			WhereContains("product_ids", "prod-102"),
		)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
