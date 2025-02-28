package keystone

import "github.com/keystonedb/sdk-go/proto"

// WithChildren is a retrieve option that loads Children
func WithChildren(childType string, ids ...string) ChildLoader {
	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = false
	}
	return ChildLoader{childType: childType, ids: idMap}
}

type ChildLoader struct {
	childType string
	ids       map[string]bool
	loaded    []*proto.EntityChild
}

func ChildrenFromLoader[T any](loader ChildLoader) []*T {
	entities := make([]*T, 0)
	for _, child := range loader.loaded {
		entity := new(T)
		setChildData(entity, child.GetCid(), child.GetValue(), child.GetData())
		entities = append(entities, entity)
	}
	return entities
}

func setChildData(child any, cid string, val int64, data map[string][]byte) {
	if c, ok := child.(NestedChild); ok && cid != "" {
		c.SetChildID(cid)
	}
	if c, ok := child.(NestedChildAggregateValue); ok {
		c.SetAggregateValue(val)
	}

	if c, ok := child.(NestedChildDataReceiver); ok {
		c.FromKeystoneData(data)
	} else {
		FromByteMap(data, child)
	}
}

func (l *ChildLoader) ObserveRetrieve(resp *proto.EntityResponse) {
	if len(resp.GetChildren()) == 0 {
		return
	}
	for _, child := range resp.GetChildren() {
		if child.GetType().GetKey() != l.childType {
			continue
		}
		if len(l.ids) == 0 {
			l.loaded = append(l.loaded, child)
		} else {
			if _, ok := l.ids[child.GetCid()]; ok {
				l.loaded = append(l.loaded, child)
				l.ids[child.GetCid()] = true
			}
		}
	}
}
func (l *ChildLoader) Apply(config *proto.EntityView) {
	if config.Children == nil {
		config.Children = make([]*proto.ChildRequest, 0)
	}

	ids := make([]string, 0)
	for id := range l.ids {
		ids = append(ids, id)
	}

	config.Children = append(config.Children, &proto.ChildRequest{
		Type: &proto.Key{Key: l.childType},
		Cid:  ids,
	})
}
