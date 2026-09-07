package event_stream

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
)

const delayedNak = 500 * time.Millisecond

var errStreamComplete = errors.New("event stream acknowledgement test complete")

type Requirement struct {
	eventType string
}

func (d *Requirement) Name() string {
	return "Event Stream"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.User{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor, report requirements.Reporter) {
	report(d.createEvent(actor))
	if d.eventType != "" {
		report(d.acknowledgementLifecycle(actor))
	}
}

func (d *Requirement) createEvent(actor *keystone.Actor) requirements.TestResult {
	result := requirements.TestResult{Name: "Create Stream Event"}
	d.eventType = "event-stream-" + uuid.NewString()

	user := &models.User{
		ExternalID: "event-stream-" + uuid.NewString(),
		Validate:   "event-stream-acknowledgements",
	}
	user.AddEvent(d.eventType, map[string]string{"test": "acknowledgements"})
	if err := actor.Mutate(context.Background(), user, keystone.WithMutationComment("Create event stream acknowledgement test event")); err != nil {
		d.eventType = ""
		return result.WithError(err)
	}
	return result
}

func (d *Requirement) acknowledgementLifecycle(actor *keystone.Actor) requirements.TestResult {
	result := requirements.TestResult{Name: "ACK, NAK, Delayed NAK, In Progress, and Delivery Attempts"}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	deliveries := 0
	messageID := ""
	delayedAt := time.Time{}
	err := actor.EventStreamWithAck(ctx, func(message *keystone.EventStreamMessage) error {
		if message.GetMessageId() == "" {
			return errors.New("event stream response did not include a message ID")
		}
		if messageID == "" {
			messageID = message.GetMessageId()
		} else if message.GetMessageId() != messageID {
			return fmt.Errorf("redelivery message ID changed from %q to %q", messageID, message.GetMessageId())
		}

		deliveries++
		if want := uint64(deliveries); message.GetDeliveryAttempts() != want {
			return fmt.Errorf("delivery %d reported %d delivery attempts, expected %d", deliveries, message.GetDeliveryAttempts(), want)
		}
		switch deliveries {
		case 1:
			if err := message.InProgress(); err != nil {
				return fmt.Errorf("report in progress: %w", err)
			}
			if err := message.Nak(); err != nil {
				return fmt.Errorf("NAK: %w", err)
			}
			return nil
		case 2:
			delayedAt = time.Now()
			if err := message.NakWithDelay(delayedNak); err != nil {
				return fmt.Errorf("delayed NAK: %w", err)
			}
			return nil
		case 3:
			if elapsed := time.Since(delayedAt); elapsed < delayedNak/2 {
				return fmt.Errorf("delayed NAK redelivered after %s, expected approximately %s", elapsed, delayedNak)
			}
			if err := message.Ack(); err != nil {
				return fmt.Errorf("ACK: %w", err)
			}
			return errStreamComplete
		default:
			return fmt.Errorf("received unexpected delivery %d", deliveries)
		}
	}, "acknowledgements-"+uuid.NewString(), keystone.OwnKey(d.eventType))

	if errors.Is(err, errStreamComplete) {
		return result
	}
	if err == nil {
		return result.WithError(fmt.Errorf("event stream ended after %d deliveries, expected 3", deliveries))
	}
	return result.WithError(err)
}
