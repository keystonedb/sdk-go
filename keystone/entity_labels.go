package keystone

import "github.com/keystonedb/sdk-go/proto"

// LabelProvider is an interface for entities that can provide labels
type LabelProvider interface {
	ClearLabels() error
	GetLabels() []*proto.EntityLabel
}

// EmbeddedLabels is a struct that implements LabelProvider
type EmbeddedLabels struct {
	ksEntityLabels []*proto.EntityLabel
}

// ClearLabels clears the labels
func (e *EmbeddedLabels) ClearLabels() error {
	e.ksEntityLabels = []*proto.EntityLabel{}
	return nil
}

// GetLabels returns the labels
func (e *EmbeddedLabels) GetLabels() []*proto.EntityLabel {
	return e.ksEntityLabels
}

// AddLabel adds a label
func (e *EmbeddedLabels) AddLabel(name, value string) {
	e.ksEntityLabels = append(e.ksEntityLabels, &proto.EntityLabel{
		Name:  name,
		Value: value,
	})
}
