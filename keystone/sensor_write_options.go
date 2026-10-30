package keystone

import "google.golang.org/grpc"

type sensorBackgroundOption struct {
	grpc.EmptyCallOption
	background bool
}

// WithSensorBackground controls asynchronous ingestion for Actor.WriteSensors
// and Connection.WriteSensors. When enabled, success means the server validated
// and durably queued the measurements; reads may not reflect them immediately.
// The last supplied option wins and overrides the request's Background field.
func WithSensorBackground(background bool) grpc.CallOption {
	return sensorBackgroundOption{background: background}
}
