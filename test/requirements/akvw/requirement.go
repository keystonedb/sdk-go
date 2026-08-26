package akvw

import (
	"context"
	"fmt"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/keystonedb/sdk-go/test/requirements"
)

const isolationKey = "akvw-workspace-isolation"

type Requirement struct{}

func (d *Requirement) Name() string {
	return "Workspace App Key Value"
}

func (d *Requirement) Register(*keystone.Connection) error {
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor, report requirements.Reporter) {
	report(d.workspaceIsolation(actor))
}

func (d *Requirement) workspaceIsolation(actor *keystone.Actor) requirements.TestResult {
	result := requirements.TestResult{Name: "Put, get, isolate and delete"}
	ctx := context.Background()

	other := actor.Connection().Actor(
		actor.WorkspaceID()+"-akvw-other",
		actor.RemoteIP(),
		actor.UserID(),
		actor.UserAgent(),
	)

	// Start clean and ensure the secondary workspace is cleaned up even when an
	// assertion below fails.
	_, _ = actor.AKVWDel(ctx, isolationKey)
	_, _ = other.AKVWDel(ctx, isolationKey)
	defer func() {
		_, _ = actor.AKVWDel(ctx, isolationKey)
		_, _ = other.AKVWDel(ctx, isolationKey)
	}()

	if err := put(ctx, actor, "primary"); err != nil {
		return result.WithError(err)
	}
	if err := put(ctx, &other, "secondary"); err != nil {
		return result.WithError(err)
	}

	if err := expect(ctx, actor, "primary"); err != nil {
		return result.WithError(err)
	}
	if err := expect(ctx, &other, "secondary"); err != nil {
		return result.WithError(err)
	}

	delResp, err := actor.AKVWDel(ctx, isolationKey)
	if err != nil {
		return result.WithError(fmt.Errorf("delete primary workspace value: %w", err))
	}
	if err := responseError("delete primary workspace value", delResp); err != nil {
		return result.WithError(err)
	}

	values, err := actor.AKVWGet(ctx, isolationKey)
	if err != nil {
		return result.WithError(fmt.Errorf("get deleted primary workspace value: %w", err))
	}
	if _, exists := values[isolationKey]; exists {
		return result.WithError(fmt.Errorf("primary workspace value still exists after delete"))
	}
	if err := expect(ctx, &other, "secondary"); err != nil {
		return result.WithError(fmt.Errorf("deleting primary workspace affected secondary workspace: %w", err))
	}

	return result
}

func put(ctx context.Context, actor *keystone.Actor, value string) error {
	resp, err := actor.AKVWPut(ctx, keystone.AKV(isolationKey, value))
	if err != nil {
		return fmt.Errorf("put %q: %w", value, err)
	}
	return responseError("put "+value, resp)
}

func expect(ctx context.Context, actor *keystone.Actor, want string) error {
	values, err := actor.AKVWGet(ctx, isolationKey)
	if err != nil {
		return fmt.Errorf("get %q: %w", want, err)
	}
	value, exists := values[isolationKey]
	if !exists {
		return fmt.Errorf("value %q not found", want)
	}
	if got := value.GetText(); got != want {
		return fmt.Errorf("value = %q, want %q", got, want)
	}
	return nil
}

func responseError(operation string, resp *proto.GenericResponse) error {
	if resp == nil {
		return fmt.Errorf("%s: empty response", operation)
	}
	if !resp.GetSuccess() {
		return fmt.Errorf("%s: %d - %s", operation, resp.GetErrorCode(), resp.GetErrorMessage())
	}
	return nil
}
