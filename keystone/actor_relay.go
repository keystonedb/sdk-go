package keystone

import (
	"context"
	"errors"

	"github.com/keystonedb/sdk-go/proto"
)

// RelayCreateSession opens a new relay session for the current workspace. ttlMs is clamped by the server
// to [60_000, 604_800_000] (60s..7d). metadata is an optional opaque JSON blob (<=1 KB). If idempotencyKey
// is non-empty, the server uses it as the session_id, making retries safe.
func (a *Actor) RelayCreateSession(ctx context.Context, ttlMs int64, metadata []byte, idempotencyKey string) (*proto.RelayCreateSessionResponse, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}
	return a.connection.RelayCreateSession(ctx, &proto.RelayCreateSessionRequest{
		Authorization:  a.Authorization(),
		TtlMs:          ttlMs,
		Metadata:       metadata,
		IdempotencyKey: idempotencyKey,
	})
}

// RelayExtendSession applies a monotonic extension to an existing session. The server sets
// expires_at = max(current_expires_at, min(now+extendByMs, now+7d)); an extend request never
// shortens a session.
func (a *Actor) RelayExtendSession(ctx context.Context, sessionID string, extendByMs int64) (*proto.RelayExtendSessionResponse, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}
	return a.connection.RelayExtendSession(ctx, &proto.RelayExtendSessionRequest{
		Authorization: a.Authorization(),
		SessionId:     sessionID,
		ExtendByMs:    extendByMs,
	})
}

// RelayDestroySession terminates a session immediately. The reason is surfaced to connected
// devices in the system.closing frame.
func (a *Actor) RelayDestroySession(ctx context.Context, sessionID, reason string) error {
	if a == nil || a.connection == nil {
		return errors.New("actor or connection is nil")
	}
	_, err := a.connection.RelayDestroySession(ctx, &proto.RelayDestroySessionRequest{
		Authorization: a.Authorization(),
		SessionId:     sessionID,
		Reason:        reason,
	})
	return err
}

// RelayCreateShortCode mints a human-typable short code (AAA-BBB-CCC) that resolves back to the
// session. ttlMs is clamped by the server to [60_000, 7_200_000] (60s..2h). Codes are vendor-scoped.
func (a *Actor) RelayCreateShortCode(ctx context.Context, sessionID string, ttlMs int64) (*proto.RelayCreateShortCodeResponse, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}
	return a.connection.RelayCreateShortCode(ctx, &proto.RelayCreateShortCodeRequest{
		Authorization: a.Authorization(),
		SessionId:     sessionID,
		TtlMs:         ttlMs,
	})
}

// RelayResolveShortCode looks up a previously minted short code and returns the associated session.
// The caller must be under the same vendor as the code's creator and the session's workspace.
func (a *Actor) RelayResolveShortCode(ctx context.Context, code string) (*proto.RelayResolveShortCodeResponse, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}
	return a.connection.RelayResolveShortCode(ctx, &proto.RelayResolveShortCodeRequest{
		Authorization: a.Authorization(),
		Code:          code,
	})
}

// RelayDeleteShortCode removes a short code (typically after one-shot consumption).
func (a *Actor) RelayDeleteShortCode(ctx context.Context, code string) error {
	if a == nil || a.connection == nil {
		return errors.New("actor or connection is nil")
	}
	_, err := a.connection.RelayDeleteShortCode(ctx, &proto.RelayDeleteShortCodeRequest{
		Authorization: a.Authorization(),
		Code:          code,
	})
	return err
}

// RelayPublish publishes a message envelope into the session's channel. Every connected device
// receives the envelope. data is an opaque JSON payload and must keep the assembled envelope
// under 5 KB. If originatorUser is empty, the authorization's user is used.
func (a *Actor) RelayPublish(ctx context.Context, sessionID, eventType string, data []byte, originatorUser string) (*proto.RelayPublishResponse, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}
	return a.connection.RelayPublish(ctx, &proto.RelayPublishRequest{
		Authorization:  a.Authorization(),
		SessionId:      sessionID,
		Type:           eventType,
		Data:           data,
		OriginatorUser: originatorUser,
	})
}

// RelayGetPresence returns the list of devices currently connected to the session.
func (a *Actor) RelayGetPresence(ctx context.Context, sessionID string) ([]*proto.PresenceDevice, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}
	resp, err := a.connection.RelayGetPresence(ctx, &proto.RelayGetPresenceRequest{
		Authorization: a.Authorization(),
		SessionId:     sessionID,
	})
	if err != nil {
		return nil, err
	}
	return resp.GetDevices(), nil
}
