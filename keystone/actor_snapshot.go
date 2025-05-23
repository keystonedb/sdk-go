package keystone

import (
	"context"
	"github.com/keystonedb/sdk-go/proto"
)

func (a *Actor) Snapshot(entityType interface{}, eid ID) (bool, error) {
	schema, _ := a.connection.registerType(entityType)
	resp, err := a.Connection().SnapshotReport(context.Background(), &proto.SnapshotReportRequest{
		Authorization: a.Authorization(),
		EntityId:      eid.String(),
		Schema:        &proto.Key{Key: schema.Type, Source: a.VendorApp()},
	})

	if resp == nil {
		return false, err
	}
	return resp.GetSuccess(), err
}
