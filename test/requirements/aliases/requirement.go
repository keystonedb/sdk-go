package aliases

import (
	"context"
	"fmt"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"github.com/kubex/k4id"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"sync"
	"time"
)

type Customer struct {
	keystone.BaseEntity
	Name string
}

func (*Customer) GetKeystoneDefinition() keystone.TypeDefinition {
	return keystone.TypeDefinition{Type: "alias-customer"}
}

type PrivateCustomer struct {
	keystone.BaseEntity
	Name string
}

func (*PrivateCustomer) GetKeystoneDefinition() keystone.TypeDefinition {
	return keystone.TypeDefinition{Type: "private-alias-customer"}
}

type LinkedRecord struct {
	keystone.BaseEntity
	CustomerID string
}
type Requirement struct{}

func (*Requirement) Name() string { return "Entity Aliases" }
func (*Requirement) Register(c *keystone.Connection) error {
	c.RegisterTypes(Customer{}, PrivateCustomer{}, models.Hashi{})
	return nil
}
func (*Requirement) Verify(a *keystone.Actor, report requirements.Reporter) {
	for _, test := range []struct {
		name string
		run  func(context.Context, *fixture) error
	}{
		{"Attach, retry, resolve, list and load customer page", lifecycle},
		{"Remove, retry, reactivate and permanent ownership", removal},
		{"Routing, vendor, app, workspace and schema isolation", isolation},
		{"Concurrent claims have one winner", concurrent},
		{"Hashed-ID customer resolves and loads by alias", hashed},
		{"Linked app resolves and updates its own record", linked},
		{"App namespace enforcement without schema opt-in", permissions},
		{"Validation, missing aliases and missing entities", validation},
	} {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		f := &fixture{actor: a, key: k4id.New().String()}
		err := test.run(ctx, f)
		cancel()
		report(requirements.NewResult(test.name, err))
		if err := f.cleanup(); err != nil {
			report(requirements.NewResult(test.name+" cleanup", err))
		}
	}
}

type attached struct {
	schema *proto.Key
	id     keystone.ID
	alias  *keystone.EntityAlias
}
type fixture struct {
	actor *keystone.Actor
	key   string
	rows  []attached
}

