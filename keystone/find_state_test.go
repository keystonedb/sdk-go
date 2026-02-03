package keystone

import (
	"testing"

	"github.com/keystonedb/sdk-go/proto"
)

func TestWithStates(t *testing.T) {
	tests := []struct {
		name           string
		states         []proto.EntityState
		expectedValues int
	}{
		{"Single state", []proto.EntityState{proto.EntityState_Active}, 1},
		{"Two states", []proto.EntityState{proto.EntityState_Active, proto.EntityState_Archived}, 2},
		{"All four states", []proto.EntityState{proto.EntityState_Active, proto.EntityState_Offline, proto.EntityState_Corrupt, proto.EntityState_Archived}, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := WithStates(tt.states...)
			if opt == nil {
				t.Fatal("Expected non-nil option")
			}

			req := &filterRequest{}
			opt.Apply(req)

			if len(req.Filters) != 1 {
				t.Fatalf("Expected 1 filter, got %d", len(req.Filters))
			}

			filter := req.Filters[0]
			if filter.Property != statePropertyName {
				t.Errorf("Expected property '%s', got '%s'", statePropertyName, filter.Property)
			}
			if filter.Operator != proto.Operator_In {
				t.Errorf("Expected In operator, got %v", filter.Operator)
			}
			if len(filter.Values) != tt.expectedValues {
				t.Errorf("Expected %d values, got %d", tt.expectedValues, len(filter.Values))
			}

			// Verify the state values
			for i, state := range tt.states {
				if filter.Values[i].Int != int64(state) {
					t.Errorf("Expected state value %d, got %d", int64(state), filter.Values[i].Int)
				}
			}
		})
	}
}

func TestWithStatesEmpty(t *testing.T) {
	opt := WithStates()
	if opt != nil {
		t.Error("Expected nil option for empty states")
	}
}

func TestIncludeArchived(t *testing.T) {
	opt := IncludeArchived()
	req := &filterRequest{}
	opt.Apply(req)

	if len(req.Filters) != 1 {
		t.Fatalf("Expected 1 filter, got %d", len(req.Filters))
	}

	filter := req.Filters[0]
	if len(filter.Values) != 2 {
		t.Fatalf("Expected 2 values, got %d", len(filter.Values))
	}

	// Should contain Active and Archived
	hasActive := false
	hasArchived := false
	for _, v := range filter.Values {
		if v.Int == int64(proto.EntityState_Active) {
			hasActive = true
		}
		if v.Int == int64(proto.EntityState_Archived) {
			hasArchived = true
		}
	}
	if !hasActive {
		t.Error("Expected Active state in IncludeArchived")
	}
	if !hasArchived {
		t.Error("Expected Archived state in IncludeArchived")
	}
}

func TestOnlyArchived(t *testing.T) {
	opt := OnlyArchived()
	req := &filterRequest{}
	opt.Apply(req)

	if len(req.Filters) != 1 {
		t.Fatalf("Expected 1 filter, got %d", len(req.Filters))
	}

	filter := req.Filters[0]
	if len(filter.Values) != 1 {
		t.Fatalf("Expected 1 value, got %d", len(filter.Values))
	}
	if filter.Values[0].Int != int64(proto.EntityState_Archived) {
		t.Errorf("Expected Archived state, got %d", filter.Values[0].Int)
	}
}

func TestOnlyActive(t *testing.T) {
	opt := OnlyActive()
	req := &filterRequest{}
	opt.Apply(req)

	if len(req.Filters) != 1 {
		t.Fatalf("Expected 1 filter, got %d", len(req.Filters))
	}

	filter := req.Filters[0]
	if len(filter.Values) != 1 {
		t.Fatalf("Expected 1 value, got %d", len(filter.Values))
	}
	if filter.Values[0].Int != int64(proto.EntityState_Active) {
		t.Errorf("Expected Active state, got %d", filter.Values[0].Int)
	}
}

func TestIncludeOffline(t *testing.T) {
	opt := IncludeOffline()
	req := &filterRequest{}
	opt.Apply(req)

	filter := req.Filters[0]
	if len(filter.Values) != 2 {
		t.Fatalf("Expected 2 values, got %d", len(filter.Values))
	}

	hasActive := false
	hasOffline := false
	for _, v := range filter.Values {
		if v.Int == int64(proto.EntityState_Active) {
			hasActive = true
		}
		if v.Int == int64(proto.EntityState_Offline) {
			hasOffline = true
		}
	}
	if !hasActive || !hasOffline {
		t.Error("Expected Active and Offline states in IncludeOffline")
	}
}

func TestOnlyOffline(t *testing.T) {
	opt := OnlyOffline()
	req := &filterRequest{}
	opt.Apply(req)

	filter := req.Filters[0]
	if len(filter.Values) != 1 {
		t.Fatalf("Expected 1 value, got %d", len(filter.Values))
	}
	if filter.Values[0].Int != int64(proto.EntityState_Offline) {
		t.Errorf("Expected Offline state, got %d", filter.Values[0].Int)
	}
}

func TestIncludeCorrupt(t *testing.T) {
	opt := IncludeCorrupt()
	req := &filterRequest{}
	opt.Apply(req)

	filter := req.Filters[0]
	if len(filter.Values) != 2 {
		t.Fatalf("Expected 2 values, got %d", len(filter.Values))
	}

	hasActive := false
	hasCorrupt := false
	for _, v := range filter.Values {
		if v.Int == int64(proto.EntityState_Active) {
			hasActive = true
		}
		if v.Int == int64(proto.EntityState_Corrupt) {
			hasCorrupt = true
		}
	}
	if !hasActive || !hasCorrupt {
		t.Error("Expected Active and Corrupt states in IncludeCorrupt")
	}
}

func TestOnlyCorrupt(t *testing.T) {
	opt := OnlyCorrupt()
	req := &filterRequest{}
	opt.Apply(req)

	filter := req.Filters[0]
	if len(filter.Values) != 1 {
		t.Fatalf("Expected 1 value, got %d", len(filter.Values))
	}
	if filter.Values[0].Int != int64(proto.EntityState_Corrupt) {
		t.Errorf("Expected Corrupt state, got %d", filter.Values[0].Int)
	}
}

func TestAllStates(t *testing.T) {
	opt := AllStates()
	req := &filterRequest{}
	opt.Apply(req)

	filter := req.Filters[0]
	if len(filter.Values) != 4 {
		t.Fatalf("Expected 4 values, got %d", len(filter.Values))
	}

	expectedStates := map[int64]bool{
		int64(proto.EntityState_Active):   false,
		int64(proto.EntityState_Offline):  false,
		int64(proto.EntityState_Corrupt):  false,
		int64(proto.EntityState_Archived): false,
	}

	for _, v := range filter.Values {
		if _, ok := expectedStates[v.Int]; ok {
			expectedStates[v.Int] = true
		} else {
			t.Errorf("Unexpected state value %d in AllStates", v.Int)
		}
	}

	for state, found := range expectedStates {
		if !found {
			t.Errorf("Expected state %d not found in AllStates", state)
		}
	}
}

func TestStateFilterImplementsFindOption(t *testing.T) {
	var _ FindOption = WithStates(proto.EntityState_Active)
	var _ FindOption = IncludeArchived()
	var _ FindOption = OnlyArchived()
	var _ FindOption = OnlyActive()
	var _ FindOption = IncludeOffline()
	var _ FindOption = OnlyOffline()
	var _ FindOption = IncludeCorrupt()
	var _ FindOption = OnlyCorrupt()
	var _ FindOption = AllStates()
}
