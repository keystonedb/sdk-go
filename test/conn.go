package main

import (
	"log"
	"time"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

const (
	vendorID = "testing"
	appID    = "tester"
)

var keystoneConnection *keystone.Connection
var ksGrpcConn *grpc.ClientConn
var pClient proto.KeystoneClient

func InitKeystone(kHost, kPort, accessToken string) {
	var err error

	ksGrpcConn, err = grpc.NewClient(kHost+":"+kPort,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithIdleTimeout(time.Minute*5),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  200 * time.Millisecond,
				Multiplier: 1.6,
				Jitter:     0.2,
				MaxDelay:   5 * time.Second,
			},
			MinConnectTimeout: 3 * time.Second,
		}),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithDefaultServiceConfig(`{
			"methodConfig": [{
				"name": [{"service": ""}],
				"waitForReady": true,
				"retryPolicy": {
					"maxAttempts": 3,
					"initialBackoff": "0.5s",
					"maxBackoff": "5s",
					"backoffMultiplier": 2,
					"retryableStatusCodes": ["UNAVAILABLE"]
				}
			}]
		}`),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	pClient = proto.NewKeystoneClient(ksGrpcConn)

	keystoneConnection = keystone.NewConnection(pClient, vendorID, appID, accessToken)
}

func Actor() *keystone.Actor {
	a := keystoneConnection.Actor("tt", "91.91.91.91", "random-userid", "UserAgent")
	return &a
}

func CloseKeystone() error {
	if ksGrpcConn != nil {
		return ksGrpcConn.Close()
	}
	return nil
}
