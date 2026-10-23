package keystone

import (
	"fmt"
	"github.com/keystonedb/sdk-go/proto"
	"testing"
)

type pendingTestChild struct {
	Child
	Name string
}

func TestEmbeddedChildrenPreservesPendingChildren(t *testing.T) {
	for parent := 0; parent < 1000; parent++ {
		var e EmbeddedChildren
		children := make([]*pendingTestChild, 10)
		for n := range children {
			children[n] = &pendingTestChild{Name: fmt.Sprint(n)}
			e.AddChild(children[n])
		}
		pending := e.GetChildrenToStore()
		if len(pending) != len(children) {
			t.Fatalf("parent %d: got %d pending children, want %d", parent, len(pending), len(children))
		}
		ids := make(map[string]string)
		for n, child := range pending {
			if child.WriteReference == "" {
				t.Fatal("empty write reference")
			}
			if _, exists := ids[child.WriteReference]; exists {
				t.Fatal("duplicate write reference")
			}
			ids[child.WriteReference] = fmt.Sprint(n)
		}
		e.updateChildren(&proto.MutateResponse{CreatedChildren: ids})
		for _, child := range children {
			if child.ChildID() == "" {
				t.Fatal("child ID not assigned")
			}
		}
	}
}

func TestEmbeddedChildrenDeduplicatesAndClears(t *testing.T) {
	var e EmbeddedChildren
	child := &pendingTestChild{Name: "pending"}
	e.AddChild(child)
	e.AddChild(child)
	persisted := &pendingTestChild{Name: "persisted"}
	persisted.SetChildID("existing")
	e.AddChild(persisted)
	e.AddChild(persisted)
	if len(e.GetChildrenToStore()) != 2 {
		t.Fatal("children were not deduplicated")
	}
	if err := e.ClearChildChanges(); err != nil {
		t.Fatal(err)
	}
	e.AddChild(&pendingTestChild{Name: "new"})
	if len(e.GetChildrenToStore()) != 1 {
		t.Fatal("children were not cleared")
	}
}
