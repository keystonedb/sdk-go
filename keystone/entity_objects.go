package keystone

import "github.com/keystonedb/sdk-go/proto"

// ObjectProvider is an interface for entities that can provide Objects
type ObjectProvider interface {
	ClearObjects() error
	GetObjects() []*proto.EntityObject
	addObject(*proto.EntityObject)
}

// EmbeddedObjects is a struct that implements ObjectProvider
type EmbeddedObjects struct {
	ksEntityObjects []*proto.EntityObject
}

// ClearObjects clears the Objects
func (e *EmbeddedObjects) ClearObjects() error {
	e.ksEntityObjects = []*proto.EntityObject{}
	return nil
}

// GetObjects returns the Objects
func (e *EmbeddedObjects) GetObjects() []*proto.EntityObject {
	return e.ksEntityObjects
}

// GetObject returns an Object by path
func (e *EmbeddedObjects) GetObject(path string) *proto.EntityObject {
	for _, obj := range e.ksEntityObjects {
		if obj.Path == path {
			return obj
		}
	}
	return nil
}

// addObject adds an Object
func (e *EmbeddedObjects) addObject(obj *proto.EntityObject) {
	e.ksEntityObjects = append(e.ksEntityObjects, obj)
}
