# sdk-go

Sensor measurements can be written independently of entity mutations:

```go
response, err := actor.WriteSensors(ctx, &proto.WriteSensorsRequest{
    AppScope: true, // Omit Schema and EntityId for app-scoped sensors.
    TtlSeconds: 300,
    Measurements: []*proto.EntitySensorMeasurement{
        {Sensor: "requests", At: timestamppb.Now(), Value: 1, MeasurementId: "stable-event-id"},
        {Sensor: "requests", At: timestamppb.Now(), Value: 2, TtlSeconds: 600},
    },
})
```

Entity-scoped writes use `Schema` and `EntityId` with `AppScope: false`. They
require a registered entity schema and stable ID, and do not create or update
the entity. `ChartSensor`, `HeatmapSensor`, `SensorActivity`, `PredictSensor`,
and `CreateSensorStream` also accept app scope. `ListAppSensors` discovers the
actor's app sensor names. Actor methods supply authorization and preserve the
caller’s request.

TTL starts at server receipt and removes measurements from raw history,
summaries, and activity profiles. A measurement TTL overrides the request TTL;
zero uses the request default, or normal retention when both are zero. Negative
TTLs are rejected by the server. Existing embedded measurements also accept
options:

```go
entity.AddSensorMeasurement("requests", 1,
    keystone.WithSensorTTL(300),
    keystone.WithSensorMeasurementID("stable-event-id"),
)
```

`WithSensorTimestamp` and `WithSensorBackfill` set event time and backfill status.
Reuse an event's measurement ID, timestamp, and payload when retrying a write.

Run SDK tests with `go test ./...`. Run server requirement tests with
`go run ./test`; the runner uses `127.0.0.1:50051` and the test credentials in
`test/main.go` and `test/conn.go`. The Sensor requirement covers entity and app
scopes, idempotent writes, charts, heatmaps, activity, prediction baselines,
discovery, TTL defaults and overrides, validation, and expiry.

Enable the optional Sensor Live requirement with an origin allowed by the server:

```sh
KS_SENSOR_TEST_ORIGIN=http://localhost:3000 go run ./test
```

The server needs sensor storage, NATS JetStream, `KS_SENSOR_LIVE_KEY`,
`KS_SENSOR_LIVE_ORIGINS`, and an accessible `KS_SENSOR_LIVE_PUBLIC_URL`.
Live requirements check grants, HTTP snapshots, origin rejection, SSE updates,
connection closure on revocation, and denial of new grants for a revoked login,
for both entity and app scopes. Stream creation requires an actor with an
end-user ID and a verified login expiry in the request.
