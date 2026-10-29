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

Pass `keystone.WithSensorBackground(true)` as a call option to either
`actor.WriteSensors(ctx, request, ...)` or `conn.WriteSensors(ctx, request, ...)`
to enable background ingestion. Success then means the measurements were
validated and durably queued; poll reads when you need to observe completion.
`WithSensorBackground(false)` overrides a request that enables background mode.
The option preserves the caller's request and can be mixed with gRPC call options.

TTL starts at server receipt and removes measurements from raw history,
summaries, and activity profiles. A measurement TTL overrides the request TTL;
zero uses the request default, or normal retention when both are zero. Negative
TTLs are rejected by the server. Build standalone measurements with options:

```go
measurement := keystone.NewSensorMeasurement("requests", 1,
    keystone.WithSensorTTL(300),
    keystone.WithSensorMeasurementID("stable-event-id"),
)
```

`WithSensorTimestamp` and `WithSensorBackfill` set event time and backfill status.
Pass these measurements in `WriteSensorsRequest.Measurements`. Entity mutations
no longer collect or write sensors: `SensorProvider`, `EmbeddedSensors`, and the
entity `AddSensorMeasurement` methods have been removed. Migrate callers to
`NewSensorMeasurement` and `actor.WriteSensors`.

Reuse an event's measurement ID, timestamp, and payload when retrying a write.

Run SDK tests with `go test ./...`. Run server requirement tests with
`go run ./test`; the runner uses `127.0.0.1:50051` and the test credentials in
`test/main.go` and `test/conn.go`. The Sensor requirement covers entity and app
scopes, idempotent writes, charts, heatmaps, activity, prediction baselines,
discovery, TTL defaults and overrides, validation, and expiry.

The requirement runner reports elapsed time, effective wall time (including
RPC latency), aggregate active work, idle waiting time, and active time per
requirement. Effective wall time merges overlapping active intervals; aggregate
work sums them and may exceed elapsed time when requirements run concurrently.
Scheduled delays are excluded from active work. Registration and schema sync run
sequentially before verification and are included in the timing report.

Requirements are exclusive by default. Independent requirements can opt into
concurrency by implementing:

```go
func (d *Requirement) Parallel() bool { return true }
```

Sensor and TaskQ opt in, using unique fixture identities and isolated task queues.
The runner allows four active parallel requirements by default; use
`go run ./test -parallel 2` to change the limit, or `-parallel 1` for sequential
execution. Requirements without the opt-in never overlap active parallel work.
Report callbacks are serialized, and each requirement gets its own actor.

Requirements implementing `requirements.ScheduledRequirement` start early.
`VerifyScheduled` uses the scoped `scheduler.After(delay, check)` to resume
assertions or poll later. Due continuations take priority over ordinary checks.
Callbacks within one requirement remain sequential, even when it opts into
parallel execution, and waiting callbacks release their worker slot. Use only
the scheduler supplied to `VerifyScheduled` for continuations. Callbacks should
return promptly and schedule another attempt rather than sleeping; blocking RPCs
still count toward active time. Ordinary exclusive requirements retain their
registration order.

Sensor TTL and background-ingestion checks use this scheduling. Task queue
checks no longer add fixed producer/consumer sleeps; their 20-second timeout
remains a failure bound. The 1ms timestamp-spacing sleeps in the child/grandchild
fixtures remain part of setup time; these are not expiry tests.

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
