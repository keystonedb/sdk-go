package keystone

type sortBy struct {
	property   string
	descending bool
}

func (f sortBy) Apply(config *filterRequest) {
	config.SortProperty = f.property
	config.SortDescending = f.descending
}

func SortBy(property string, descending bool) FindOption {
	return sortBy{property: property, descending: descending}
}
