package aliases

import (
	"context"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/keystonedb/sdk-go/test/requirements"
	"github.com/packaged/environment/environment"
	"github.com/packaged/logger/v3/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"testing"
	"time"
)

// This is intentionally opt-in: it creates real entities and permanent alias
// ownership claims. Use a disposable test workspace and a migrated server.
func TestAliasesAgainstServer(t *testing.T) {
	endpoint := os.Getenv("KEYSTONE_ALIAS_TEST_ENDPOINT")
	if endpoint == "" {
		t.Skip("set KEYSTONE_ALIAS_TEST_ENDPOINT to run against a disposable Keystone server")
	}
	if err := logger.Setup(environment.UnitTest); err != nil {
		t.Fatal(err)
	}
	token := os.Getenv("KEYSTONE_ALIAS_TEST_TOKEN")
	if token == "" {
		t.Fatal("KEYSTONE_ALIAS_TEST_TOKEN is required")
	}
	ws := os.Getenv("KEYSTONE_ALIAS_TEST_WORKSPACE")
	if ws == "" {
		t.Fatal("KEYSTONE_ALIAS_TEST_WORKSPACE is required")
	}
	transport, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer transport.Close()
	conn := keystone.NewConnection(proto.NewKeystoneClient(transport), "testing", "tester", token)
	// Bound startup instead of waiting indefinitely on schema registration.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	actor := conn.Actor(ws, "127.0.0.1", "alias-test", "go-test")
	if _, err := conn.DirectClient().Status(ctx, actor.Authorization()); err != nil {
		t.Fatalf("server unavailable: %v", err)
	}
	req := &Requirement{}
	if err := req.Register(conn); err != nil {
		t.Fatal(err)
	}
	conn.SyncSchema().Wait()
	count := 0
	req.Verify(&actor, func(result requirements.TestResult) {
		count++
		t.Run(result.Name, func(t *testing.T) {
			if result.Error != nil {
				t.Fatal(result.Error)
			}
		})
	})
	if os.Getenv("KEYSTONE_ALIAS_TEST_DESTROY") == "1" {
		(&DestructionRequirement{}).Verify(&actor, func(result requirements.TestResult) {
			t.Run(result.Name, func(t *testing.T) {
				if result.Error != nil {
					t.Fatal(result.Error)
				}
			})
		})
	}
	if count < 8 {
		t.Fatalf("only %d scenarios executed", count)
	}
}
