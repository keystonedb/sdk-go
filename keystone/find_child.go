package keystone

type childOf struct {
	parentEntityID string
}

func (f childOf) Apply(config *filterRequest) {
	config.ParentEntityID = f.parentEntityID
}

func ChildOf(entityID string) FindOption {
	return childOf{
		parentEntityID: entityID,
	}
}
