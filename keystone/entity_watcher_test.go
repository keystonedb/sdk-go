package keystone

import (
	"github.com/keystonedb/sdk-go/proto"
	"testing"
)

type watchedTestEntity struct {
	BaseEntity
	Name string
	Age  int
}

func Test_EntityWatcher(t *testing.T) {
	entity := watchedTestEntity{}
	resp := &proto.EntityResponse{Entity: &proto.Entity{EntityId: "abc"}, Properties: []*proto.EntityProperty{{Property: "name", Value: &proto.Value{Text: "nma"}}}}

	err := Unmarshal(resp, &entity)
	if err != nil {
		t.Error(err)
	}

	changes, err := entity.Watcher().Changes(entity, false)
	if err != nil {
		t.Error(err)
	}
	if len(changes) != 0 {
		t.Errorf("expected 0 changes, got %d", len(changes))
	}

	entity.Name = "new name"
	changes, err = entity.Watcher().Changes(entity, false)
	if err != nil {
		t.Error(err)
	}
	if len(changes) != 1 {
		t.Errorf("expected 1 change, got %d", len(changes))
	}

}
