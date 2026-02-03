package keystone

import "github.com/keystonedb/sdk-go/proto"

const statePropertyName = "_state"

// WithStates filters entities to only return those with the specified states.
// By default, only Active entities are returned. Use this to include other states.
// Example: WithStates(proto.EntityState_Active, proto.EntityState_Archived)
func WithStates(states ...proto.EntityState) FindOption {
	if len(states) == 0 {
		return nil
	}
	values := make([]*proto.Value, len(states))
	for i, state := range states {
		values[i] = &proto.Value{Int: int64(state)}
	}
	return propertyFilter{key: statePropertyName, values: values, operator: proto.Operator_In}
}

// IncludeArchived returns entities with Active or Archived state.
// Use this when you want to see archived entities alongside active ones.
func IncludeArchived() FindOption {
	return WithStates(proto.EntityState_Active, proto.EntityState_Archived)
}

// OnlyArchived returns only entities with Archived state.
func OnlyArchived() FindOption {
	return WithStates(proto.EntityState_Archived)
}

// OnlyActive returns only entities with Active state.
// This is the default behavior, but can be used explicitly.
func OnlyActive() FindOption {
	return WithStates(proto.EntityState_Active)
}

// IncludeOffline returns entities with Active or Offline state.
func IncludeOffline() FindOption {
	return WithStates(proto.EntityState_Active, proto.EntityState_Offline)
}

// OnlyOffline returns only entities with Offline state.
func OnlyOffline() FindOption {
	return WithStates(proto.EntityState_Offline)
}

// IncludeCorrupt returns entities with Active or Corrupt state.
func IncludeCorrupt() FindOption {
	return WithStates(proto.EntityState_Active, proto.EntityState_Corrupt)
}

// OnlyCorrupt returns only entities with Corrupt state.
func OnlyCorrupt() FindOption {
	return WithStates(proto.EntityState_Corrupt)
}

// AllStates returns entities in any state (Active, Offline, Corrupt, Archived).
// Does not include Invalid or Removed states.
func AllStates() FindOption {
	return WithStates(
		proto.EntityState_Active,
		proto.EntityState_Offline,
		proto.EntityState_Corrupt,
		proto.EntityState_Archived,
	)
}
