package aliases

import (
	"context"
	"fmt"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/requirements"
	"github.com/kubex/k4id"
	"google.golang.org/grpc/codes"
	"time"
)

// DestructionRequirement needs the server's normal backup/object-storage setup.
// It destroys only a new customer created by this scenario.
type DestructionRequirement struct{ Requirement }

func (*DestructionRequirement) Name() string { return "Entity Alias Destruction" }
func (*DestructionRequirement) Verify(actor *keystone.Actor, report requirements.Reporter) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	f := &fixture{actor: actor, key: k4id.New().String()}
	report(requirements.NewResult("Destroy deactivates aliases and retains ownership", destroy(ctx, f)))
	if err := f.cleanup(); err != nil {
		report(requirements.NewResult("cleanup", err))
	}
}
func destroy(ctx context.Context, f *fixture) error {
	c, err := f.customer(ctx)
	if err != nil {
		return err
	}
	s := f.schema()
	a := f.alias()
	if err := f.attach(ctx, s, c.GetKeystoneID(), a); err != nil {
		return err
	}
	destroyed, err := f.actor.PermanentlyDestroyEntity(ctx, s.Key, c.GetKeystoneID(), keystone.EidHash(c.GetKeystoneID()), "Alias requirement disposable customer")
	if err != nil {
		return err
	}
	if !destroyed {
		return fmt.Errorf("customer was not destroyed")
	}
	_, err = f.actor.ResolveAlias(ctx, s.Key, a)
	if err := expectCode(err, codes.NotFound); err != nil {
		return err
	}
	rows, err := f.actor.ListAliases(ctx, s.Key, c.GetKeystoneID())
	if err != nil {
		return err
	}
	if len(rows) != 0 {
		return fmt.Errorf("destroyed customer still lists aliases")
	}
	other, err := f.customer(ctx)
	if err != nil {
		return err
	}
	return expectCode(f.actor.AttachAlias(ctx, s.Key, other.GetKeystoneID(), a), codes.AlreadyExists)
}
