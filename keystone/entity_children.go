package keystone

import (
	"github.com/keystonedb/sdk-go/proto"
	"github.com/kubex/k4id"
	"reflect"
	"time"
)

// ChildProvider is an interface for entities that can have Children
type ChildProvider interface {
	GetChildrenToStore() []*proto.EntityChild
	GetChildrenToRemove() []*proto.EntityChild
	ClearChildChanges() error
	GetTruncateChildrenType() []*proto.Key
}

type ChildUpdateProvider interface {
	updateChildren(resp *proto.MutateResponse)
}

// EmbeddedChildren is a struct that implements ChildProvider
type EmbeddedChildren struct {
	ksEntityChildren map[string]NestedChild
	// ksEntityChildrenToRemove is a map keyed by child ID, with the value being the child type
	ksEntityChildrenToRemove map[string]string
	ksTruncateChildrenType   []*proto.Key
}

// WARNING: Adding children by this same type will not work in a single mutate call
func (e *EmbeddedChildren) TruncateByType(childType interface{}) {
	e.ksTruncateChildrenType = append(e.ksTruncateChildrenType, &proto.Key{Key: Type(childType)})
}

func (e *EmbeddedChildren) ClearChildChanges() error {
	e.ksEntityChildren = make(map[string]NestedChild)
	e.ksEntityChildrenToRemove = nil
	e.ksTruncateChildrenType = nil
	return nil
}

func (e *EmbeddedChildren) updateChildren(resp *proto.MutateResponse) {
	for writeRef, cid := range resp.GetCreatedChildren() {
		e.setChildID(writeRef, cid)
	}
}

func (e *EmbeddedChildren) setChildID(writeReference, cid string) {
	for ref, child := range e.ksEntityChildren {
		if ref == writeReference && child.ChildID() == "" {
			child.SetChildID(cid)
		}
	}
}

func (e *EmbeddedChildren) GetChildrenToStore() []*proto.EntityChild {
	var children []*proto.EntityChild
	for writeRef, child := range e.ksEntityChildren {
		eChild := &proto.EntityChild{
			WriteReference: writeRef,
			Cid:            child.ChildID(),
		}

		if c, o := child.(*DynamicChild); o {
			eChild.Type = &proto.Key{Key: c.keyType()}
		} else {
			eChild.Type = &proto.Key{Key: Type(child)}
		}

		if nca, ok := child.(NestedChildAggregateValue); ok {
			eChild.Value = nca.AggregateValue()
		}

		simpleData := true
		if nc, ok := child.(NestedChildDataProvider); ok {
			eChild.Data = nc.KeystoneData()
			simpleData = false
		}
		if nc, ok := child.(NestedChildDataMutator); ok {
			eChild.AppendData = nc.KeystoneDataAppend()
			eChild.RemoveData = nc.KeystoneRemoveData()
			simpleData = false
		}

		if simpleData {
			eChild.Data = ToByteMap(child)
		}

		children = append(children, eChild)
	}
	return children
}

func (e *EmbeddedChildren) GetTruncateChildrenType() []*proto.Key {
	return e.ksTruncateChildrenType
}

func (e *EmbeddedChildren) GetChildrenToRemove() []*proto.EntityChild {
	if e.ksEntityChildrenToRemove == nil {
		return nil
	}

	var children []*proto.EntityChild
	for cid, childType := range e.ksEntityChildrenToRemove {
		children = append(children, &proto.EntityChild{
			Type: &proto.Key{Key: childType},
			Cid:  cid,
		})
	}
	return children
}

// RemoveChild removes a child from the list of children, and prepares for removal
func (e *EmbeddedChildren) RemoveChild(childType any, cid ...string) {
	if e.ksEntityChildrenToRemove == nil {
		e.ksEntityChildrenToRemove = make(map[string]string)
	}
	for _, c := range cid {
		e.ksEntityChildrenToRemove[c] = Type(childType)
	}
}

// AddChildren adds multiple children to storage
func (e *EmbeddedChildren) AddChildren(child ...any) {
	for _, c := range child {
		refC := reflect.ValueOf(c)
		if refC.Kind() == reflect.Slice {
			for i := 0; i < refC.Len(); i++ {
				if nc, ok := refC.Index(i).Interface().(NestedChild); ok {
					e.AddChild(nc)
				}
			}
			continue
		}

		if nc, ok := c.(NestedChild); ok {
			e.AddChild(nc)
		}
	}
}

// AddChild adds a child to storage
func (e *EmbeddedChildren) AddChild(child NestedChild) {
	if e.ksEntityChildren == nil {
		e.ksEntityChildren = make(map[string]NestedChild)
	}
	for _, existingChild := range e.ksEntityChildren {
		if (child.ChildID() != "" && existingChild.ChildID() == child.ChildID()) || reflect.DeepEqual(existingChild, child) {
			return
		}
	}
	if child.ChildID() != "" {
		e.ksEntityChildren[child.ChildID()] = child
	} else {
		e.ksEntityChildren[k4id.TimeGeneratorNano.Generate(time.Now())] = child
	}
}
