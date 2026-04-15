package keystone

import (
	"context"
	"errors"
	"testing"

	"github.com/keystonedb/sdk-go/proto"
)

// --- Nil-actor / nil-connection guards ---

func TestActor_RelayCreateSession_NilActor(t *testing.T) {
	var actor *Actor
	_, err := actor.RelayCreateSession(context.Background(), 60_000, nil, "")
	if err == nil || err.Error() != "actor or connection is nil" {
		t.Errorf("expected 'actor or connection is nil'; got %v", err)
	}
}

func TestActor_RelayCreateSession_NilConnection(t *testing.T) {
	actor := &Actor{}
	_, err := actor.RelayCreateSession(context.Background(), 60_000, nil, "")
	if err == nil || err.Error() != "actor or connection is nil" {
		t.Errorf("expected 'actor or connection is nil'; got %v", err)
	}
}

func TestActor_RelayExtendSession_NilActor(t *testing.T) {
	var actor *Actor
	_, err := actor.RelayExtendSession(context.Background(), "sid", 60_000)
	if err == nil || err.Error() != "actor or connection is nil" {
		t.Errorf("expected 'actor or connection is nil'; got %v", err)
	}
}

func TestActor_RelayExtendSession_NilConnection(t *testing.T) {
	actor := &Actor{}
	_, err := actor.RelayExtendSession(context.Background(), "sid", 60_000)
	if err == nil || err.Error() != "actor or connection is nil" {
		t.Errorf("expected 'actor or connection is nil'; got %v", err)
	}
}

func TestActor_RelayDestroySession_NilActor(t *testing.T) {
	var actor *Actor
	err := actor.RelayDestroySession(context.Background(), "sid", "test")
	if err == nil || err.Error() != "actor or connection is nil" {
		t.Errorf("expected 'actor or connection is nil'; got %v", err)
	}
}

func TestActor_RelayDestroySession_NilConnection(t *testing.T) {
	actor := &Actor{}
	err := actor.RelayDestroySession(context.Background(), "sid", "test")
	if err == nil || err.Error() != "actor or connection is nil" {
		t.Errorf("expected 'actor or connection is nil'; got %v", err)
	}
}

func TestActor_RelayCreateShortCode_NilActor(t *testing.T) {
	var actor *Actor
	_, err := actor.RelayCreateShortCode(context.Background(), "sid", 60_000)
	if err == nil || err.Error() != "actor or connection is nil" {
		t.Errorf("expected 'actor or connection is nil'; got %v", err)
	}
}

func TestActor_RelayCreateShortCode_NilConnection(t *testing.T) {
	actor := &Actor{}
	_, err := actor.RelayCreateShortCode(context.Background(), "sid", 60_000)
	if err == nil || err.Error() != "actor or connection is nil" {
		t.Errorf("expected 'actor or connection is nil'; got %v", err)
	}
}

func TestActor_RelayResolveShortCode_NilActor(t *testing.T) {
	var actor *Actor
	_, err := actor.RelayResolveShortCode(context.Background(), "AAA-BBB-CCC")
	if err == nil || err.Error() != "actor or connection is nil" {
		t.Errorf("expected 'actor or connection is nil'; got %v", err)
	}
}

func TestActor_RelayResolveShortCode_NilConnection(t *testing.T) {
	actor := &Actor{}
	_, err := actor.RelayResolveShortCode(context.Background(), "AAA-BBB-CCC")
	if err == nil || err.Error() != "actor or connection is nil" {
		t.Errorf("expected 'actor or connection is nil'; got %v", err)
	}
}

func TestActor_RelayDeleteShortCode_NilActor(t *testing.T) {
	var actor *Actor
	err := actor.RelayDeleteShortCode(context.Background(), "AAA-BBB-CCC")
	if err == nil || err.Error() != "actor or connection is nil" {
		t.Errorf("expected 'actor or connection is nil'; got %v", err)
	}
}

func TestActor_RelayDeleteShortCode_NilConnection(t *testing.T) {
	actor := &Actor{}
	err := actor.RelayDeleteShortCode(context.Background(), "AAA-BBB-CCC")
	if err == nil || err.Error() != "actor or connection is nil" {
		t.Errorf("expected 'actor or connection is nil'; got %v", err)
	}
}

func TestActor_RelayPublish_NilActor(t *testing.T) {
	var actor *Actor
	_, err := actor.RelayPublish(context.Background(), "sid", "cart.updated", []byte(`{}`), "")
	if err == nil || err.Error() != "actor or connection is nil" {
		t.Errorf("expected 'actor or connection is nil'; got %v", err)
	}
}

func TestActor_RelayPublish_NilConnection(t *testing.T) {
	actor := &Actor{}
	_, err := actor.RelayPublish(context.Background(), "sid", "cart.updated", []byte(`{}`), "")
	if err == nil || err.Error() != "actor or connection is nil" {
		t.Errorf("expected 'actor or connection is nil'; got %v", err)
	}
}

func TestActor_RelayGetPresence_NilActor(t *testing.T) {
	var actor *Actor
	_, err := actor.RelayGetPresence(context.Background(), "sid")
	if err == nil || err.Error() != "actor or connection is nil" {
		t.Errorf("expected 'actor or connection is nil'; got %v", err)
	}
}

