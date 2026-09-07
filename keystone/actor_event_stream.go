package keystone

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/durationpb"
)

type Key struct {
	VendorID string
	AppID    string
	Type     string
}

func NewKey(vendorID, appID, entityType string) *Key {
	return &Key{
		VendorID: vendorID,
		AppID:    appID,
		Type:     entityType,
	}
}

func OwnKey(key string) *Key {
	return NewKey("", "", key)
}

func (k *Key) toProto(a *Actor) *proto.Key {
	if k == nil {
		return nil
	}
	ret := &proto.Key{
		Source: &proto.VendorApp{
			VendorId: k.VendorID,
			AppId:    k.AppID,
		},
		Key: k.Type,
	}

	if k.VendorID == "" {
		ret.Source = a.Authorization().GetSource()
	} else if k.AppID == "" {
		ret.Source.AppId = a.Authorization().GetSource().GetAppId()
	}

	return ret
}

// EventStreamMessage is an event-stream response with acknowledgement controls.
// DeliveryAttempts and GetDeliveryAttempts expose how many times the server has
// delivered the message. InProgress may be called any number of times before one
// terminal call to Ack, Nak, or NakWithDelay.
type EventStreamMessage struct {
	*proto.EventStreamResponse

	sender  *eventStreamSender
	mu      sync.Mutex
	settled bool
}

type eventStreamSender struct {
	stream grpc.BidiStreamingClient[proto.EventStreamRequest, proto.EventStreamResponse]
	mu     sync.Mutex
}

func (s *eventStreamSender) send(ack *proto.EventStreamAck) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stream.Send(&proto.EventStreamRequest{Ack: ack})
}

// Ack marks the event as successfully processed.
func (m *EventStreamMessage) Ack() error {
	return m.ack(proto.EventStreamAck_EVENT_STREAM_ACK, 0, true)
}

// Nak marks the event as unsuccessfully processed and makes it available for redelivery.
func (m *EventStreamMessage) Nak() error {
	return m.ack(proto.EventStreamAck_EVENT_STREAM_NAK, 0, true)
}

// NakWithDelay marks the event as unsuccessfully processed and requests
// redelivery after delay.
func (m *EventStreamMessage) NakWithDelay(delay time.Duration) error {
	if delay <= 0 {
		return errors.New("event stream NAK delay must be greater than zero")
	}
	return m.ack(proto.EventStreamAck_EVENT_STREAM_NAK_WITH_DELAY, delay, true)
}

// InProgress tells the server that processing is still active, extending the
// acknowledgement deadline without settling the event.
func (m *EventStreamMessage) InProgress() error {
	return m.ack(proto.EventStreamAck_EVENT_STREAM_IN_PROGRESS, 0, false)
}

func (m *EventStreamMessage) ack(action proto.EventStreamAck_Action, delay time.Duration, terminal bool) error {
	if m == nil || m.EventStreamResponse == nil || m.sender == nil {
		return errors.New("invalid event stream message")
	}
	if m.GetMessageId() == "" {
		return errors.New("event stream response has no message ID")
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.settled {
		return errors.New("event stream message is already settled")
	}

	ack := &proto.EventStreamAck{MessageId: m.GetMessageId(), Action: action}
	if delay > 0 {
		ack.Delay = durationpb.New(delay)
	}
	if err := m.sender.send(ack); err != nil {
		return err
	}
	m.settled = terminal
	return nil
}

func (m *EventStreamMessage) isSettled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.settled
}

// EventStream consumes events and automatically ACKs successful handler calls.
// When the handler returns an error, the event is NAKed before that error is returned.
func (a *Actor) EventStream(ctx context.Context, handler func(response *proto.EventStreamResponse) error, name string, eventType *Key) error {
	if handler == nil {
		return errors.New("event stream handler is nil")
	}
	return a.EventStreamWithAck(ctx, func(message *EventStreamMessage) error {
		return handler(message.EventStreamResponse)
	}, name, eventType)
}

// EventStreamWithAck consumes events and gives the handler explicit control of
// acknowledgements. A handler that does not settle a message is automatically
// ACKed on nil error or NAKed on non-nil error.
func (a *Actor) EventStreamWithAck(ctx context.Context, handler func(message *EventStreamMessage) error, name string, eventType *Key) error {
	if a == nil || a.Connection() == nil {
		return errors.New("actor connection is nil")
	}
	if handler == nil {
		return errors.New("event stream handler is nil")
	}

	req := &proto.EventStreamRequest{
		Authorization: a.Authorization(),
		StreamName:    name,
		AllWorkspaces: a.WorkspaceID() == noWorkspace,
		EventType:     eventType.toProto(a),
	}

	stream, err := a.Connection().EventStream(ctx, req)
	if err != nil {
		return err
	}
	defer func() { _ = stream.CloseSend() }()

	sender := &eventStreamSender{stream: stream}

	for {
		evt, recErr := stream.Recv()
		if recErr == io.EOF {
			return nil
		}
		if recErr != nil {
			return recErr
		}

		message := &EventStreamMessage{EventStreamResponse: evt, sender: sender}
		handleErr := handler(message)
		if message.isSettled() {
			if handleErr != nil {
				return handleErr
			}
			continue
		}

		if handleErr != nil {
			if nakErr := message.Nak(); nakErr != nil {
				return fmt.Errorf("event stream handler failed: %w; NAK failed: %v", handleErr, nakErr)
			}
			return handleErr
		}
		if ackErr := message.Ack(); ackErr != nil {
			return ackErr
		}
	}
}
