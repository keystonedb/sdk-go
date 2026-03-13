package keystone

import (
	"github.com/keystonedb/sdk-go/proto"
)

// RetrieveBy is an interface that defines a retriever
type RetrieveBy interface {
	BaseRequest() *proto.EntityRequest
	EntityType() string
}

// byEntityID is a retriever that retrieves an entity by its ID
type byEntityID struct {
	EntityID ID
	Type     string
}

func ByEntityID(entityType interface{}, entityID ID) RetrieveBy {
	return byEntityID{EntityID: entityID, Type: Type(entityType)}
}

// ByHashID creates a retriever that retrieves an entity by its hash ID.
// Note: This function will panic if the entityID contains invalid characters.
// The entityID must not contain the '#' character.
func ByHashID(entityType interface{}, entityID string) RetrieveBy {
	hashID, err := HashID(entityID)
	if err != nil {
		// Panic here to maintain backward compatibility with the original function signature
		// Users should validate input before calling this function, or use HashID directly
		panic("ByHashID: " + err.Error())
	}
	return byEntityID{EntityID: hashID, Type: Type(entityType)}
}

// BaseRequest returns the base byEntityID request
func (l byEntityID) BaseRequest() *proto.EntityRequest {
	return &proto.EntityRequest{
		View:     &proto.EntityView{},
		EntityId: string(l.EntityID),
	}
}

func (l byEntityID) EntityType() string { return l.Type }

// byUniqueProperty is a retriever that retrieves an entity by its unique ID
type byUniqueProperty struct {
	UniqueID string
	Property string
	Type     string
}

func ByUniqueProperty(entityType interface{}, uniqueID, property string) RetrieveBy {
	return byUniqueProperty{UniqueID: uniqueID, Property: property, Type: Type(entityType)}
}

// BaseRequest returns the base byUniqueProperty request
func (l byUniqueProperty) BaseRequest() *proto.EntityRequest {
	return &proto.EntityRequest{
		View: &proto.EntityView{},
		UniqueId: &proto.IDLookup{
			SchemaId: l.Type,
			Property: l.Property,
			UniqueId: l.UniqueID,
		},
	}
}

func (l byUniqueProperty) EntityType() string { return l.Type }

// RetrieveOption is an interface for options to be applied to an entity request
type RetrieveOption interface {
	Apply(config *proto.EntityView)
}

type RetrieveEntityOption interface {
	ApplyRequest(config *proto.EntityRequest)
}

type retrieveOptions []RetrieveOption

func (o retrieveOptions) Apply(config *proto.EntityView) {
	for _, opt := range o {
		opt.Apply(config)
	}
}

func RetrieveOptions(opts ...RetrieveOption) RetrieveOption {
	return retrieveOptions(opts)
}

// WithProperties is a retrieve option that loads properties
func WithProperties(properties ...string) RetrieveOption {
	return propertyLoader{properties: properties}
}

// WithDecryptedProperties is a retrieve option that loads decrypted properties
func WithDecryptedProperties(properties ...string) RetrieveOption {
	return propertyLoader{properties: properties, decrypt: true}
}

// WithProperty is a retrieve option that loads properties
func WithProperty(decrypt bool, properties ...string) RetrieveOption {
	return propertyLoader{properties: properties, decrypt: decrypt}
}

// WithRelationships is a retrieve option that loads relationships
func WithRelationships(keys ...string) RetrieveOption {
	return relationshipsLoader{keys: keys}
}

// WithLabels is a retrieve option that loads labels
func WithLabels() RetrieveOption {
	return labelLoader{labels: true}
}

type propertyLoader struct {
	properties []string
	decrypt    bool
}

func (l propertyLoader) Apply(config *proto.EntityView) {
	if config.Properties == nil {
		config.Properties = make([]*proto.PropertyRequest, 0)
	}

	config.Properties = append(config.Properties, &proto.PropertyRequest{Properties: l.properties, Decrypt: l.decrypt})
}

type relationshipsLoader struct{ keys []string }

func (l relationshipsLoader) Apply(config *proto.EntityView) {
	if config.RelationshipByType == nil {
		config.RelationshipByType = make([]*proto.Key, 0)
	}
	for _, key := range l.keys {
		config.RelationshipByType = append(config.RelationshipByType, &proto.Key{Key: key})
	}
}

type viewName struct{ name string }

func (l viewName) Apply(config *proto.EntityView) { config.Name = l.name }
func WithView(name string) RetrieveOption {
	return viewName{name: name}
}

type summaryLoader struct{ summary bool }

func (l summaryLoader) Apply(config *proto.EntityView) { config.Summary = l.summary }

// WithSummary is a retrieve option that loads summaries
func WithSummary() RetrieveOption {
	return summaryLoader{summary: true}
}

type documentLoader struct {
	doc        *Document
	revisionID string
}

