package keystone

import (
	"github.com/keystonedb/sdk-go/proto"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// RelationshipProvider is an interface for entities that can have relationships
type RelationshipProvider interface {
	ClearRelationships() error
	GetRelationships() []*proto.EntityRelationship
	SetRelationships(links []*proto.EntityRelationship)
}

// EmbeddedRelationships is a struct that implements RelationshipProvider
type EmbeddedRelationships struct {
	ksEntityRelationships []*proto.EntityRelationship
}

// ClearRelationships clears the relationships
func (e *EmbeddedRelationships) ClearRelationships() error {
	e.ksEntityRelationships = []*proto.EntityRelationship{}
	return nil
}

// GetRelationships returns the relationships
func (e *EmbeddedRelationships) GetRelationships() []*proto.EntityRelationship {
	return e.ksEntityRelationships
}

// SetRelationships sets the relationships
func (e *EmbeddedRelationships) SetRelationships(links []*proto.EntityRelationship) {
	e.ksEntityRelationships = links
}

// AddRelationship adds a relationship
func (e *EmbeddedRelationships) AddRelationship(relationshipType string, target ID, meta map[string]string, since time.Time) {
	e.ksEntityRelationships = append(e.ksEntityRelationships, &proto.EntityRelationship{
		Relationship: &proto.Key{Key: relationshipType},
		TargetId:     target.String(),
		Data:         meta,
		Since:        timestamppb.New(since),
	})
}
