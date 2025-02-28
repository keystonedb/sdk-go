package keystone

type Child struct {
	_childID   string
	_aggregate int64
}

func (e *Child) ChildID() string      { return e._childID }
func (e *Child) SetChildID(id string) { e._childID = id }

func (e *Child) AggregateValue() int64       { return e._aggregate }
func (e *Child) SetAggregateValue(val int64) { e._aggregate = val }

type DynamicChild struct {
	Child
	_src    any
	_type   string
	_data   map[string][]byte
	_append map[string][]byte
	_reduce map[string]bool
}

func (e *DynamicChild) ReplaceData(with map[string][]byte) {
	e._data = with
	e._append = make(map[string][]byte)
	e._reduce = make(map[string]bool)
}

func (e *DynamicChild) Append(key string, value []byte) {
	e._append[key] = value
	delete(e._reduce, key)
}

func (e *DynamicChild) Reduce(key string) {
	e._reduce[key] = true
	delete(e._append, key)
}

func (e *DynamicChild) keyType() string {
	if e._type == "" && e._src != nil {
		e._type = Type(e._src)
	}
	return e._type
}

func (e *DynamicChild) KeystoneData() map[string][]byte {
	return e._data
}

func (e *DynamicChild) KeystoneDataAppend() map[string][]byte {
	return e._append
}
func (e *DynamicChild) KeystoneRemoveData() []string {
	var keys []string
	for k := range e._reduce {
		keys = append(keys, k)
	}
	return keys
}

func (e *DynamicChild) SetChildID(id string) {
	e._childID = id
	if e._src == nil {
		return
	}
	if c, o := e._src.(NestedChild); o {
		c.SetChildID(id)
	}
}

func NewDynamicChild(from interface{}) *DynamicChild {
	if from == nil {
		return nil
	}

	c := &DynamicChild{
		_src:    from,
		_type:   Type(from),
		_data:   make(map[string][]byte),
		_append: make(map[string][]byte),
		_reduce: make(map[string]bool),
	}

	if nc, ok := from.(NestedChild); ok {
		c.SetChildID(nc.ChildID())
	}

	if nc, ok := from.(NestedChildDataProvider); ok {
		c._data = nc.KeystoneData()
	}

	if nc, ok := from.(NestedChildDataMutator); ok {
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
