package keystone

import (
	"github.com/keystonedb/sdk-go/proto"
	"testing"
)

func TestPropertyLoader(t *testing.T) {
	l := WithProperty(true, "name", "email")
	view := &proto.EntityView{}

	l.Apply(view)
	if len(view.Properties) != 1 {
		t.Error("Expected 1 Property request")
	}

	req0 := view.Properties[0]
	if len(req0.GetProperties()) != 2 {
		t.Error("Expected 2 properties")
	}

	if req0.GetProperties()[0] != "name" {
		t.Error("Expected name Property")
	}
}

func TestRetrieveOptions(t *testing.T) {
	l := RetrieveOptions(WithProperty(true, "name", "email"))
	view := &proto.EntityView{}

	l.Apply(view)
	if len(view.Properties) != 1 {
		t.Error("Expected 1 Property request")
	}
}