func (f *fixture) schema() *proto.Key {
	return &proto.Key{Source: f.actor.VendorApp(), Key: "alias-customer"}
}
func (f *fixture) alias() *keystone.EntityAlias {
	return keystone.NewEntityAlias("ven2", "ap2", f.key)
}
func (f *fixture) customer(ctx context.Context) (*Customer, error) {
	c := &Customer{Name: "Customer " + f.key}
	err := f.actor.Mutate(ctx, c)
	return c, err
}
func (f *fixture) attach(ctx context.Context, s *proto.Key, id keystone.ID, a *keystone.EntityAlias) error {
	f.rows = append(f.rows, attached{s, id, a})
	return f.actor.AttachAlias(ctx, s.Key, id, a)
}
func (f *fixture) cleanup() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	for _, r := range f.rows {
		if err := f.actor.RemoveAlias(ctx, r.schema.Key, r.id, r.alias); err != nil && status.Code(err) != codes.AlreadyExists {
			return err
		}
	}
	return nil
}
func expectCode(err error, code codes.Code) error {
	if status.Code(err) != code {
		return fmt.Errorf("expected %s, got %v", code, err)
	}
	return nil
}
func expectID(ctx context.Context, a *keystone.Actor, s *proto.Key, alias *keystone.EntityAlias, want keystone.ID) error {
	id, err := a.ResolveAlias(ctx, s.Key, alias, keystone.WithAliasOwner(s.Source.VendorId, s.Source.AppId))
	if err != nil {
		return err
	}
	if id != want {
		return fmt.Errorf("resolved %q, want %q", id, want)
	}
	return nil
}
func lifecycle(ctx context.Context, f *fixture) error {
	c, err := f.customer(ctx)
	if err != nil {
		return err
	}
	s := f.schema()
	a := f.alias()
	rows, err := f.actor.ListAliases(ctx, s.Key, c.GetKeystoneID())
	if err != nil {
		return err
	}
	if len(rows) != 0 {
		return fmt.Errorf("new entity has aliases")
	}
	for n := 0; n < 2; n++ {
		if err := f.attach(ctx, s, c.GetKeystoneID(), a); err != nil {
			return err
		}
	}
	if err := expectID(ctx, f.actor, s, a, c.GetKeystoneID()); err != nil {
		return err
	}
	rows, err = f.actor.ListAliases(ctx, s.Key, c.GetKeystoneID())
	if err != nil {
		return err
	}
	if len(rows) != 1 || rows[0].ExternalId != a.ExternalId {
		return fmt.Errorf("idempotent attach listed %v", rows)
	}
	for _, by := range []keystone.RetrieveBy{keystone.ByAlias(c, a), keystone.ByEntityID(c, c.GetKeystoneID())} {
		dst := &Customer{}
		aliases := keystone.WithAliases()
		if err := f.actor.Get(ctx, by, dst, keystone.WithProperties("name"), aliases); err != nil {
			return err
		}
		if dst.Name != c.Name || dst.GetKeystoneID() != c.GetKeystoneID() || len(aliases.Aliases()) != 1 {
			return fmt.Errorf("customer page lost properties, ID or aliases")
		}
	}
	dst := &Customer{}
	return f.actor.GetByAlias(ctx, a, dst)
}
func removal(ctx context.Context, f *fixture) error {
	c, err := f.customer(ctx)
	if err != nil {
		return err
	}
	other, err := f.customer(ctx)
	if err != nil {
		return err
	}
	s := f.schema()
	a := f.alias()
	if err := f.attach(ctx, s, c.GetKeystoneID(), a); err != nil {
		return err
	}
	if err := expectCode(f.actor.AttachAlias(ctx, s.Key, other.GetKeystoneID(), a), codes.AlreadyExists); err != nil {
		return err
	}
	if err := expectCode(f.actor.RemoveAlias(ctx, s.Key, other.GetKeystoneID(), a), codes.AlreadyExists); err != nil {
		return err
	}
	for n := 0; n < 2; n++ {
		if err := f.actor.RemoveAlias(ctx, s.Key, c.GetKeystoneID(), a); err != nil {
			return err
		}
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
		return fmt.Errorf("removed alias still listed")
	}
	if err := expectCode(f.actor.AttachAlias(ctx, s.Key, other.GetKeystoneID(), a), codes.AlreadyExists); err != nil {
		return err
	}
	if err := f.attach(ctx, s, c.GetKeystoneID(), a); err != nil {
		return err
	}
	absent := f.alias()
	absent.ExternalId += "absent"
	if err := f.actor.RemoveAlias(ctx, s.Key, c.GetKeystoneID(), absent); err != nil {
		return err
	}
	if err := f.attach(ctx, s, other.GetKeystoneID(), absent); err != nil {
		return err
	}
	return expectID(ctx, f.actor, s, a, c.GetKeystoneID())
}
func isolation(ctx context.Context, f *fixture) error {
	c, err := f.customer(ctx)
	if err != nil {
		return err
	}
	other, err := f.customer(ctx)
	if err != nil {
		return err
	}
	s := f.schema()
	for _, a := range []*keystone.EntityAlias{
		f.alias(), keystone.NewEntityAlias("ven2", "ap2", f.key).WithRouting("eu"),
		{Source: &proto.VendorApp{VendorId: "ven2", AppId: "other"}, ExternalId: f.key},
		{Source: &proto.VendorApp{VendorId: "other", AppId: "ap2"}, ExternalId: f.key},
	} {
		id := other.GetKeystoneID()
		if a.RoutingKey == "eu" {
			id = c.GetKeystoneID()
		}
		if err := f.attach(ctx, s, id, a); err != nil {
			return err
		}
		if err := expectID(ctx, f.actor, s, a, id); err != nil {
			return err
		}
	}
	unknownRoute := f.alias()
	unknownRoute.RoutingKey = "unknown"
	_, err = f.actor.ResolveAlias(ctx, s.Key, unknownRoute)
	if err := expectCode(err, codes.NotFound); err != nil {
		return err
	}
	otherWorkspace := f.actor.Connection().Actor("zz", "", "", "")
	_, err = otherWorkspace.ResolveAlias(ctx, s.Key, f.alias())
	if err := expectCode(err, codes.NotFound); err != nil {
		return err
	}
	private := &PrivateCustomer{Name: "private"}
	if err := f.actor.Mutate(ctx, private); err != nil {
		return err
	}
	ps := &proto.Key{Source: s.Source, Key: "private-alias-customer"}
	if err := f.attach(ctx, ps, private.GetKeystoneID(), f.alias()); err != nil {
		return err
	}
	return expectID(ctx, f.actor, ps, f.alias(), private.GetKeystoneID())
}
func concurrent(ctx context.Context, f *fixture) error {
	c, err := f.customer(ctx)
	if err != nil {
		return err
	}
	other, err := f.customer(ctx)
	if err != nil {
		return err
	}
	s := f.schema()
	a := f.alias()
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, id := range []keystone.ID{c.GetKeystoneID(), other.GetKeystoneID()} {
		f.rows = append(f.rows, attached{s, id, a})
		wg.Add(1)
		go func(id keystone.ID) { defer wg.Done(); results <- f.actor.AttachAlias(ctx, s.Key, id, a) }(id)
	}
	wg.Wait()
	close(results)
	successes, conflicts := 0, 0
	for err := range results {
		if err == nil {
			successes++
		} else if status.Code(err) == codes.AlreadyExists {
			conflicts++
		} else {
			return err
		}
	}
	if successes != 1 || conflicts != 1 {
		return fmt.Errorf("claims: %d successes, %d conflicts", successes, conflicts)
	}
	winner, err := f.actor.ResolveAlias(ctx, s.Key, a)
	if err != nil {
		return err
	}
	if winner != c.GetKeystoneID() && winner != other.GetKeystoneID() {
		return fmt.Errorf("unexpected winner %s", winner)
	}
	for _, id := range []keystone.ID{c.GetKeystoneID(), other.GetKeystoneID()} {
		rows, err := f.actor.ListAliases(ctx, s.Key, id)
		if err != nil {
			return err
		}
		want := 0
		if id == winner {
			want = 1
		}
		if len(rows) != want {
			return fmt.Errorf("claim projection mismatch")
		}
	}
	return nil
}
func (f *fixture) linkedActor(vendor, app string) *keystone.Actor {
	conn := keystone.NewConnection(f.actor.Connection().DirectClient(), vendor, app, f.actor.Authorization().Token)
	a := conn.Actor(f.actor.WorkspaceID(), "", "", "")
	return &a
}
func linked(ctx context.Context, f *fixture) error {
	c, err := f.customer(ctx)
	if err != nil {
		return err
	}
	a := f.alias()
	linked := f.linkedActor(a.Source.VendorId, a.Source.AppId)
	f.rows = append(f.rows, attached{f.schema(), c.GetKeystoneID(), a})
	if err := linked.AttachAlias(ctx, f.schema().Key, c.GetKeystoneID(), a, keystone.WithAliasOwner(f.schema().Source.VendorId, f.schema().Source.AppId)); err != nil {
		return err
	}
	if err := linked.RemoveAlias(ctx, f.schema().Key, c.GetKeystoneID(), a, keystone.WithAliasOwner(f.schema().Source.VendorId, f.schema().Source.AppId)); err != nil {
		return err
	}
	if _, err := linked.ResolveAlias(ctx, f.schema().Key, a, keystone.WithAliasOwner(f.schema().Source.VendorId, f.schema().Source.AppId)); status.Code(err) != codes.NotFound {
		return fmt.Errorf("linked removal: %v", err)
	}
	if err := linked.AttachAlias(ctx, f.schema().Key, c.GetKeystoneID(), a, keystone.WithAliasOwner(f.schema().Source.VendorId, f.schema().Source.AppId)); err != nil {
		return err
	}
	id, err := linked.ResolveAlias(ctx, f.schema().Key, a, keystone.WithAliasOwner(f.schema().Source.VendorId, f.schema().Source.AppId))
	if err != nil {
		return err
	}
	record := &LinkedRecord{CustomerID: string(id)}
	if err := linked.Mutate(ctx, record); err != nil {
		return err
	}
	loaded := &LinkedRecord{}
	if err := linked.GetByID(ctx, record.GetKeystoneID(), loaded, keystone.WithProperties("customer_id")); err != nil {
		return err
	}
	if loaded.CustomerID != string(c.GetKeystoneID()) {
		return fmt.Errorf("linked record did not retain customer ID")
	}
	return nil
}
func permissions(ctx context.Context, f *fixture) error {
	c, err := f.customer(ctx)
	if err != nil {
		return err
	}
	s := f.schema()
	a := f.alias()
	if err := f.attach(ctx, s, c.GetKeystoneID(), a); err != nil {
		return err
	}
	linked := f.linkedActor(a.Source.VendorId, a.Source.AppId)
	denied := []func() error{
		func() error {
			return f.linkedActor("ven2", "other").AttachAlias(ctx, s.Key, c.GetKeystoneID(), a, keystone.WithAliasOwner(s.Source.VendorId, s.Source.AppId))
		},
		func() error {
			return f.linkedActor("ven2", "other").RemoveAlias(ctx, s.Key, c.GetKeystoneID(), a, keystone.WithAliasOwner(s.Source.VendorId, s.Source.AppId))
		},
		func() error {
			_, err := linked.ListAliases(ctx, s.Key, c.GetKeystoneID(), keystone.WithAliasOwner(s.Source.VendorId, s.Source.AppId))
			return err
		},
		func() error {
			_, err := linked.Connection().Retrieve(ctx, &proto.EntityRequest{Authorization: linked.Authorization(), Schema: s, Alias: a})
			return err
		},
		func() error {
			_, err := linked.Connection().Retrieve(ctx, &proto.EntityRequest{Authorization: linked.Authorization(), Schema: s, EntityId: string(c.GetKeystoneID()), View: &proto.EntityView{Aliases: true}})
			return err
		},
		func() error {
			_, err := f.linkedActor("ven2", "other").ResolveAlias(ctx, s.Key, a, keystone.WithAliasOwner(s.Source.VendorId, s.Source.AppId))
			return err
		},
	}
	for _, call := range denied {
		if err := expectCode(call(), codes.PermissionDenied); err != nil {
			return err
		}
	}
	p := &PrivateCustomer{}
	if err := f.actor.Mutate(ctx, p); err != nil {
		return err
	}
	ps := &proto.Key{Source: s.Source, Key: "private-alias-customer"}
	if err := f.attach(ctx, ps, p.GetKeystoneID(), a); err != nil {
		return err
	}
	if err := expectID(ctx, linked, ps, a, p.GetKeystoneID()); err != nil {
		return err
	}
	return expectID(ctx, f.actor, ps, a, p.GetKeystoneID())
}
func validation(ctx context.Context, f *fixture) error {
	c, err := f.customer(ctx)
	if err != nil {
		return err
	}
	s := f.schema()
	a := f.alias()
	_, err = f.actor.ResolveAlias(ctx, s.Key, a)
	if err := expectCode(err, codes.NotFound); err != nil {
		return err
	}
	if err := expectCode(f.actor.AttachAlias(ctx, s.Key, keystone.ID(k4id.New().String()), a), codes.NotFound); err != nil {
		return err
	}
	for _, bad := range []*keystone.EntityAlias{nil, {}, {Source: a.Source}, {ExternalId: "id"}, {Source: a.Source, ExternalId: strings.Repeat("x", 1025)}, {Source: a.Source, ExternalId: "id", RoutingKey: strings.Repeat("r", 513)}} {
		if err := expectCode(f.actor.AttachAlias(ctx, s.Key, c.GetKeystoneID(), bad), codes.InvalidArgument); err != nil {
			return err
		}
	}
	if err := expectCode(f.actor.AttachAlias(ctx, s.Key, "bad/id", a), codes.InvalidArgument); err != nil {
		return err
	}
	_, err = f.actor.ResolveAlias(ctx, "", a)
	if err == nil {
		return fmt.Errorf("empty entity type accepted")
	}
	_, err = f.actor.Connection().Retrieve(ctx, &proto.EntityRequest{Authorization: f.actor.Authorization(), Schema: s, EntityId: string(c.GetKeystoneID()), Alias: a})
	return expectCode(err, codes.InvalidArgument)
}

