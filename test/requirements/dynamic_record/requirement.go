package dynamic_record

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/requirements"
	"github.com/kubex/k4id"
)

// settings deliberately has no schema definition or registration.
type settings struct {
	keystone.DynamicRecordEntity
	Name    string `keystone:"display_name"`
	Enabled bool
	Retries int64
	Ratio   float64
	Updated time.Time
	Budget  keystone.Amount
	Labels  map[string]string
	Names   []string
	Counts  []int64
	Data    []byte
	Secret  keystone.SecureString
	Nested  *endpoint
}

type endpoint struct {
	URL     string
	Enabled bool
}

type Requirement struct{}

func (*Requirement) Name() string { return "Dynamic Records" }

func (*Requirement) Register(*keystone.Connection) error { return nil }

func (*Requirement) Verify(actor *keystone.Actor, report requirements.Reporter) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	key := "requirements/dynamic-record/" + k4id.New().String()
	client := actor.DynamicRecord(key)
	deleted := false
	// Stop dependent steps after a failure, but always clean up the unique key.
	defer func() {
		if deleted {
			return
		}
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		_, err := client.Delete(cleanupCtx, "Requirement cleanup")
		report(requirements.NewResult("Cleanup", err))
	}()
	run := func(name string, check func() error) bool {
		err := check()
		report(requirements.NewResult(name, err))
		return err == nil
	}

	expected := settings{
		DynamicRecordEntity: keystone.NewDynamicRecordEntity(key),
		Name:                "Initial settings", Enabled: true, Retries: 3, Ratio: 1.25,
		Updated: time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC),
		Budget:  *keystone.NewAmount("GBP", 1250),
		Labels:  map[string]string{"region": "uk", "tier": "test"},
		Names:   []string{"first", "second", "second"}, Counts: []int64{0, 2, 4},
		Data: []byte{0, 1, 127, 255}, Secret: keystone.NewSecureString("credential", "****"),
		Nested: &endpoint{URL: "https://initial.example", Enabled: true},
	}
	assertStored := func() error {
		var got settings
		if err := actor.GetDynamicRecord(ctx, key, &got); err != nil {
			return err
		}
		if !got.DynamicRecordExists() || got.GetDynamicRecordKey() != key || got.DynamicRecordRevision() == "" {
			return fmt.Errorf("missing record identity, existence or revision after read")
		}
		return equalSettings(got, expected)
	}

	if !run("Missing record clears stale fields", func() error {
		got := expected
		if err := actor.GetDynamicRecord(ctx, key, &got); err != nil {
			return err
		}
		if got.DynamicRecordExists() || got.GetDynamicRecordKey() != key {
			return fmt.Errorf("missing record has incorrect metadata")
		}
		return equalSettings(got, settings{})
	}) {
		return
	}
	if !run("Write full typed record", func() error {
		if err := actor.Mutate(ctx, &expected, keystone.WithMutationComment("Create settings")); err != nil {
			return err
		}
		if !expected.DynamicRecordExists() || expected.DynamicRecordRevision() == "" {
			return fmt.Errorf("write did not populate existence and revision")
		}
		return nil
	}) {
		return
	}
	if !run("Read full typed record", assertStored) {
		return
	}
	if !run("Read modify write full record", func() error {
		var loaded settings
		if err := client.Load(ctx, &loaded); err != nil {
			return err
		}
		loaded.Name = "Modified settings"
		loaded.Enabled = false
		loaded.Retries = 0
		loaded.Ratio = 2.5
		loaded.Updated = expected.Updated.Add(time.Hour)
		loaded.Budget = *keystone.NewAmount("USD", 500)
		loaded.Labels = map[string]string{"region": "us"}
		loaded.Names = []string{"replacement"}
		loaded.Counts = []int64{9, 0}
		loaded.Data = []byte("modified")
		loaded.Secret = keystone.NewSecureString("new credential", "new ****")
		loaded.Nested = &endpoint{URL: "https://modified.example", Enabled: false}
		if err := actor.Mutate(ctx, &loaded); err != nil {
			return err
		}
		expected = loaded
		return assertStored()
	}) {
		return
	}
	if !run("Unchanged mutation preserves revision", func() error {
		revision := expected.DynamicRecordRevision()
		if err := actor.Mutate(ctx, &expected); err != nil {
			return err
		}
		raw, err := client.Get(ctx)
		if err != nil {
			return err
		}
		if raw.GetRevision() != revision || expected.DynamicRecordRevision() != revision {
			return fmt.Errorf("unchanged mutation changed revision")
		}
		return nil
	}) {
		return
	}
	if !run("Update one typed property preserves other properties", func() error {
		pending := expected
		pending.Name = "Single property update"
		pending.Retries = 999 // Deliberately changed but excluded from this write.
		if err := actor.Mutate(ctx, &pending, keystone.MutateProperties("display_name")); err != nil {
			return err
		}
		expected.Name = pending.Name
		return assertStored()
	}) {
		return
	}
	if !run("Set one raw property preserves other properties", func() error {
		if _, err := client.Set(ctx, "retries", int64(7)); err != nil {
			return err
		}
		expected.Retries = 7
		return assertStored()
	}) {
		return
	}
	if !run("Read selected raw properties only", func() error {
		raw, err := client.Get(ctx, "display_name", "enabled", "absent")
		if err != nil {
			return err
		}
		if !raw.GetExists() || raw.GetKey() != key || len(raw.GetProperties()) != 2 {
			return fmt.Errorf("unexpected selection: %v", raw)
		}
		name, hasName := raw.Properties["display_name"]
		enabled, hasEnabled := raw.Properties["enabled"]
		if !hasName || !hasEnabled || name.GetText() != expected.Name || enabled.GetBool() != expected.Enabled {
			return fmt.Errorf("selected values mismatch: %v", raw)
		}
		return nil
	}) {
		return
	}
	if !run("Read selected typed properties preserves unselected fields", func() error {
		got := settings{Name: "stale", Enabled: true, Retries: 123, Nested: &endpoint{URL: "local", Enabled: true}}
		if err := actor.GetDynamicRecord(ctx, key, &got, "display_name", "enabled", "nested.url"); err != nil {
			return err
		}
		want := settings{Name: expected.Name, Enabled: expected.Enabled, Retries: 123, Nested: &endpoint{URL: expected.Nested.URL, Enabled: true}}
		if err := equalSettings(got, want); err != nil {
			return err
		}
		// Saving a partially loaded object must not submit its local defaults.
		got.Name = "Saved partial record"
		if err := actor.Mutate(ctx, &got); err != nil {
			return err
		}
		expected.Name = got.Name
		return assertStored()
	}) {
		return
	}
	if !run("Update nested property preserves sibling", func() error {
		var loaded settings
		if err := client.Load(ctx, &loaded); err != nil {
			return err
		}
		loaded.Nested.URL = "https://single.example"
		loaded.Nested.Enabled = true // Excluded sibling change must stay local.
		if err := actor.Mutate(ctx, &loaded, keystone.MutateProperties("nested.url")); err != nil {
			return err
		}
		expected.Nested = &endpoint{URL: loaded.Nested.URL, Enabled: false}
		return assertStored()
	}) {
		return
	}
	if !run("Write explicit zero values", func() error {
		if _, err := client.Mutate(ctx, keystone.DynamicRecordPatch{Set: map[string]any{
			"display_name": "", "enabled": false, "retries": int64(0), "ratio": float64(0),
		}}); err != nil {
			return err
		}
		expected.Name, expected.Enabled, expected.Retries, expected.Ratio = "", false, 0, 0
		return assertStored()
	}) {
		return
	}
	if !run("Remove property and clear missing selected field", func() error {
		if _, err := client.Remove(ctx, "display_name"); err != nil {
			return err
		}
		raw, err := client.Get(ctx, "display_name")
		if err != nil {
			return err
		}
		if !raw.GetExists() || len(raw.GetProperties()) != 0 {
			return fmt.Errorf("removed property still returned: %v", raw)
		}
		got := settings{Name: "stale", Retries: 88}
		if err := client.Load(ctx, &got, "display_name"); err != nil {
			return err
		}
		if err := equalSettings(got, settings{Retries: 88}); err != nil {
			return err
		}
		return assertStored()
	}) {
		return
	}
	if !run("Remove nested typed properties", func() error {
		var loaded settings
		if err := client.Load(ctx, &loaded); err != nil {
			return err
		}
		loaded.Nested = nil
		if err := actor.Mutate(ctx, &loaded); err != nil {
			return err
		}
		expected.Nested = nil
		return assertStored()
	}) {
		return
	}
	if !run("Delete record", func() error {
		if _, err := client.Delete(ctx, "Requirement delete"); err != nil {
			return err
		}
		deleted = true
		raw, err := client.Get(ctx)
		if err != nil {
			return err
		}
		if raw.GetExists() || len(raw.GetProperties()) != 0 {
			return fmt.Errorf("deleted record still exists: %v", raw)
		}
		got := expected
		if err := client.Load(ctx, &got); err != nil {
			return err
		}
		if got.DynamicRecordExists() {
			return fmt.Errorf("typed deleted record still exists")
		}
		return equalSettings(got, settings{})
	}) {
		return
	}
}

func equalSettings(got, want settings) error {
	// Tracking metadata is checked separately; compare every application field.
	got.DynamicRecordEntity = keystone.DynamicRecordEntity{}
	want.DynamicRecordEntity = keystone.DynamicRecordEntity{}
	if !reflect.DeepEqual(got, want) {
		return fmt.Errorf("record fields mismatch: got %+v, want %+v", got, want)
	}
	return nil
}
