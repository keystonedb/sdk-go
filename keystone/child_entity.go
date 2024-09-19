package keystone

import "strings"

// BaseChildEntity is a full entity, but keyed under a parent entity
type BaseChildEntity struct {
	BaseEntity
	_parentID string
	_childID  string
}

func (e *BaseChildEntity) SetKeystoneID(id string) {
	e._entityID = id
	split := strings.Split(e._entityID, "-")
	e._parentID = split[0]
	if len(split) > 1 {
		e._childID = split[1]
	}
}

func (e *BaseChildEntity) SetKeystoneParentID(id string) {
	if strings.Contains(id, "-") {
		e.SetKeystoneID(id)
	} else {
		e._parentID = id
	}

	if e._entityID == "" {
		e._entityID = e._parentID
	}
}

func (e *BaseChildEntity) SetKeystoneChildID(id string) {
	e._childID = id
}

func (e *BaseChildEntity) GetKeystoneParentID() string {
	if e._parentID == "" {
		split := strings.Split(e._entityID, "-")
		e._parentID = split[0]
	}
	return e._parentID
}

func (e *BaseChildEntity) GetKeystoneChildID() string {
	if e._childID == "" {
		split := strings.Split(e._entityID, "-")
		if len(split) < 2 {
			return ""
		}

		e._childID = split[1]
	}
	return e._childID
}
