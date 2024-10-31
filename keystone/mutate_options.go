package keystone

import "github.com/keystonedb/sdk-go/proto"

type MutateOption interface {
	apply(*proto.MutateRequest)
}

type MutationOptionWatcherPrepare interface {
	prepare(*Watcher) error
}

func WithMutationComment(comment string) MutateOption {
	return withMutationComment{Comment: comment}
}

type withMutationComment struct {
	Comment string
}

func (m withMutationComment) apply(mutate *proto.MutateRequest) {
	mutate.Mutation.Comment = m.Comment
}

// OnConflictUseID should set the unique properties that can be used to identify an existing identity
func OnConflictUseID(property ...string) MutateOption {
	return onConflictUseID{Property: property}
}

type onConflictUseID struct {
	Property []string
}

func (m onConflictUseID) apply(mutate *proto.MutateRequest) {
	mutate.ConflictUniquePropertyAcquire = m.Property
}

// MutateProperties Only mutate the specified properties
func MutateProperties(property ...string) MutateOption {
	return mutateProperties{Property: property}
}

type mutateProperties struct {
	Property []string
}

func (m mutateProperties) apply(mutate *proto.MutateRequest) {
	var keepProps []*proto.EntityProperty

	for _, prop := range mutate.Mutation.Properties {
		for _, p := range m.Property {
			if prop.Property == p {
				keepProps = append(keepProps, prop)
				break
			}
		}
	}
	mutate.Mutation.Properties = keepProps
}

func (m mutateProperties) prepare(w *Watcher) error {
	for _, prop := range m.Property {
		delete(w.knownValues, knownProperty(prop))
	}
	return nil
}
