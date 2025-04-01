package keystone

import (
	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

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

// OnConflictIgnore will skip the mutation if the entity already exists
func OnConflictIgnore() MutateOption {
	return onConflictIgnore{}
}

type onConflictIgnore struct {
}

func (m onConflictIgnore) apply(mutate *proto.MutateRequest) {
	mutate.Options = append(mutate.Options, proto.MutateRequest_OnConflictIgnore)
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
		delete(w.knownValues, prop)
	}
	return nil
}

type matchExisting struct {
	findOptions []FindOption
}

func (m matchExisting) apply(mutate *proto.MutateRequest) {
	for _, opt := range m.findOptions {
		if filterOpt, ok := opt.(propertyFilter); ok {
			mutate.Where = append(mutate.Where, &proto.PropertyFilter{
				Property: filterOpt.key,
				Operator: filterOpt.operator,
				Values:   filterOpt.values,
			})
		}
	}
}

func MatchExisting(options ...FindOption) MutateOption {
	return matchExisting{findOptions: options}
}

type prepareObjects struct {
	objects []*EntityObject
}

func (m prepareObjects) apply(mutate *proto.MutateRequest) {
	for _, obj := range m.objects {
		pObj := &proto.EntityObject{
			Path:   obj.GetPath(),
			Type:   obj.storageClass,
			Public: obj.public,
			Data:   obj.data,
		}
		if !obj.expiry.IsZero() {
			pObj.Expiry = timestamppb.New(obj.expiry)
		}
		mutate.Mutation.Objects = append(mutate.Mutation.Objects, pObj)
	}
}
func (m prepareObjects) ObserveMutation(response *proto.MutateResponse) {
	if !response.GetSuccess() {
		return
	}
	for _, obj := range m.objects {
		for _, respObj := range response.SignedObjectUrls {
			if obj.GetPath() == respObj.GetPath() && respObj.GetUrl() != "" {
				obj.uploadURL = respObj.GetUrl()
				obj.uploadHeaders = respObj.GetUploadHeaders()
				break
			}
		}
	}
}

func PrepareUploads(objs ...*EntityObject) MutateOption {
	return prepareObjects{objects: objs}
}

func WithPiiToken(piiToken string) MutateOption {
	return withPiiToken{piiToken: piiToken}
}

type withPiiToken struct {
	piiToken string
}

func (m withPiiToken) apply(mutate *proto.MutateRequest) {
	mutate.Mutation.PiiToken = m.piiToken
}

type withPiiReference struct {
	vendorId string
	appId    string
	piiKey   string
}

func WithPiiReference(vendorId, appId, piiKey string) MutateOption {
	return withPiiReference{
		vendorId: vendorId,
		appId:    appId,
		piiKey:   piiKey,
	}
}

func (m withPiiReference) apply(mutate *proto.MutateRequest) {
	mutate.Mutation.PiiReference = &proto.Key{Key: m.piiKey, Source: proto.NewVendorApp(m.vendorId, m.appId)}
}
