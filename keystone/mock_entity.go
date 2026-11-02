package keystone

import "github.com/keystonedb/sdk-go/proto"

// MockEntity is v as keystone returns it from Retrieve, Find or QueryIndex, for a
// MockServer to answer with: its ID, Exists, an active state, and its properties marshalled
// as Mutate sends them. An entity retrieved without Exists and the active state reports
// ExistsInKeystone and IsStateActive as false.
func MockEntity(v Entity) (*proto.EntityResponse, error) {
	props, err := Marshal(v)
	if err != nil {
		return nil, err
	}
	exists := true
	resp := &proto.EntityResponse{
		Entity: &proto.Entity{EntityId: v.GetKeystoneID().String(), State: proto.EntityState_Active},
		Exists: &exists,
	}
	for p, val := range props {
		resp.Properties = append(resp.Properties, &proto.EntityProperty{Property: p.Name(), Value: val})
	}
	return resp, nil
}