func hashed(ctx context.Context, f *fixture) error {
	c := &models.Hashi{FirstName: "Alias", LastName: f.key}
	if err := c.SetHashID(f.key); err != nil {
		return err
	}
	if err := f.actor.Mutate(ctx, c); err != nil {
		return err
	}
	s := &proto.Key{Source: f.actor.VendorApp(), Key: keystone.Type(c)}
	// Hashed models retain the hash selector after Mutate; retrieve the canonical ID
	// from the raw response rather than guessing the server's hashing algorithm.
	resp, err := f.actor.Connection().Retrieve(ctx, &proto.EntityRequest{Authorization: f.actor.Authorization(), Schema: s, EntityId: string(keystone.HashID(f.key)), View: &proto.EntityView{Summary: true}})
	if err != nil {
		return err
	}
	eid := keystone.ID(resp.GetEntity().GetEntityId())
	if eid == "" {
		return fmt.Errorf("hashed entity returned no canonical ID")
	}
	a := f.alias()
	if err := f.attach(ctx, s, eid, a); err != nil {
		return err
	}
	if err := expectID(ctx, f.actor, s, a, eid); err != nil {
		return err
	}
	loaded := &models.Hashi{}
	aliases := keystone.WithAliases()
	if err := f.actor.GetByAlias(ctx, a, loaded, keystone.WithProperties("first_name", "last_name"), aliases); err != nil {
		return err
	}
	if loaded.FirstName != c.FirstName || loaded.LastName != c.LastName || len(aliases.Aliases()) != 1 {
		return fmt.Errorf("hashed alias retrieval mismatch")
	}
	return nil
}
