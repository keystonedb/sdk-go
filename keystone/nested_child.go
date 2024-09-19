package keystone

type EmbeddedChild struct {
	_childID   string
	_aggregate int64
}

func (e *EmbeddedChild) ChildID() string             { return e._childID }
func (e *EmbeddedChild) SetChildID(id string)        { e._childID = id }
func (e *EmbeddedChild) AggregateValue() int64       { return e._aggregate }
func (e *EmbeddedChild) SetAggregateValue(val int64) { e._aggregate = val }
