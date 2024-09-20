package keystone

import "github.com/keystonedb/sdk-go/proto"

type withLabel struct {
	name  string
	value string
}

func (f withLabel) Apply(config *filterRequest) {
	if config.Labels == nil {
		config.Labels = make([]*proto.EntityLabel, 0)
	}

	config.Labels = append(config.Labels, &proto.EntityLabel{
		Name:  f.name,
		Value: f.value,
	})
}

func WithLabel(name, value string) FindOption {
	return withLabel{name: name, value: value}
}
