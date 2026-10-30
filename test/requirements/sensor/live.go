package sensor

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/keystonedb/sdk-go/test/requirements"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// LiveRequirement needs a server with sensor live enabled, NATS JetStream,
// an accessible public HTTP URL, and Origin in its allowed origins.
type LiveRequirement struct {
	Origin string
}

func (d *LiveRequirement) Name() string { return "Sensor Live" }

func (d *LiveRequirement) Register(conn *keystone.Connection) error {
	return (&Requirement{}).Register(conn)
}

func (d *LiveRequirement) Verify(actor *keystone.Actor, report requirements.Reporter) {
	for _, appScope := range []bool{false, true} {
		fixture := &Requirement{fixture: sensorFixture{appScope: appScope}}
		stored := fixture.storeQueryMeasurements(actor)
		stored.Name = fmt.Sprintf("Stream fixture (app scope=%t)", appScope)
		report(stored)
		if stored.Error != nil {
			continue
		}
		report(requirements.NewResult(fmt.Sprintf("Stream snapshot, events and revocation (app scope=%t)", appScope), d.verifyStream(actor, fixture)))
	}
}

type sensorSnapshot struct {
	Metric string `json:"metric"`
	Points []struct {
		Value float64 `json:"value"`
	} `json:"points"`
}

func checkSnapshot(frame sensorSnapshot) error {
	var sum float64
	for _, point := range frame.Points {
		sum += point.Value
	}
	if frame.Metric != "sum" || !closeEnough(sum, 12) {
		return fmt.Errorf("want snapshot metric=sum and total=12; got metric=%q total=%g", frame.Metric, sum)
	}
	return nil
}

func (d *LiveRequirement) verifyStream(actor *keystone.Actor, fixture *Requirement) error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	request := &proto.CreateSensorStreamRequest{
		Schema: fixture.schema(), EntityId: fixture.fixture.entityID, AppScope: fixture.fixture.appScope,
		Sensor: fixture.sensorName(querySensor), SessionId: uuid.NewString(),
		SessionExpiresAt: timestamppb.New(time.Now().Add(time.Hour)), Origin: d.Origin,
		WindowSeconds: 3600, IntervalSeconds: 60, Metric: proto.SensorPredictionMetric_SensorPredictionSum,
	}
	grant, err := actor.CreateSensorStream(ctx, request)
	if err != nil {
		return err
	}
	// Clean up a grant even when an HTTP assertion fails.
	defer func() {
		cleanup, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		_, _ = actor.RevokeSensorSession(cleanup, &proto.RevokeSensorSessionRequest{SessionId: request.SessionId})
	}()
	if grant.GetExpiresAt() == nil || grant.GetServerTime() == nil || grant.GetExpiresAt().CheckValid() != nil || grant.GetServerTime().CheckValid() != nil {
		return fmt.Errorf("stream grant returned invalid expiry or server time")
	}
	lifetime := grant.GetExpiresAt().AsTime().Sub(grant.GetServerTime().AsTime())
	if lifetime <= 0 || lifetime > time.Minute || grant.GetExpiresAt().AsTime().After(request.SessionExpiresAt.AsTime()) {
		return fmt.Errorf("invalid stream grant lifetime %v", lifetime)
	}
	for _, endpoint := range []string{grant.GetSnapshotUrl(), grant.GetStreamUrl()} {
		parsed, err := url.Parse(endpoint)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return fmt.Errorf("stream grant returned an invalid HTTP URL")
		}
	}
	client := &http.Client{Timeout: 15 * time.Second}
	get := func(endpoint, origin string) (*http.Response, error) {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}
		request.Header.Set("Origin", origin)
		return client.Do(request)
	}
	response, err := get(grant.GetSnapshotUrl(), d.Origin)
	if err != nil {
		return err
	}
	var frame sensorSnapshot
	err = json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&frame)
	_ = response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("snapshot returned HTTP %d", response.StatusCode)
	}
	if err != nil {
		return err
	}
	if err = checkSnapshot(frame); err != nil {
		return err
	}
	response, err = get(grant.GetSnapshotUrl(), "https://invalid-origin.example")
	if err != nil {
		return err
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		return fmt.Errorf("wrong origin returned HTTP %d; want 403", response.StatusCode)
	}
	stream, err := get(grant.GetStreamUrl(), d.Origin)
	if err != nil {
		return err
	}
	defer stream.Body.Close()
	if stream.StatusCode != http.StatusOK || !strings.HasPrefix(stream.Header.Get("Content-Type"), "text/event-stream") {
		return fmt.Errorf("events returned HTTP %d with content type %q", stream.StatusCode, stream.Header.Get("Content-Type"))
	}
	scanner := bufio.NewScanner(stream.Body)
	scanner.Buffer(make([]byte, 4096), 1<<20)
	var event, data string
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" && event != "" {
			break
		}
		if strings.HasPrefix(line, "event: ") {
			event = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "data: ") {
			data = strings.TrimPrefix(line, "data: ")
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if event != "snapshot" {
		return fmt.Errorf("first stream event is %q; want snapshot", event)
	}
	if err = json.Unmarshal([]byte(data), &frame); err != nil {
		return err
	}
	if err = checkSnapshot(frame); err != nil {
		return err
	}
	// A new direct write must reach the already-open SSE connection.
	written, err := actor.WriteSensors(ctx, &proto.WriteSensorsRequest{
		Schema: fixture.schema(), EntityId: fixture.fixture.entityID, AppScope: fixture.fixture.appScope,
		Measurements: []*proto.EntitySensorMeasurement{{Sensor: request.Sensor,
			At: timestamppb.Now(), MeasurementId: uuid.NewString(), Value: 4}},
	})
	if err != nil {
		return err
	}
	if !written.GetSuccess() {
		return fmt.Errorf("live measurement write did not succeed")
	}
	foundUpdate := false
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		var update sensorSnapshot
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &update); err != nil {
			return err
		}
		for _, point := range update.Points {
			if closeEnough(point.Value, 4) {
				foundUpdate = true
				break
			}
		}
		if foundUpdate {
			break
		}
	}
	if !foundUpdate {
		return fmt.Errorf("SSE connection did not deliver the new measurement: %v", scanner.Err())
	}
	if _, err = actor.RevokeSensorSession(ctx, &proto.RevokeSensorSessionRequest{SessionId: request.SessionId}); err != nil {
		return err
	}
	// The established SSE connection must close, not keep sending until grant expiry.
	for scanner.Scan() {
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("revoked stream did not close cleanly: %w", err)
	}
	response, err = get(grant.GetSnapshotUrl(), d.Origin)
	if err != nil {
		return err
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		return fmt.Errorf("revoked snapshot returned HTTP %d; want 403", response.StatusCode)
	}
	_, err = actor.CreateSensorStream(ctx, request)
	if status.Code(err) != codes.PermissionDenied {
		return fmt.Errorf("revoked login obtained a new grant: %v", err)
	}
	return nil
}
