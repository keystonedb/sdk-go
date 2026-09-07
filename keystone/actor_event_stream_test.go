package keystone

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/grpc"
)

func newEventStreamTestActor(t *testing.T) (*Actor, *MockServer, func()) {
	t.Helper()
	conn, mock, listener, server := MockConnection()
	go func() { _ = server.Serve(listener) }()
	actor := conn.Actor("workspace-1", "127.0.0.1", "user-1", "go-test")
	return &actor, mock, func() {
		server.Stop()
		_ = listener.Close()
	}
}

func TestActorEventStreamAcknowledgements(t *testing.T) {
	handlerErr := errors.New("handler failed")
	tests := []struct {
		name        string
		handler     func(*EventStreamMessage) error
		wantActions []proto.EventStreamAck_Action
		wantDelay   time.Duration
		wantErr     error
	}{
		{
			name:        "automatic ack",
			handler:     func(*EventStreamMessage) error { return nil },
			wantActions: []proto.EventStreamAck_Action{proto.EventStreamAck_EVENT_STREAM_ACK},
		},
		{
			name: "explicit ack",
			handler: func(message *EventStreamMessage) error {
				return message.Ack()
			},
			wantActions: []proto.EventStreamAck_Action{proto.EventStreamAck_EVENT_STREAM_ACK},
		},
		{
			name:        "automatic nak on handler error",
			handler:     func(*EventStreamMessage) error { return handlerErr },
			wantActions: []proto.EventStreamAck_Action{proto.EventStreamAck_EVENT_STREAM_NAK},
			wantErr:     handlerErr,
		},
		{
			name: "explicit nak",
			handler: func(message *EventStreamMessage) error {
				return message.Nak()
			},
			wantActions: []proto.EventStreamAck_Action{proto.EventStreamAck_EVENT_STREAM_NAK},
		},
		{
			name: "delayed nak",
			handler: func(message *EventStreamMessage) error {
				return message.NakWithDelay(15 * time.Second)
			},
			wantActions: []proto.EventStreamAck_Action{proto.EventStreamAck_EVENT_STREAM_NAK_WITH_DELAY},
			wantDelay:   15 * time.Second,
		},
		{
			name: "in progress then automatic ack",
			handler: func(message *EventStreamMessage) error {
				return message.InProgress()
			},
			wantActions: []proto.EventStreamAck_Action{
				proto.EventStreamAck_EVENT_STREAM_IN_PROGRESS,
				proto.EventStreamAck_EVENT_STREAM_ACK,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actor, mock, cleanup := newEventStreamTestActor(t)
			defer cleanup()

			var initial *proto.EventStreamRequest
			callbacks := make(chan *proto.EventStreamAck, len(tt.wantActions))
			mock.EventStreamFunc = func(stream grpc.BidiStreamingServer[proto.EventStreamRequest, proto.EventStreamResponse]) error {
				var err error
				initial, err = stream.Recv()
				if err != nil {
					return err
				}
				if err = stream.Send(&proto.EventStreamResponse{MessageId: "message-42", Eid: "entity-1"}); err != nil {
					return err
				}
				for range tt.wantActions {
					req, recvErr := stream.Recv()
					if recvErr != nil {
						return recvErr
					}
					callbacks <- req.GetAck()
				}
				return nil
			}

			err := actor.EventStreamWithAck(context.Background(), tt.handler, "orders", OwnKey("updated"))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("EventStreamWithAck() error = %v, want %v", err, tt.wantErr)
			}
			if initial.GetStreamName() != "orders" || initial.GetAuthorization().GetWorkspaceId() != "workspace-1" {
				t.Fatalf("initial stream request not forwarded: %+v", initial)
			}
			if initial.GetEventType().GetKey() != "updated" {
				t.Fatalf("event type = %q, want updated", initial.GetEventType().GetKey())
			}
			got := make([]*proto.EventStreamAck, 0, len(tt.wantActions))
			for range tt.wantActions {
				select {
				case ack := <-callbacks:
					got = append(got, ack)
				case <-time.After(time.Second):
					t.Fatalf("received %d callbacks, want %d", len(got), len(tt.wantActions))
				}
			}
			for i, want := range tt.wantActions {
				if got[i].GetMessageId() != "message-42" || got[i].GetAction() != want {
					t.Errorf("callback %d = %+v, want message-42/%s", i, got[i], want)
				}
			}
			if tt.wantDelay > 0 && got[0].GetDelay().AsDuration() != tt.wantDelay {
				t.Errorf("delay = %v, want %v", got[0].GetDelay().AsDuration(), tt.wantDelay)
			}
		})
	}
}

func TestActorEventStreamLegacyHandlerAcknowledges(t *testing.T) {
	actor, mock, cleanup := newEventStreamTestActor(t)
	defer cleanup()

	mock.EventStreamFunc = func(stream grpc.BidiStreamingServer[proto.EventStreamRequest, proto.EventStreamResponse]) error {
		if _, err := stream.Recv(); err != nil {
			return err
		}
		if err := stream.Send(&proto.EventStreamResponse{MessageId: "legacy-message"}); err != nil {
			return err
		}
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		if req.GetAck().GetAction() != proto.EventStreamAck_EVENT_STREAM_ACK {
			t.Errorf("legacy callback action = %s, want ACK", req.GetAck().GetAction())
		}
		return nil
	}

	if err := actor.EventStream(context.Background(), func(*proto.EventStreamResponse) error { return nil }, "legacy", nil); err != nil {
		t.Fatalf("EventStream() error = %v", err)
	}
}

func TestEventStreamMessageValidation(t *testing.T) {
	message := &EventStreamMessage{}
	if err := message.NakWithDelay(0); err == nil {
		t.Fatal("NakWithDelay(0) returned nil error")
	}
	if err := message.NakWithDelay(-time.Second); err == nil {
		t.Fatal("NakWithDelay(-1s) returned nil error")
	}
	if err := message.Ack(); err == nil {
		t.Fatal("Ack() on invalid message returned nil error")
	}
}

func TestActorEventStreamRejectsNilInputs(t *testing.T) {
	var actor *Actor
	if err := actor.EventStreamWithAck(context.Background(), func(*EventStreamMessage) error { return nil }, "events", nil); err == nil {
		t.Fatal("nil actor returned nil error")
	}

	actor, _, cleanup := newEventStreamTestActor(t)
	defer cleanup()
	if err := actor.EventStreamWithAck(context.Background(), nil, "events", nil); err == nil {
		t.Fatal("nil handler returned nil error")
	}
}