func TestActor_RelayGetPresence_NilConnection(t *testing.T) {
	actor := &Actor{}
	_, err := actor.RelayGetPresence(context.Background(), "sid")
	if err == nil || err.Error() != "actor or connection is nil" {
		t.Errorf("expected 'actor or connection is nil'; got %v", err)
	}
}

// --- Mock-backed happy-path / request-shape tests ---

func newRelayTestActor(t *testing.T) (*Actor, *MockServer, func()) {
	t.Helper()
	conn, mock, _, server := MockConnection()
	go func() { _ = server.Serve(mockListener) }()
	actor := conn.Actor("ws-1", "127.0.0.1", "user-1", "go-test")
	return &actor, mock, func() {
		server.Stop()
		_ = mockListener.Close()
	}
}

func TestActor_RelayCreateSession_ForwardsFields(t *testing.T) {
	actor, mock, cleanup := newRelayTestActor(t)
	defer cleanup()

	var got *proto.RelayCreateSessionRequest
	mock.RelayCreateSessionFunc = func(_ context.Context, req *proto.RelayCreateSessionRequest) (*proto.RelayCreateSessionResponse, error) {
		got = req
		return &proto.RelayCreateSessionResponse{
			SessionId:     "sid-123",
			WorkspaceSlug: "ABCDEFGH",
			ExpiresAtMs:   1_700_000_000_000,
			ServerTsMs:    1_699_999_000_000,
		}, nil
	}

	meta := []byte(`{"hello":"world"}`)
	resp, err := actor.RelayCreateSession(context.Background(), 30*60*1000, meta, "idem-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("handler not invoked")
	}
	if got.GetAuthorization() == nil {
		t.Error("authorization not forwarded")
	}
	if got.GetTtlMs() != 30*60*1000 {
		t.Errorf("TtlMs = %d; want %d", got.GetTtlMs(), 30*60*1000)
	}
	if string(got.GetMetadata()) != string(meta) {
		t.Errorf("Metadata = %q; want %q", got.GetMetadata(), meta)
	}
	if got.GetIdempotencyKey() != "idem-key" {
		t.Errorf("IdempotencyKey = %q; want %q", got.GetIdempotencyKey(), "idem-key")
	}
	if resp.GetSessionId() != "sid-123" || resp.GetWorkspaceSlug() != "ABCDEFGH" {
		t.Errorf("response passthrough failed: %+v", resp)
	}
}

