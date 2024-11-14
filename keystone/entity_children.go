package keystone

import (
	"errors"
	"github.com/keystonedb/sdk-go/proto"
)

var ErrNotNestedChild = errors.New("not a nested child")

// ChildProvider is an interface for entities that can have Children
type ChildProvider interface {
	GetChildrenToStore() []*proto.EntityChild
	GetChildrenToRemove() []*proto.EntityChild
}

// EmbeddedChildren is a struct that implements ChildProvider
type EmbeddedChildren struct {
	ksEntityChildren         []*proto.EntityChild
	ksEntityChildrenToRemove []*proto.EntityChild
}

func (e *EmbeddedChildren) GetChildrenToStore() []*proto.EntityChild {
	return e.ksEntityChildren
}

func (e *EmbeddedChildren) GetChildrenToRemove() []*proto.EntityChild {
	return e.ksEntityChildrenToRemove
}

// AddChild adds a child to the entity
func (e *EmbeddedChildren) AddChild(child interface{}, appendTo ...any) error {
	_, isNested := child.(NestedChild)
	if !isNested {
		return ErrNotNestedChild
	}
	e.ksEntityChildren = append(e.ksEntityChildren, NewChild(child).toProtoChild())
	return nil
}
