package keystone

import (
	"testing"
	"time"

	"github.com/keystonedb/sdk-go/proto"
)

func TestWithAckWait(t *testing.T) {
	req := &proto.EventStreamRequest{}

	WithAckWait(90 * time.Second)(req)

	if req.GetAckWait() == nil {
		t.Fatal("expected the deadline to be carried on the subscription")
	}
	if got := req.GetAckWait().AsDuration(); got != 90*time.Second {
		t.Errorf("expected 90s, got %s", got)
	}
}

func TestEventStreamOptionsAreOptional(t *testing.T) {
	req := &proto.EventStreamRequest{StreamName: "external-purchase"}

	for _, opt := range []EventStreamOption{nil} {
		if opt != nil {
			opt(req)
		}
	}

	if req.GetAckWait() != nil {
		t.Errorf("expected no deadline to be set, got %s", req.GetAckWait().AsDuration())
	}
	if req.GetStreamName() != "external-purchase" {
		t.Error("expected the rest of the subscription to be untouched")
	}
}
