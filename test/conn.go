package main

import (
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
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

	ksGrpcConn, err = grpc.Dial(kHost+":"+kPort, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithIdleTimeout(time.Minute*5), grpc.WithConnectParams(grpc.ConnectParams{MinConnectTimeout: time.Second * 5}))
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
