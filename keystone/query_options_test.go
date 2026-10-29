package keystone

import (
	"context"
	"testing"

	"github.com/keystonedb/sdk-go/proto"
	gproto "google.golang.org/protobuf/proto"
)

func TestActorQueryReadableOptions(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	want := &proto.QueryRequest{
		Authorization:  actor.Authorization(),
		From:           &proto.Key{Key: "offer", Source: actor.Authorization().Source},
		Where:          []*proto.PropertyFilter{{Property: "s", Operator: proto.Operator_Equal, Values: []*proto.Value{{Text: "Y"}}}},
		ParentEntityId: "parent",
		Select:         []string{"prop", "prop2", "secProp2"}, Decrypt: []string{"secProp2"},
		Limit: 11, Cursor: "next", IncludeTotal: true,
		Predicate: &proto.QueryPredicate{Condition: &proto.QueryPredicate_Relationship{
			Relationship: &proto.QueryRelationship{Type: &proto.Key{Key: "owns"}, EntityIds: []string{"related"}},
		}},
	}
	mock.QueryFunc = func(_ context.Context, req *proto.QueryRequest) (*proto.QueryResponse, error) {
		if !gproto.Equal(req, want) {
			t.Errorf("request = %v, want %v", req, want)
		}
		return &proto.QueryResponse{}, nil
	}
	_, err := actor.Query(context.Background(), "offer",
		Where("s", "=", "Y"), HasParent("parent"), HasRelation("owns", "related"),
		Select("prop", "prop2"), SelectSecure("secProp2"), Limit(11), Cursor("next"), IncludeTotal())
	if err != nil {
		t.Fatal(err)
	}
}

func TestActorQueryOptionsMergeAndReuse(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	properties := []string{"name", "secret"}
	ids := []string{"related", "other"}
	options := []QueryOption{
		Select(properties...), SelectSecure("secret", "secret"), Select("name"),
		HasRelation("owns", ids...), HasRelation("contains"), WithLabel("category", "featured"),
	}
	properties[0] = "changed"
	ids[0] = "changed"
	request := &proto.QueryRequest{Select: []string{"original"}, Predicate: &proto.QueryPredicate{
		Condition: &proto.QueryPredicate_Label{Label: &proto.EntityLabel{Name: "existing"}},
	}}
	original := gproto.Clone(request)
	var first *proto.QueryRequest
	calls := 0
	mock.QueryFunc = func(_ context.Context, req *proto.QueryRequest) (*proto.QueryResponse, error) {
		calls++
		if calls == 1 {
			first = gproto.Clone(req).(*proto.QueryRequest)
		} else if !gproto.Equal(req, first) {
			t.Errorf("reusing options changed the request: %v", req)
		}
		if len(req.Select) != 3 || req.Select[1] != "name" || req.Select[2] != "secret" || len(req.Decrypt) != 1 || req.Decrypt[0] != "secret" {
			t.Errorf("projections not merged: %v", req)
		}
		var relationships []*proto.QueryRelationship
		var labels []string
		var visit func(*proto.QueryPredicate)
		visit = func(p *proto.QueryPredicate) {
			if rel := p.GetRelationship(); rel != nil {
				relationships = append(relationships, rel)
			}
			if label := p.GetLabel(); label != nil {
				labels = append(labels, label.Name)
			}
			for _, child := range p.GetAll().GetPredicates() {
				visit(child)
			}
		}
		visit(req.Predicate)
		if len(relationships) != 2 || len(labels) != 2 {
			t.Fatalf("predicates lost: %v", req.Predicate)
		}
		if relationships[0].GetType().GetKey() != "owns" || len(relationships[0].EntityIds) != 2 || relationships[0].EntityIds[0] != "related" || relationships[1].GetType().GetKey() != "contains" || len(relationships[1].EntityIds) != 0 {
			t.Errorf("relationship constraints changed: %v", req.Predicate)
		}
		// Mutation of the outgoing request must not affect a reusable option.
		relationships[0].EntityIds[0] = "mutated"
		return &proto.QueryResponse{}, nil
	}
	for range 2 {
		if _, err := actor.QueryWithRequest(context.Background(), "offer", request, options...); err != nil {
			t.Fatal(err)
		}
	}
	if !gproto.Equal(request, original) {
		t.Fatal("caller request was modified")
	}
}
