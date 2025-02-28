package keystone

type Child struct {
	_src       any
	_childID   string
	_aggregate int64
	_type      string
	_data      map[string][]byte
	_append    map[string][]byte
	_reduce    map[string]bool
}

func NewChild(from interface{}) *Child {
	if from == nil {
		return nil
	}

	c := &Child{
		_src:    from,
		_type:   Type(from),
		_data:   make(map[string][]byte),
		_append: make(map[string][]byte),
		_reduce: make(map[string]bool),
	}

	if nc, ok := from.(NestedChild); ok {
		c.SetChildID(nc.ChildID())
		c._data = nc.KeystoneData()
		c._append = nc.KeystoneDataAppend()
		for _, key := range nc.KeystoneRemoveData() {
			c._reduce[key] = true
		}
	}

	if nca, ok := from.(NestedChildAggregateValue); ok {
		c.SetAggregateValue(nca.AggregateValue())
	}

	return c
}

func (e *Child) ReplaceData(with map[string][]byte) {
	e._data = with
	e._append = make(map[string][]byte)
	e._reduce = make(map[string]bool)
}

func (e *Child) Append(key string, value []byte) {
	e._append[key] = value
	delete(e._reduce, key)
}

func (e *Child) Reduce(key string) {
	e._reduce[key] = true
	delete(e._append, key)
}

func (e *Child) keyType() string {
	if e._type == "" && e._src != nil {
		e._type = Type(e._src)
	}
	return e._type
}
func (e *Child) ChildID() string { return e._childID }
func (e *Child) SetChildID(id string) {
	e._childID = id
	if e._src == nil {
		return
	}
	if c, o := e._src.(NestedChild); o {
		c.SetChildID(id)
	}
}

func (e *Child) AggregateValue() int64                 { return e._aggregate }
func (e *Child) SetAggregateValue(val int64)           { e._aggregate = val }
func (e *Child) KeystoneData() map[string][]byte       { return e._data }
func (e *Child) KeystoneDataAppend() map[string][]byte { return nil }
func (e *Child) KeystoneRemoveData() []string          { return nil }
