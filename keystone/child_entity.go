package keystone

// BaseChildEntity is a full entity, but keyed under a parent entity
type BaseChildEntity struct {
	BaseEntity
	_parentID string
	_childID  string
}

func (e *BaseChildEntity) SetKeystoneID(id ID) {
	e._entityID = id
	e._parentID = id.ParentID()
	e._childID = id.ChildID()
}

func (e *BaseChildEntity) SetKeystoneParentID(id ID) {
	e._parentID = id.ParentID()
	if e._entityID == "" {
		if e._childID != "" {
			e._entityID = ID(id.ParentID() + "-" + e._childID)
		} else {
			e._entityID = ID(id.ParentID())
		}
	}
}

func (e *BaseChildEntity) SetKeystoneChildID(id string) {
	e._childID = id
}

func (e *BaseChildEntity) GetKeystoneParentID() string {
	if e._parentID == "" {
		return e._entityID.ParentID()
	}
	return e._parentID
}

func (e *BaseChildEntity) GetKeystoneChildID() string {
	if e._childID == "" {
		return e._entityID.ChildID()
	}
	return e._childID
}
