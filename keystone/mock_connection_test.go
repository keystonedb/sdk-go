package keystone

import (
	"context"
	"testing"
	"time"

	"github.com/keystonedb/sdk-go/proto"
)

func TestMockConnectionsKeepTheirOwnListeners(t *testing.T) {
	first, firstMock, firstListener, firstServer := MockConnection()
	defer firstServer.Stop()
	defer firstListener.Close()
	second, secondMock, secondListener, secondServer := MockConnection()
	defer secondServer.Stop()
	defer secondListener.Close()
	firstMock.FeedReadFunc = func(context.Context, *proto.FeedReadRequest) (*proto.FeedReadResponse, error) {
		return &proto.FeedReadResponse{Entries: []*proto.FeedEntry{{EntityId: "first"}}}, nil
	}
	secondMock.FeedReadFunc = func(context.Context, *proto.FeedReadRequest) (*proto.FeedReadResponse, error) {
		return &proto.FeedReadResponse{Entries: []*proto.FeedEntry{{EntityId: "second"}}}, nil
	}
	go firstServer.Serve(firstListener)
	go secondServer.Serve(secondListener)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	for _, test := range []struct {
		connection *Connection
		expected   string
	}{{first, "first"}, {second, "second"}} {
		response, err := test.connection.FeedRead(ctx, &proto.FeedReadRequest{})
		if err != nil || len(response.GetEntries()) != 1 || response.Entries[0].EntityId != test.expected {
			t.Fatalf("mock connection routed to another listener: %v %v", response, err)
		}
	}
}
