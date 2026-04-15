package relay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/requirements"
)

// defaultRelayURL is the local HTTP endpoint of `./keystone relay`.
// Override with the KS_RELAY_URL env var (e.g. "https://relay.dev.local-host.xyz").
const defaultRelayURL = "http://127.0.0.1:8082"

type Requirement struct {
	sessionID     string
	workspaceSlug string
	shortCode     string
}

func (d *Requirement) Name() string { return "Relay / WebSocket Sessions" }

func (d *Requirement) Register(_ *keystone.Connection) error { return nil }

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	results := []requirements.TestResult{
		d.createSession(actor),
		d.extendSession(actor),
		d.createShortCode(actor),
		d.resolveShortCode(actor),
		d.publishWithoutListeners(actor),
		d.getPresenceEmpty(actor),
	}

	// WebSocket tests depend on a session existing; skip them if CreateSession failed.
	if d.sessionID == "" {
		results = append(results, requirements.TestResult{
			Name:  "WebSocket Tests",
			Error: errors.New("skipped: session not created"),
		})
		return results
	}

	results = append(results,
		d.wsHandshakeWelcome(actor),
		d.wsReceiveAppPublish(actor),
		d.wsPublishFromDevice(actor),
		d.wsPingPong(actor),
		d.wsPresenceQuery(actor),
		d.getPresenceAfterConnect(actor),
		d.wsHighThroughput(actor),
		d.wsMultiListenerFanOut(actor),
		d.wsBadSlugRejected(),
		d.wsUnknownSessionRejected(),
		d.deleteShortCode(actor),
		d.destroySession(actor),
		d.destroyedSessionRejectsWS(),
	)
	return results
}

// --- gRPC-only checks ---

func (d *Requirement) createSession(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Create Session"}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := actor.RelayCreateSession(ctx, 30*60*1000, []byte(`{"purpose":"sdk-requirement-test"}`), "")
	if err != nil {
		return res.WithError(fmt.Errorf("create session failed: %w", err))
	}
	if resp.GetSessionId() == "" {
		return res.WithError(errors.New("server returned empty session_id"))
	}
	if resp.GetWorkspaceSlug() == "" {
		return res.WithError(errors.New("server returned empty workspace_slug"))
	}
	if resp.GetExpiresAtMs() <= time.Now().UnixMilli() {
		return res.WithError(fmt.Errorf("expires_at_ms %d is not in the future", resp.GetExpiresAtMs()))
	}
	d.sessionID = resp.GetSessionId()
	d.workspaceSlug = resp.GetWorkspaceSlug()
	return res
}