func TestActor_RelayExtendSession_ForwardsFields(t *testing.T) {
	actor, mock, cleanup := newRelayTestActor(t)
	defer cleanup()

	var got *proto.RelayExtendSessionRequest
	mock.RelayExtendSessionFunc = func(_ context.Context, req *proto.RelayExtendSessionRequest) (*proto.RelayExtendSessionResponse, error) {
		got = req
		return &proto.RelayExtendSessionResponse{ExpiresAtMs: 42}, nil
	}

	resp, err := actor.RelayExtendSession(context.Background(), "sid-9", 10_000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.GetSessionId() != "sid-9" || got.GetExtendByMs() != 10_000 {
		t.Errorf("forwarded fields incorrect: %+v", got)
	}
	if resp.GetExpiresAtMs() != 42 {
		t.Errorf("ExpiresAtMs = %d; want 42", resp.GetExpiresAtMs())
	}
}

func TestActor_RelayDestroySession_ForwardsFields(t *testing.T) {
	actor, mock, cleanup := newRelayTestActor(t)
	defer cleanup()

	var got *proto.RelayDestroySessionRequest
	mock.RelayDestroySessionFunc = func(_ context.Context, req *proto.RelayDestroySessionRequest) (*proto.RelayDestroySessionResponse, error) {
		got = req
		return &proto.RelayDestroySessionResponse{}, nil
	}

	if err := actor.RelayDestroySession(context.Background(), "sid-9", "user-logout"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.GetSessionId() != "sid-9" || got.GetReason() != "user-logout" {
		t.Errorf("forwarded fields incorrect: %+v", got)
	}
}

func TestActor_RelayDestroySession_PropagatesError(t *testing.T) {
	actor, mock, cleanup := newRelayTestActor(t)
	defer cleanup()

	mock.RelayDestroySessionFunc = func(_ context.Context, _ *proto.RelayDestroySessionRequest) (*proto.RelayDestroySessionResponse, error) {
		return nil, errors.New("boom")
	}
	err := actor.RelayDestroySession(context.Background(), "sid", "reason")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestActor_RelayCreateShortCode_ForwardsFields(t *testing.T) {
	actor, mock, cleanup := newRelayTestActor(t)
	defer cleanup()

	var got *proto.RelayCreateShortCodeRequest
	mock.RelayCreateShortCodeFunc = func(_ context.Context, req *proto.RelayCreateShortCodeRequest) (*proto.RelayCreateShortCodeResponse, error) {
		got = req
		return &proto.RelayCreateShortCodeResponse{Code: "AAA-BBB-CCC", ExpiresAtMs: 123}, nil
	}

	resp, err := actor.RelayCreateShortCode(context.Background(), "sid-1", 2*60*1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.GetSessionId() != "sid-1" || got.GetTtlMs() != 2*60*1000 {
		t.Errorf("forwarded fields incorrect: %+v", got)
	}
	if resp.GetCode() != "AAA-BBB-CCC" || resp.GetExpiresAtMs() != 123 {
		t.Errorf("response passthrough failed: %+v", resp)
	}
}

func TestActor_RelayResolveShortCode_ForwardsFields(t *testing.T) {
	actor, mock, cleanup := newRelayTestActor(t)
	defer cleanup()

	var got *proto.RelayResolveShortCodeRequest
	mock.RelayResolveShortCodeFunc = func(_ context.Context, req *proto.RelayResolveShortCodeRequest) (*proto.RelayResolveShortCodeResponse, error) {
		got = req
		return &proto.RelayResolveShortCodeResponse{
			SessionId:          "sid-x",
			WorkspaceSlug:      "ABCDEFGH",
			SessionExpiresAtMs: 7,
		}, nil
	}

	resp, err := actor.RelayResolveShortCode(context.Background(), "AAA-BBB-CCC")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.GetCode() != "AAA-BBB-CCC" {
		t.Errorf("code not forwarded: %q", got.GetCode())
	}
	if resp.GetSessionId() != "sid-x" || resp.GetWorkspaceSlug() != "ABCDEFGH" {
		t.Errorf("response passthrough failed: %+v", resp)
	}
}

func TestActor_RelayDeleteShortCode_ForwardsFields(t *testing.T) {
	actor, mock, cleanup := newRelayTestActor(t)
	defer cleanup()

	var got *proto.RelayDeleteShortCodeRequest
	mock.RelayDeleteShortCodeFunc = func(_ context.Context, req *proto.RelayDeleteShortCodeRequest) (*proto.RelayDeleteShortCodeResponse, error) {
		got = req
		return &proto.RelayDeleteShortCodeResponse{}, nil
	}
	if err := actor.RelayDeleteShortCode(context.Background(), "AAA-BBB-CCC"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.GetCode() != "AAA-BBB-CCC" {
		t.Errorf("code not forwarded: %q", got.GetCode())
	}
}

func TestActor_RelayPublish_ForwardsFields(t *testing.T) {
	actor, mock, cleanup := newRelayTestActor(t)
	defer cleanup()

	var got *proto.RelayPublishRequest
	mock.RelayPublishFunc = func(_ context.Context, req *proto.RelayPublishRequest) (*proto.RelayPublishResponse, error) {
		got = req
		return &proto.RelayPublishResponse{MsgId: "01HX0123", TsMs: 9}, nil
	}

	payload := []byte(`{"price":42}`)
	resp, err := actor.RelayPublish(context.Background(), "sid-1", "cart.updated", payload, "user-7")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.GetSessionId() != "sid-1" || got.GetType() != "cart.updated" {
		t.Errorf("forwarded fields incorrect: %+v", got)
	}
	if string(got.GetData()) != string(payload) {
		t.Errorf("Data = %q; want %q", got.GetData(), payload)
	}
	if got.GetOriginatorUser() != "user-7" {
		t.Errorf("OriginatorUser = %q; want user-7", got.GetOriginatorUser())
	}
	if resp.GetMsgId() != "01HX0123" || resp.GetTsMs() != 9 {
		t.Errorf("response passthrough failed: %+v", resp)
	}
}

func TestActor_RelayGetPresence_ReturnsDevices(t *testing.T) {
	actor, mock, cleanup := newRelayTestActor(t)
	defer cleanup()

	var got *proto.RelayGetPresenceRequest
	mock.RelayGetPresenceFunc = func(_ context.Context, req *proto.RelayGetPresenceRequest) (*proto.RelayGetPresenceResponse, error) {
		got = req
		return &proto.RelayGetPresenceResponse{
			Devices: []*proto.PresenceDevice{
				{DeviceId: "dev-a", ConnectedSinceMs: 1},
				{DeviceId: "dev-b", ConnectedSinceMs: 2},
			},
		}, nil
	}

	devices, err := actor.RelayGetPresence(context.Background(), "sid-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.GetSessionId() != "sid-1" {
		t.Errorf("session id not forwarded: %q", got.GetSessionId())
	}
	if len(devices) != 2 || devices[0].GetDeviceId() != "dev-a" || devices[1].GetDeviceId() != "dev-b" {
		t.Errorf("unexpected devices: %+v", devices)
	}
}

func TestActor_RelayGetPresence_PropagatesError(t *testing.T) {
	actor, mock, cleanup := newRelayTestActor(t)
	defer cleanup()

	mock.RelayGetPresenceFunc = func(_ context.Context, _ *proto.RelayGetPresenceRequest) (*proto.RelayGetPresenceResponse, error) {
		return nil, errors.New("kaboom")
	}

	if _, err := actor.RelayGetPresence(context.Background(), "sid"); err == nil {
		t.Fatal("expected error, got nil")
	}
}