func (l documentLoader) Apply(config *proto.EntityView) {
	config.LatestDocument = l.revisionID == ""
	config.DocumentRevision = l.revisionID
}

func (l documentLoader) ObserveRetrieve(resp *proto.EntityResponse) {
	if l.doc == nil {
		return
	}
	for _, doc := range resp.GetDocuments() {
		if doc.GetRevisionId() != "" && l.revisionID != "" && doc.GetRevisionId() != l.revisionID {
			continue
		}
		l.doc.Data = doc.GetData()
		l.doc.Meta = doc.GetMeta()
		l.doc.RevisionID = doc.GetRevisionId()
	}
}

// WithDocument is a retrieve option that loads the latest document - passing a document pointer to hydrate
func WithDocument(hydrate *Document) RetrieveOption {
	return documentLoader{doc: hydrate}
}

// WithDocumentRevision loads a specific document revision
func WithDocumentRevision(revision string, hydrate *Document) RetrieveOption {
	return documentLoader{revisionID: revision, doc: hydrate}
}

type documentRevisionsLister struct{ listRevisions bool }

func (l documentRevisionsLister) Apply(config *proto.EntityView) {
	config.DocumentRevisions = l.listRevisions
}

// WithDocumentRevisionList lists document revisions
func WithDocumentRevisionList() RetrieveOption {
	return documentRevisionsLister{listRevisions: true}
}

type labelLoader struct{ labels bool }

func (l labelLoader) Apply(config *proto.EntityView) { config.Labels = l.labels }

type relationshipCount struct{ count bool }

func (l relationshipCount) Apply(config *proto.EntityView) { config.RelationshipCount = l.count }

type relationshipTypeCount struct{ relationType, appId, vendorId string }

func (l relationshipTypeCount) Apply(config *proto.EntityView) {
	config.RelationshipCountType = append(config.RelationshipByType, &proto.Key{Source: &proto.VendorApp{
		VendorId: l.vendorId, AppId: l.appId,
	}, Key: l.relationType})
}

func WithTotalRelationshipCount() RetrieveOption {
	return relationshipCount{count: true}
}

func WithRelationshipCount(relationType, appId, vendorId string) RetrieveOption {
	if relationType == "" && appId == "" && vendorId == "" {
		return WithTotalRelationshipCount()
	} else {
		return relationshipTypeCount{
			relationType: relationType,
			appId:        appId,
			vendorId:     vendorId,
		}
	}
}
func WithSiblingRelationshipCount(relationType string) RetrieveOption {
	return relationshipTypeCount{
		relationType: relationType,
	}
}

type childSummary struct{ retrieveSummary bool }

func (l childSummary) Apply(config *proto.EntityView) { config.ChildSummary = l.retrieveSummary }
func WithChildSummary() RetrieveOption                { return childSummary{retrieveSummary: true} }

type descendantTypeCount struct{ entityType, appId, vendorId string }

func (l descendantTypeCount) Apply(config *proto.EntityView) {
	config.DescendantCountType = append(config.DescendantCountType, &proto.Key{Source: &proto.VendorApp{
		VendorId: l.vendorId, AppId: l.appId,
	}, Key: l.entityType})
}

func WithDescendantCount(entityType string) RetrieveOption {
	if entityType == "" {
		return nil
	}

	return descendantTypeCount{
		entityType: entityType,
	}
}

type withLock struct {
	message    string
	ttlSeconds int32
}

func (l withLock) Apply(config *proto.EntityView) {}

func (l withLock) ApplyRequest(config *proto.EntityRequest) {
	config.RequestLock = true
	config.LockTtlSeconds = l.ttlSeconds
	config.LockMessage = l.message
}

func WithLock(note string, ttlSeconds int32) RetrieveOption {
	return withLock{message: note, ttlSeconds: ttlSeconds}
}

type verifiedProperty struct {
	property string
	compare  string
}

func WithVerifiedProperty(property, compare string) RetrieveOption {
	return verifiedProperty{property: property, compare: compare}
}

func (v verifiedProperty) Apply(view *proto.EntityView) {
	view.Properties = append(view.Properties, &proto.PropertyRequest{
		Properties: []string{v.property},
	})
}

func (v verifiedProperty) ApplyRequest(req *proto.EntityRequest) {
	req.VerifyProperties = append(req.VerifyProperties, &proto.EntityProperty{
		Property: v.property,
		Value:    &proto.Value{SecureText: v.compare},
	})
}

type withObjects struct {
	paths []string
}

func (f withObjects) ApplyRequest(req *proto.EntityRequest) {}
func (f withObjects) Apply(view *proto.EntityView) {
	if len(f.paths) > 0 {
		view.ObjectPaths = f.paths
	} else {
		view.ListObjects = true
	}
}

func WithObjects(paths ...string) RetrieveOption {
	return withObjects{paths: paths}
}