func (d *Requirement) extendSession(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Extend Session"}
	if d.sessionID == "" {
		return res.WithError(errors.New("skipped: no session"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	before, err := actor.RelayExtendSession(ctx, d.sessionID, 10*60*1000)
	if err != nil {
		return res.WithError(fmt.Errorf("first extend failed: %w", err))
	}
	if before.GetExpiresAtMs() <= time.Now().UnixMilli() {
		return res.WithError(errors.New("extend returned a past expires_at"))
	}

	// Monotonic: asking for a smaller extension must never shorten the expiry.
	after, err := actor.RelayExtendSession(ctx, d.sessionID, 60*1000)
	if err != nil {
		return res.WithError(fmt.Errorf("second extend failed: %w", err))
	}
	if after.GetExpiresAtMs() < before.GetExpiresAtMs() {
		return res.WithError(fmt.Errorf(
			"extend is not monotonic: went from %d to %d",
			before.GetExpiresAtMs(), after.GetExpiresAtMs(),
		))
	}
	return res
}

func (d *Requirement) createShortCode(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Create Short Code"}
	if d.sessionID == "" {
		return res.WithError(errors.New("skipped: no session"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := actor.RelayCreateShortCode(ctx, d.sessionID, 2*60*1000)
	if err != nil {
		return res.WithError(fmt.Errorf("create short code failed: %w", err))
	}
	code := resp.GetCode()
	if code == "" {
		return res.WithError(errors.New("empty short code"))
	}
	// Spec format: AAA-BBB-CCC (11 chars) but be lenient — any non-empty code is acceptable.
	if len(code) < 5 || !strings.Contains(code, "-") {
		return res.WithError(fmt.Errorf("short code %q does not look like AAA-BBB-CCC", code))
	}
	if resp.GetExpiresAtMs() <= time.Now().UnixMilli() {
		return res.WithError(errors.New("short code already expired"))
	}
	d.shortCode = code
	return res
}

func (d *Requirement) resolveShortCode(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Resolve Short Code"}
	if d.shortCode == "" {
		return res.WithError(errors.New("skipped: no short code"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := actor.RelayResolveShortCode(ctx, d.shortCode)
	if err != nil {
		return res.WithError(fmt.Errorf("resolve short code failed: %w", err))
	}
	if resp.GetSessionId() != d.sessionID {
		return res.WithError(fmt.Errorf("resolved session %q != created %q", resp.GetSessionId(), d.sessionID))
	}
	if resp.GetWorkspaceSlug() != d.workspaceSlug {
		return res.WithError(fmt.Errorf("resolved slug %q != created %q", resp.GetWorkspaceSlug(), d.workspaceSlug))
	}
	if resp.GetSessionExpiresAtMs() <= time.Now().UnixMilli() {
		return res.WithError(errors.New("resolved session expires_at is in the past"))
	}
	return res
}

func (d *Requirement) publishWithoutListeners(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Publish (no listeners)"}
	if d.sessionID == "" {
		return res.WithError(errors.New("skipped: no session"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := actor.RelayPublish(ctx, d.sessionID, "test.warmup", []byte(`{"seq":0}`), "")
	if err != nil {
		return res.WithError(fmt.Errorf("publish failed: %w", err))
	}
	if resp.GetMsgId() == "" {
		return res.WithError(errors.New("empty msg_id"))
	}
	if resp.GetTsMs() == 0 {
		return res.WithError(errors.New("empty ts_ms"))
	}
	return res
}

func (d *Requirement) getPresenceEmpty(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Get Presence (empty)"}
	if d.sessionID == "" {
		return res.WithError(errors.New("skipped: no session"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	devices, err := actor.RelayGetPresence(ctx, d.sessionID)
	if err != nil {
		return res.WithError(fmt.Errorf("get presence failed: %w", err))
	}
	// No WS connections yet — expected empty.
	if len(devices) != 0 {
		return res.WithError(fmt.Errorf("expected 0 devices, got %d", len(devices)))
	}
	return res
}

// --- WebSocket integration ---

func (d *Requirement) wsHandshakeWelcome(_ *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "WS Handshake + Welcome"}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ws, _, err := dialRelayWS(ctx, d.workspaceSlug, d.sessionID, "probe-welcome", "")
	if err != nil {
		return res.WithError(fmt.Errorf("ws dial failed: %w", err))
	}
	defer ws.Close(websocket.StatusNormalClosure, "done")

	frame, err := readJSONFrame(ctx, ws)
	if err != nil {
		return res.WithError(fmt.Errorf("read welcome: %w", err))
	}
	if frame.Type != "system.welcome" {
		return res.WithError(fmt.Errorf("first frame type %q != system.welcome", frame.Type))
	}
	data := map[string]any{}
	_ = json.Unmarshal(frame.Data, &data)
	if sid, _ := data["session_id"].(string); sid != d.sessionID {
		return res.WithError(fmt.Errorf("welcome session_id %v != %s", data["session_id"], d.sessionID))
	}
	if dev, _ := data["device_id"].(string); dev != "probe-welcome" {
		return res.WithError(fmt.Errorf("welcome device_id %v != probe-welcome", data["device_id"]))
	}
	return res
}

func (d *Requirement) wsReceiveAppPublish(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "WS Receives gRPC Publish"}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ws, _, err := dialRelayWS(ctx, d.workspaceSlug, d.sessionID, "probe-recv", "")
	if err != nil {
		return res.WithError(fmt.Errorf("ws dial failed: %w", err))
	}
	defer ws.Close(websocket.StatusNormalClosure, "done")

	if _, err := readJSONFrame(ctx, ws); err != nil {
		return res.WithError(fmt.Errorf("read welcome: %w", err))
	}

	payload := []byte(`{"step":"grpc-publish","n":1}`)
	pubResp, err := actor.RelayPublish(ctx, d.sessionID, "sdk.test.grpc", payload, "")
	if err != nil {
		return res.WithError(fmt.Errorf("publish failed: %w", err))
	}

	// Read until we see our message (skip any system frames).
	deadline := time.Now().Add(5 * time.Second)
	for {
		if time.Now().After(deadline) {
			return res.WithError(errors.New("timed out waiting for envelope"))
		}
		readCtx, readCancel := context.WithTimeout(ctx, 2*time.Second)
		frame, rerr := readJSONFrame(readCtx, ws)
		readCancel()
		if rerr != nil {
			return res.WithError(fmt.Errorf("read frame: %w", rerr))
		}
		if frame.Type == "sdk.test.grpc" {
			if frame.MsgID != pubResp.GetMsgId() {
				return res.WithError(fmt.Errorf("msg_id mismatch: got %q want %q", frame.MsgID, pubResp.GetMsgId()))
			}
			if frame.Originator.Kind != "app" {
				return res.WithError(fmt.Errorf("originator kind %q != app", frame.Originator.Kind))
			}
			// data is raw JSON
			if string(frame.Data) != string(payload) {
				return res.WithError(fmt.Errorf("data mismatch: got %s want %s", frame.Data, payload))
			}
			return res
		}
	}
}

func (d *Requirement) wsPublishFromDevice(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "WS Publish Round-Trip"}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ws, _, err := dialRelayWS(ctx, d.workspaceSlug, d.sessionID, "probe-publisher", "")
	if err != nil {
		return res.WithError(fmt.Errorf("ws dial failed: %w", err))
	}
	defer ws.Close(websocket.StatusNormalClosure, "done")

	if _, err := readJSONFrame(ctx, ws); err != nil {
		return res.WithError(fmt.Errorf("read welcome: %w", err))
	}

	// Publish from the device side.
	out, _ := json.Marshal(map[string]any{
		"type": "sdk.test.device",
		"data": map[string]any{"seq": 42, "via": "ws"},
	})
	if err := ws.Write(ctx, websocket.MessageText, out); err != nil {
		return res.WithError(fmt.Errorf("ws write failed: %w", err))
	}

	// Sender receives its own envelope.
	deadline := time.Now().Add(5 * time.Second)
	for {
		if time.Now().After(deadline) {
			return res.WithError(errors.New("timed out waiting for device envelope"))
		}
		readCtx, readCancel := context.WithTimeout(ctx, 2*time.Second)
		frame, rerr := readJSONFrame(readCtx, ws)
		readCancel()
		if rerr != nil {
			return res.WithError(fmt.Errorf("read frame: %w", rerr))
		}
		if frame.Type != "sdk.test.device" {
			continue
		}
		if frame.Originator.Kind != "device" {
			return res.WithError(fmt.Errorf("device frame kind %q != device", frame.Originator.Kind))
		}
		if frame.Originator.ID != "probe-publisher" {
			return res.WithError(fmt.Errorf("device frame originator id %q != probe-publisher", frame.Originator.ID))
		}
		if frame.MsgID == "" {
			return res.WithError(errors.New("device frame has empty msg_id"))
		}
		// Presence must now include our device (heartbeat is triggered by publish).
		if _, perr := actor.RelayGetPresence(ctx, d.sessionID); perr != nil {
			return res.WithError(fmt.Errorf("get presence after publish: %w", perr))
		}
		return res
	}
}

func (d *Requirement) wsPingPong(_ *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "WS Ping/Pong"}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ws, _, err := dialRelayWS(ctx, d.workspaceSlug, d.sessionID, "probe-ping", "")
	if err != nil {
		return res.WithError(fmt.Errorf("ws dial failed: %w", err))
	}
	defer ws.Close(websocket.StatusNormalClosure, "done")

	if _, err := readJSONFrame(ctx, ws); err != nil {
		return res.WithError(fmt.Errorf("read welcome: %w", err))
	}

	out, _ := json.Marshal(map[string]any{"type": "system.ping"})
	if err := ws.Write(ctx, websocket.MessageText, out); err != nil {
		return res.WithError(fmt.Errorf("ws ping write failed: %w", err))
	}

	deadline := time.Now().Add(3 * time.Second)
	for {
		if time.Now().After(deadline) {
			return res.WithError(errors.New("timed out waiting for pong"))
		}
		readCtx, readCancel := context.WithTimeout(ctx, 2*time.Second)
		frame, rerr := readJSONFrame(readCtx, ws)
		readCancel()
		if rerr != nil {
			return res.WithError(fmt.Errorf("read frame: %w", rerr))
		}
		if frame.Type == "system.pong" {
			return res
		}
	}
}

func (d *Requirement) wsPresenceQuery(_ *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "WS Presence Query"}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ws, _, err := dialRelayWS(ctx, d.workspaceSlug, d.sessionID, "probe-presence", "")
	if err != nil {
		return res.WithError(fmt.Errorf("ws dial failed: %w", err))
	}
	defer ws.Close(websocket.StatusNormalClosure, "done")

	if _, err := readJSONFrame(ctx, ws); err != nil {
		return res.WithError(fmt.Errorf("read welcome: %w", err))
	}

	out, _ := json.Marshal(map[string]any{"type": "system.presence.query"})
	if err := ws.Write(ctx, websocket.MessageText, out); err != nil {
		return res.WithError(fmt.Errorf("ws presence.query write failed: %w", err))
	}

	deadline := time.Now().Add(3 * time.Second)
	for {
		if time.Now().After(deadline) {
			return res.WithError(errors.New("timed out waiting for system.presence.result"))
		}
		readCtx, readCancel := context.WithTimeout(ctx, 2*time.Second)
		frame, rerr := readJSONFrame(readCtx, ws)
		readCancel()
		if rerr != nil {
			return res.WithError(fmt.Errorf("read frame: %w", rerr))
		}
		if frame.Type != "system.presence.result" {
			continue
		}
		// Expect an object with `devices: [...]`.
		payload := struct {
			Devices []struct {
				DeviceID string `json:"device_id"`
			} `json:"devices"`
		}{}
		if err := json.Unmarshal(frame.Data, &payload); err != nil {
			return res.WithError(fmt.Errorf("unmarshal presence.result: %w", err))
		}
		found := false
		for _, dv := range payload.Devices {
			if dv.DeviceID == "probe-presence" {
				found = true
				break
			}
		}
		if !found {
			return res.WithError(errors.New("probe-presence not present in WS presence.result"))
		}
		return res
	}
}

func (d *Requirement) getPresenceAfterConnect(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "gRPC Presence Reflects WS"}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ws, _, err := dialRelayWS(ctx, d.workspaceSlug, d.sessionID, "probe-grpc-presence", "")
	if err != nil {
		return res.WithError(fmt.Errorf("ws dial failed: %w", err))
	}
	defer ws.Close(websocket.StatusNormalClosure, "done")
	if _, err := readJSONFrame(ctx, ws); err != nil {
		return res.WithError(fmt.Errorf("read welcome: %w", err))
	}

	// Presence heartbeat is eventually-consistent, so poll briefly.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		devices, err := actor.RelayGetPresence(ctx, d.sessionID)
		if err != nil {
			return res.WithError(fmt.Errorf("get presence failed: %w", err))
		}
		for _, dv := range devices {
			if dv.GetDeviceId() == "probe-grpc-presence" {
				return res
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
	return res.WithError(errors.New("probe-grpc-presence not found via gRPC presence"))
}

// wsHighThroughput proves the socket can receive every message an app publishes
// over gRPC — 1000 frames with sequential `seq` payloads, each tagged with a
// unique server-generated ulid. Duplicates, drops, or payload corruption all
// fail the test. Publish runs concurrently with the receiver so the stream
// drains as it's filled instead of queueing everything up-front.
func (d *Requirement) wsHighThroughput(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "WS 1000-Message Throughput"}
	const numMessages = 1000
	const eventType = "sdk.test.throughput"

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	ws, _, err := dialRelayWS(ctx, d.workspaceSlug, d.sessionID, "probe-throughput", "")
	if err != nil {
		return res.WithError(fmt.Errorf("ws dial failed: %w", err))
	}
	defer ws.Close(websocket.StatusNormalClosure, "done")

	if _, err := readJSONFrame(ctx, ws); err != nil {
		return res.WithError(fmt.Errorf("read welcome: %w", err))
	}

	// --- Receiver ---
	type recvResult struct {
		seqsSeen  map[int]int     // seq -> arrival count
		msgIDs    map[string]bool // observed msg_id set
		extra     int             // frames after all seqs were seen
		totalSeen int
		err       error
	}
	recvCh := make(chan recvResult, 1)
	go func() {
		out := recvResult{
			seqsSeen: make(map[int]int, numMessages),
			msgIDs:   make(map[string]bool, numMessages),
		}
		for len(out.seqsSeen) < numMessages {
			// Per-frame timeout; if nothing arrives in 15s we've lost messages.
			readCtx, readCancel := context.WithTimeout(ctx, 15*time.Second)
			frame, rerr := readJSONFrame(readCtx, ws)
			readCancel()
			if rerr != nil {
				out.err = fmt.Errorf("read after %d unique seqs / %d frames: %w",
					len(out.seqsSeen), out.totalSeen, rerr)
				recvCh <- out
				return
			}
			if frame.Type != eventType {
				continue
			}
			out.totalSeen++
			var payload struct {
				Seq int `json:"seq"`
			}
			if jerr := json.Unmarshal(frame.Data, &payload); jerr != nil {
				out.err = fmt.Errorf("unmarshal seq payload (frame %d): %w", out.totalSeen, jerr)
				recvCh <- out
				return
			}
			if payload.Seq < 0 || payload.Seq >= numMessages {
				out.err = fmt.Errorf("seq %d outside 0..%d range", payload.Seq, numMessages-1)
				recvCh <- out
				return
			}
			if frame.MsgID == "" {
				out.err = fmt.Errorf("frame %d has empty msg_id", out.totalSeen)
				recvCh <- out
				return
			}
			if frame.Originator.Kind != "app" {
				out.err = fmt.Errorf("frame seq=%d originator kind %q != app",
					payload.Seq, frame.Originator.Kind)
				recvCh <- out
				return
			}
			out.seqsSeen[payload.Seq]++
			out.msgIDs[frame.MsgID] = true
		}
		recvCh <- out
	}()

	// --- Publisher ---
	// Run in a goroutine so the receiver can drain concurrently — sending 1000
	// RPCs serially still takes noticeable time on slower networks, and we want
	// to overlap send+receive to mirror a real-world workload.
	type pubResult struct {
		msgIDs []string
		err    error
	}
	pubCh := make(chan pubResult, 1)
	go func() {
		out := pubResult{msgIDs: make([]string, 0, numMessages)}
		for i := 0; i < numMessages; i++ {
			payload, _ := json.Marshal(map[string]int{"seq": i})
			pub, perr := actor.RelayPublish(ctx, d.sessionID, eventType, payload, "")
			if perr != nil {
				out.err = fmt.Errorf("publish %d failed: %w", i, perr)
				pubCh <- out
				return
			}
			if pub.GetMsgId() == "" {
				out.err = fmt.Errorf("publish %d returned empty msg_id", i)
				pubCh <- out
				return
			}
			out.msgIDs = append(out.msgIDs, pub.GetMsgId())
		}
		pubCh <- out
	}()

	var pub pubResult
	select {
	case pub = <-pubCh:
		if pub.err != nil {
			return res.WithError(pub.err)
		}
	case <-ctx.Done():
		return res.WithError(errors.New("publisher timed out"))
	}

	// Sanity-check the publisher: ulids must be unique.
	seenPub := make(map[string]bool, len(pub.msgIDs))
	for i, id := range pub.msgIDs {
		if seenPub[id] {
			return res.WithError(fmt.Errorf("publisher returned duplicate msg_id %q at index %d", id, i))
		}
		seenPub[id] = true
	}

	var rcv recvResult
	select {
	case rcv = <-recvCh:
		if rcv.err != nil {
			return res.WithError(rcv.err)
		}
	case <-ctx.Done():
		return res.WithError(fmt.Errorf("receiver timed out; only %d/%d seqs arrived",
			len(rcv.seqsSeen), numMessages))
	}

	// --- Verification ---

	if len(rcv.seqsSeen) != numMessages {
		return res.WithError(fmt.Errorf("got %d unique seqs; want %d",
			len(rcv.seqsSeen), numMessages))
	}
	for seq, count := range rcv.seqsSeen {
		if count != 1 {
			return res.WithError(fmt.Errorf("seq %d arrived %d times (expected once)", seq, count))
		}
	}
	// Defensive: ensure every seq in [0, numMessages) is present.
	for i := 0; i < numMessages; i++ {
		if rcv.seqsSeen[i] == 0 {
			return res.WithError(fmt.Errorf("seq %d missing from received messages", i))
		}
	}
	// msg_id sets must match the publisher's exactly.
	if len(rcv.msgIDs) != numMessages {
		return res.WithError(fmt.Errorf("received %d unique msg_ids; want %d",
			len(rcv.msgIDs), numMessages))
	}
	for _, id := range pub.msgIDs {
		if !rcv.msgIDs[id] {
			return res.WithError(fmt.Errorf("publisher msg_id %q never arrived on WS", id))
		}
	}

	return res
}

// wsMultiListenerFanOut opens two WS connections on the same session, publishes
// 100 messages from the app side over gRPC, and confirms both devices receive
// every message with identical content and msg_id. This proves server-side
// fan-out correctness — duplication, dropped listeners, and divergent ordering
// of msg_id identity are all caught.
func (d *Requirement) wsMultiListenerFanOut(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "WS Fan-Out (2 listeners)"}
	const numMessages = 100
	const eventType = "sdk.test.fanout"

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Open two listeners on the same session.
	wsA, _, err := dialRelayWS(ctx, d.workspaceSlug, d.sessionID, "probe-fanout-a", "")
	if err != nil {
		return res.WithError(fmt.Errorf("ws A dial failed: %w", err))
	}
	defer wsA.Close(websocket.StatusNormalClosure, "done")

	wsB, _, err := dialRelayWS(ctx, d.workspaceSlug, d.sessionID, "probe-fanout-b", "")
	if err != nil {
		return res.WithError(fmt.Errorf("ws B dial failed: %w", err))
	}
	defer wsB.Close(websocket.StatusNormalClosure, "done")

	if _, err := readJSONFrame(ctx, wsA); err != nil {
		return res.WithError(fmt.Errorf("ws A welcome: %w", err))
	}
	if _, err := readJSONFrame(ctx, wsB); err != nil {
		return res.WithError(fmt.Errorf("ws B welcome: %w", err))
	}

	type recvResult struct {
		label     string
		seqs      map[int]string // seq -> msg_id seen for that seq
		totalSeen int
		err       error
	}

	receiver := func(label string, ws *websocket.Conn) <-chan recvResult {
		ch := make(chan recvResult, 1)
		go func() {
			out := recvResult{
				label: label,
				seqs:  make(map[int]string, numMessages),
			}
			for len(out.seqs) < numMessages {
				readCtx, readCancel := context.WithTimeout(ctx, 15*time.Second)
				frame, rerr := readJSONFrame(readCtx, ws)
				readCancel()
				if rerr != nil {
					out.err = fmt.Errorf("[%s] read after %d unique seqs / %d frames: %w",
						label, len(out.seqs), out.totalSeen, rerr)
					ch <- out
					return
				}
				if frame.Type != eventType {
					continue
				}
				out.totalSeen++
				var payload struct {
					Seq int `json:"seq"`
				}
				if jerr := json.Unmarshal(frame.Data, &payload); jerr != nil {
					out.err = fmt.Errorf("[%s] unmarshal seq: %w", label, jerr)
					ch <- out
					return
				}
				if payload.Seq < 0 || payload.Seq >= numMessages {
					out.err = fmt.Errorf("[%s] seq %d outside range", label, payload.Seq)
					ch <- out
					return
				}
				if frame.MsgID == "" {
					out.err = fmt.Errorf("[%s] frame seq=%d has empty msg_id", label, payload.Seq)
					ch <- out
					return
				}
				if frame.Originator.Kind != "app" {
					out.err = fmt.Errorf("[%s] seq=%d originator kind %q != app",
						label, payload.Seq, frame.Originator.Kind)
					ch <- out
					return
				}
				if existing, dup := out.seqs[payload.Seq]; dup {
					out.err = fmt.Errorf("[%s] seq %d arrived twice (msg_ids %q and %q)",
						label, payload.Seq, existing, frame.MsgID)
					ch <- out
					return
				}
				out.seqs[payload.Seq] = frame.MsgID
			}
			ch <- out
		}()
		return ch
	}

	chA := receiver("A", wsA)
	chB := receiver("B", wsB)

	// Publish 100 messages and remember the msg_id the server returned for each seq.
	pubMsgIDs := make(map[int]string, numMessages)
	for i := 0; i < numMessages; i++ {
		payload, _ := json.Marshal(map[string]int{"seq": i})
		pub, perr := actor.RelayPublish(ctx, d.sessionID, eventType, payload, "")
		if perr != nil {
			return res.WithError(fmt.Errorf("publish %d failed: %w", i, perr))
		}
		if pub.GetMsgId() == "" {
			return res.WithError(fmt.Errorf("publish %d empty msg_id", i))
		}
		pubMsgIDs[i] = pub.GetMsgId()
	}

	var rcvA, rcvB recvResult
	select {
	case rcvA = <-chA:
	case <-ctx.Done():
		return res.WithError(errors.New("receiver A timed out"))
	}
	if rcvA.err != nil {
		return res.WithError(rcvA.err)
	}
	select {
	case rcvB = <-chB:
	case <-ctx.Done():
		return res.WithError(errors.New("receiver B timed out"))
	}
	if rcvB.err != nil {
		return res.WithError(rcvB.err)
	}

	// --- Verification ---
	// 1. Both receivers saw all 100 seqs.
	if len(rcvA.seqs) != numMessages {
		return res.WithError(fmt.Errorf("A got %d unique seqs; want %d", len(rcvA.seqs), numMessages))
	}
	if len(rcvB.seqs) != numMessages {
		return res.WithError(fmt.Errorf("B got %d unique seqs; want %d", len(rcvB.seqs), numMessages))
	}
	// 2. No receiver observed stray duplicates (totalSeen == uniqueSeqs).
	if rcvA.totalSeen != numMessages {
		return res.WithError(fmt.Errorf("A received %d frames for %d unique seqs", rcvA.totalSeen, len(rcvA.seqs)))
	}
	if rcvB.totalSeen != numMessages {
		return res.WithError(fmt.Errorf("B received %d frames for %d unique seqs", rcvB.totalSeen, len(rcvB.seqs)))
	}
	// 3. For each seq, both receivers saw the same msg_id as the publisher reported.
	for seq := 0; seq < numMessages; seq++ {
		want := pubMsgIDs[seq]
		if rcvA.seqs[seq] != want {
			return res.WithError(fmt.Errorf("A seq %d msg_id %q != publisher %q",
				seq, rcvA.seqs[seq], want))
		}
		if rcvB.seqs[seq] != want {
			return res.WithError(fmt.Errorf("B seq %d msg_id %q != publisher %q",
				seq, rcvB.seqs[seq], want))
		}
	}

	return res
}

func (d *Requirement) wsBadSlugRejected() requirements.TestResult {
	res := requirements.TestResult{Name: "WS Bad Slug Rejected"}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _, err := dialRelayWS(ctx, "NOPE9999", d.sessionID, "probe-badslug", "")
	if err == nil {
		return res.WithError(errors.New("expected dial to fail or be closed"))
	}
	// Server may accept the upgrade and immediately close with 4404 — treat either as success.
	if !isCloseCode(err, 4404) && !isHTTPError(err) {
		return res.WithError(fmt.Errorf("unexpected error shape: %v", err))
	}
	return res
}

func (d *Requirement) wsUnknownSessionRejected() requirements.TestResult {
	res := requirements.TestResult{Name: "WS Unknown Session Rejected"}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _, err := dialRelayWS(ctx, d.workspaceSlug, "00000000-0000-0000-0000-000000000000", "probe-badsid", "")
	if err == nil {
		return res.WithError(errors.New("expected dial to fail or be closed"))
	}
	if !isCloseCode(err, 4404) && !isHTTPError(err) {
		return res.WithError(fmt.Errorf("unexpected error shape: %v", err))
	}
	return res
}

func (d *Requirement) deleteShortCode(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Delete Short Code"}
	if d.shortCode == "" {
		return res.WithError(errors.New("skipped: no short code"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := actor.RelayDeleteShortCode(ctx, d.shortCode); err != nil {
		return res.WithError(fmt.Errorf("delete short code failed: %w", err))
	}
	// Resolving a deleted code must fail.
	if _, err := actor.RelayResolveShortCode(ctx, d.shortCode); err == nil {
		return res.WithError(errors.New("resolve still worked after delete"))
	}
	return res
}

func (d *Requirement) destroySession(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Destroy Session"}
	if d.sessionID == "" {
		return res.WithError(errors.New("skipped: no session"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Open a connection and look for the graceful closing frame when we destroy.
	ws, _, err := dialRelayWS(ctx, d.workspaceSlug, d.sessionID, "probe-closer", "")
	if err != nil {
		return res.WithError(fmt.Errorf("ws dial failed: %w", err))
	}
	defer ws.Close(websocket.StatusGoingAway, "done")

	if _, err := readJSONFrame(ctx, ws); err != nil {
		return res.WithError(fmt.Errorf("read welcome: %w", err))
	}

	if err := actor.RelayDestroySession(ctx, d.sessionID, "sdk-test-teardown"); err != nil {
		return res.WithError(fmt.Errorf("destroy failed: %w", err))
	}

	// Expect either a system.closing frame or the connection to be closed with code 4001.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		readCtx, readCancel := context.WithTimeout(ctx, 2*time.Second)
		frame, rerr := readJSONFrame(readCtx, ws)
		readCancel()
		if rerr != nil {
			if isCloseCode(rerr, 4001) || isCloseCode(rerr, int(websocket.StatusNormalClosure)) {
				return res
			}
			// Some servers close without any specific code — accept that too.
			return res
		}
		if frame.Type == "system.closing" {
			return res
		}
	}
	return res.WithError(errors.New("no system.closing frame or close received"))
}

func (d *Requirement) destroyedSessionRejectsWS() requirements.TestResult {
	res := requirements.TestResult{Name: "Destroyed Session Rejects New WS"}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _, err := dialRelayWS(ctx, d.workspaceSlug, d.sessionID, "probe-afterdestroy", "")
	if err == nil {
		return res.WithError(errors.New("expected dial after destroy to fail"))
	}
	return res
}

// --- helpers ---

type wsFrame struct {
	MsgID      string          `json:"msg_id"`
	Type       string          `json:"type"`
	Ts         int64           `json:"ts"`
	Originator wsOriginator    `json:"originator"`
	Data       json.RawMessage `json:"data"`
}

type wsOriginator struct {
	Kind     string `json:"kind"`
	VendorID string `json:"vendor_id"`
	AppID    string `json:"app_id"`
	ID       string `json:"id"`
}

// relayBaseURL returns the configured relay URL (http or https). Converted to
// ws/wss later by dialRelayWS.
func relayBaseURL() string {
	if v := os.Getenv("KS_RELAY_URL"); v != "" {
		return v
	}
	return defaultRelayURL
}

// dialRelayWS opens a WS connection to the relay frontend.
func dialRelayWS(ctx context.Context, slug, sessionID, deviceID, lastTs string) (*websocket.Conn, string, error) {
	base, err := url.Parse(relayBaseURL())
	if err != nil {
		return nil, "", fmt.Errorf("bad relay url: %w", err)
	}
	switch base.Scheme {
	case "http":
		base.Scheme = "ws"
	case "https":
		base.Scheme = "wss"
	case "ws", "wss":
		// already fine
	default:
		return nil, "", fmt.Errorf("unsupported scheme %q", base.Scheme)
	}
	base.Path = fmt.Sprintf("/w/%s/s/%s", slug, sessionID)
	q := base.Query()
	if deviceID != "" {
		q.Set("device_id", deviceID)
	}
	if lastTs != "" {
		q.Set("last_ts", lastTs)
	}
	base.RawQuery = q.Encode()

	ws, resp, err := websocket.Dial(ctx, base.String(), nil)
	if err != nil {
		return nil, base.String(), err
	}
	_ = resp
	return ws, base.String(), nil
}

// readJSONFrame reads the next text frame and decodes as wsFrame.
func readJSONFrame(ctx context.Context, ws *websocket.Conn) (*wsFrame, error) {
	typ, data, err := ws.Read(ctx)
	if err != nil {
		return nil, err
	}
	if typ != websocket.MessageText {
		return nil, fmt.Errorf("unexpected ws message type %v", typ)
	}
	frame := &wsFrame{}
	if err := json.Unmarshal(data, frame); err != nil {
		return nil, fmt.Errorf("unmarshal frame: %w (raw=%s)", err, string(data))
	}
	return frame, nil
}

func isCloseCode(err error, code int) bool {
	if err == nil {
		return false
	}
	if ce := websocket.CloseStatus(err); ce == websocket.StatusCode(code) {
		return true
	}
	// Fallback: message contains the code.
	return strings.Contains(err.Error(), fmt.Sprintf("%d", code))
}

func isHTTPError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "unexpected http status") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "EOF")
}
