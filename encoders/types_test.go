package encoders

import (
	"errors"
	"github.com/keystonedb/sdk-go/sdk-go/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"reflect"
	"testing"
	"time"
)

func Test_TimeEncoder(t *testing.T) {
	tNow := time.Now()
	tests := []struct {
		name  string
		input interface{}
		want  func(reflect.Value) (*proto.Value, error)
	}{
		{"time", tNow, Time},
		{"time pointer", &tNow, Time},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			val, err := Time(reflect.ValueOf(test.input))
			if err != nil {
				t.Errorf("timeEncoder returned error: %v", err)
			}
			if val.GetTime().AsTime().Unix() != tNow.Unix() {
				t.Errorf("timeEncoder returned %v, want %v", val.GetTime().AsTime(), tNow)
			}
		})
	}

	res, err := Time(reflect.ValueOf("string"))
	if err == nil || !errors.Is(err, InvalidTimeError) {
		t.Errorf("timeEncoder returned no error, want InvalidTimeError")
	}
	if res != nil {
		t.Errorf("timeEncoder returned %v, want nil", res)
	}
}

func Test_TimestampEncoder(t *testing.T) {
	tNow := time.Now()
	tsNow := timestamppb.New(tNow)
	tests := []struct {
		name  string
		input interface{}
		want  func(reflect.Value) (*proto.Value, error)
	}{
		{"timestamp", *tsNow, Time},
		{"timestamp pointer", tsNow, Time},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			val, err := Timestamp(reflect.ValueOf(test.input))
			if err != nil {
				t.Errorf("timestampEncoder returned error: %v", err)
			}
			if val.GetTime().AsTime().Unix() != tNow.Unix() {
				t.Errorf("timestampEncoder returned %v, want %v", val.GetTime().AsTime(), tNow)
			}
		})
	}

	res, err := Timestamp(reflect.ValueOf("string"))
	if err == nil || !errors.Is(err, InvalidTimestampError) {
		t.Errorf("timestampEncoder returned no error, want InvalidTimestampError")
	}
	if res != nil {
		t.Errorf("timestampEncoder returned %v, want nil", res)
	}
}
