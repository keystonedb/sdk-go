package keystone

type withEntityIDs struct {
	entityIDs []string
}

func (f withEntityIDs) Apply(config *filterRequest) {
	config.EntityIds = f.entityIDs
}

func WithEntityIDs(ids []string) FindOption {
	return withEntityIDs{entityIDs: ids}
}
