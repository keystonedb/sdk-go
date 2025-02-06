package keystone

import "github.com/keystonedb/sdk-go/proto"

type filterRequest struct {
	Properties     []*proto.PropertyRequest
	Filters        []*proto.PropertyFilter
	Labels         []*proto.EntityLabel
	RelationOf     *proto.RelationOf
	ParentEntityID string
	PerPage        int32
	PageNumber     int32
	SortProperty   string
	SortDescending bool
	ObjectPaths    []string
	ListObjects    bool
}
