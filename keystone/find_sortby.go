package keystone

import "github.com/keystonedb/sdk-go/proto"

type sortBy struct {
	property   string
	descending bool
	nullFirst  bool
}

func (f sortBy) Apply(config *filterRequest) {
	config.sortBy = append(config.sortBy, &proto.PropertySort{
		Property:   f.property,
		Descending: f.descending,
		NullsFirst: f.nullFirst,
	})
}

func SortBy(property string, descending bool) FindOption {
	return sortBy{property: property, descending: descending}
}
func SortDesc(property string) FindOption {
	return sortBy{property: property, descending: true}
}
func SortAsc(property string) FindOption {
	return sortBy{property: property, descending: false}
}
func SortByNullFirst(property string, descending bool) FindOption {
	return sortBy{property: property, descending: descending, nullFirst: true}
}
