package keystone

import "github.com/keystonedb/sdk-go/proto"

type filterRequest struct {
	Properties     []*proto.PropertyRequest
	Filters        []*proto.PropertyFilter
	sortBy         []*proto.PropertySort
	Labels         []*proto.EntityLabel
	RelationOf     *proto.RelationOf
	EntityIds      []string
	ParentEntityID string
	PerPage        int32
	PageNumber     int32
	ObjectPaths    []string
	ListObjects    bool
}
