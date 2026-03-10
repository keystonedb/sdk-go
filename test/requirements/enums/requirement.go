package enums

import (
	"context"
	"errors"
	"fmt"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/keystonedb/sdk-go/test/requirements"
)

const (
	typeColors    = "enum-test-colors"
	typeSizes     = "enum-test-sizes"
	typePriority  = "enum-test-priority"
	typeIsolation = "enum-test-isolation"
)

type Requirement struct{}

func (d *Requirement) Name() string {
	return "Custom Enums"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	// Clean up any leftover data from previous runs
	d.cleanup(actor)

	return []requirements.TestResult{
		// Basic CRUD on a single type
		d.putAndGet(actor),
		d.putWithMetadata(actor),
		d.getNotFound(actor),
		d.update(actor),
		d.list(actor),

		// Multiple types coexist independently
		d.multipleTypes(actor),

		// Cross-type isolation: operations on one type must not affect another
		d.typeIsolationList(actor),
		d.typeIsolationDelete(actor),
		d.typeIsolationReplace(actor),

		// Same key in different types must be independent
		d.sameKeyDifferentTypes(actor),

		// Replace within a type
		d.replace(actor),

		// Delete single key leaves others intact
		d.deleteKey(actor),

		// Delete entire type
		d.deleteType(actor),

		// Verify empty list for unknown type
		d.listEmptyType(actor),
	}
}

func (d *Requirement) cleanup(actor *keystone.Actor) {
	ctx := context.Background()
	_ = actor.EnumDelete(ctx, typeColors, "")
	_ = actor.EnumDelete(ctx, typeSizes, "")
	_ = actor.EnumDelete(ctx, typePriority, "")
	_ = actor.EnumDelete(ctx, typeIsolation, "")
}

func (d *Requirement) putAndGet(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Put and Get"}

	err := actor.EnumPut(context.Background(), typeColors, "red", "Red", "A warm color", nil)
	if err != nil {
		return res.WithError(fmt.Errorf("put failed: %w", err))
	}

	entry, err := actor.EnumGet(context.Background(), typeColors, "red")
	if err != nil {
		return res.WithError(fmt.Errorf("get failed: %w", err))
	}

	if entry.GetKey() != "red" {
		return res.WithError(fmt.Errorf("expected key 'red', got '%s'", entry.GetKey()))
	}
	if entry.GetName() != "Red" {
		return res.WithError(fmt.Errorf("expected name 'Red', got '%s'", entry.GetName()))
	}
	if entry.GetDescription() != "A warm color" {
		return res.WithError(fmt.Errorf("expected description 'A warm color', got '%s'", entry.GetDescription()))
	}
	if entry.GetType() != typeColors {
		return res.WithError(fmt.Errorf("expected type '%s', got '%s'", typeColors, entry.GetType()))
	}

	return res
}

func (d *Requirement) putWithMetadata(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Put with Metadata"}

	meta := map[string]string{"hex": "#0000FF", "rgb": "0,0,255"}
	err := actor.EnumPut(context.Background(), typeColors, "blue", "Blue", "A cool color", meta)
	if err != nil {
		return res.WithError(fmt.Errorf("put failed: %w", err))
	}

	entry, err := actor.EnumGet(context.Background(), typeColors, "blue")
	if err != nil {
		return res.WithError(fmt.Errorf("get failed: %w", err))
	}

	if entry.GetName() != "Blue" {
		return res.WithError(fmt.Errorf("expected name 'Blue', got '%s'", entry.GetName()))
	}

	gotMeta := entry.GetMetadata()
	if gotMeta == nil {
		return res.WithError(errors.New("metadata is nil"))
	}
	if gotMeta["hex"] != "#0000FF" {
		return res.WithError(fmt.Errorf("expected hex '#0000FF', got '%s'", gotMeta["hex"]))
	}
	if gotMeta["rgb"] != "0,0,255" {
		return res.WithError(fmt.Errorf("expected rgb '0,0,255', got '%s'", gotMeta["rgb"]))
	}

	return res
}

func (d *Requirement) getNotFound(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Get Not Found"}

	_, err := actor.EnumGet(context.Background(), typeColors, "nonexistent-key")
	if err == nil {
		return res.WithError(errors.New("expected error for nonexistent key, got nil"))
	}

	return res
}

func (d *Requirement) update(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Update"}

	err := actor.EnumPut(context.Background(), typeColors, "red", "Crimson Red", "An updated warm color", nil)
	if err != nil {
		return res.WithError(fmt.Errorf("put update failed: %w", err))
	}

	entry, err := actor.EnumGet(context.Background(), typeColors, "red")
	if err != nil {
		return res.WithError(fmt.Errorf("get after update failed: %w", err))
	}

	if entry.GetName() != "Crimson Red" {
		return res.WithError(fmt.Errorf("expected name 'Crimson Red', got '%s'", entry.GetName()))
	}
	if entry.GetDescription() != "An updated warm color" {
		return res.WithError(fmt.Errorf("expected description 'An updated warm color', got '%s'", entry.GetDescription()))
	}

	return res
}

func (d *Requirement) list(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "List"}

	// red and blue should exist from previous tests
	entries, err := actor.EnumList(context.Background(), typeColors)
	if err != nil {
		return res.WithError(fmt.Errorf("list failed: %w", err))
	}

	if len(entries) < 2 {
		return res.WithError(fmt.Errorf("expected at least 2 entries, got %d", len(entries)))
	}

	keys := entryKeys(entries)
	if !keys["red"] {
		return res.WithError(errors.New("'red' not found in list"))
	}
	if !keys["blue"] {
		return res.WithError(errors.New("'blue' not found in list"))
	}

	// Verify every returned entry has the correct type
	for _, e := range entries {
		if e.GetType() != typeColors {
			return res.WithError(fmt.Errorf("entry '%s' has type '%s', expected '%s'", e.GetKey(), e.GetType(), typeColors))
		}
	}

	return res
}

// multipleTypes creates entries across three independent types and verifies each type
// only contains its own entries
func (d *Requirement) multipleTypes(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Multiple Types"}
	ctx := context.Background()

	// Populate sizes type
	for _, entry := range []struct{ key, name string }{
		{"s", "Small"}, {"m", "Medium"}, {"l", "Large"},
	} {
		if err := actor.EnumPut(ctx, typeSizes, entry.key, entry.name, "", nil); err != nil {
			return res.WithError(fmt.Errorf("put size '%s' failed: %w", entry.key, err))
		}
	}

	// Populate priority type
	for _, entry := range []struct{ key, name string }{
		{"low", "Low"}, {"med", "Medium"}, {"high", "High"}, {"critical", "Critical"},
	} {
		if err := actor.EnumPut(ctx, typePriority, entry.key, entry.name, "", nil); err != nil {
			return res.WithError(fmt.Errorf("put priority '%s' failed: %w", entry.key, err))
		}
	}

	// Verify sizes has exactly 3
	sizeEntries, err := actor.EnumList(ctx, typeSizes)
	if err != nil {
		return res.WithError(fmt.Errorf("list sizes failed: %w", err))
	}
	if len(sizeEntries) != 3 {
		return res.WithError(fmt.Errorf("expected 3 size entries, got %d", len(sizeEntries)))
	}

	// Verify priority has exactly 4
	prioEntries, err := actor.EnumList(ctx, typePriority)
	if err != nil {
		return res.WithError(fmt.Errorf("list priority failed: %w", err))
	}
	if len(prioEntries) != 4 {
		return res.WithError(fmt.Errorf("expected 4 priority entries, got %d", len(prioEntries)))
	}

	// Verify colors still has its entries and was not affected
	colorEntries, err := actor.EnumList(ctx, typeColors)
	if err != nil {
		return res.WithError(fmt.Errorf("list colors failed: %w", err))
	}
	colorKeys := entryKeys(colorEntries)
	if !colorKeys["red"] || !colorKeys["blue"] {
		return res.WithError(errors.New("color entries were affected by adding other types"))
	}

	// Verify no cross-type contamination: size keys must not appear in priority, etc.
	sizeKeys := entryKeys(sizeEntries)
	prioKeys := entryKeys(prioEntries)
	for k := range sizeKeys {
		if prioKeys[k] {
			return res.WithError(fmt.Errorf("size key '%s' leaked into priority type", k))
		}
		if colorKeys[k] {
			return res.WithError(fmt.Errorf("size key '%s' leaked into colors type", k))
		}
	}
	for k := range prioKeys {
		if sizeKeys[k] {
			return res.WithError(fmt.Errorf("priority key '%s' leaked into sizes type", k))
		}
		if colorKeys[k] {
			return res.WithError(fmt.Errorf("priority key '%s' leaked into colors type", k))
		}
	}

	return res
}

// typeIsolationList verifies that listing one type never returns entries from another
func (d *Requirement) typeIsolationList(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Type Isolation - List"}
	ctx := context.Background()

	// List each type and verify every entry has the correct type field
	for _, enumType := range []string{typeColors, typeSizes, typePriority} {
		entries, err := actor.EnumList(ctx, enumType)
		if err != nil {
			return res.WithError(fmt.Errorf("list '%s' failed: %w", enumType, err))
		}
		for _, e := range entries {
			if e.GetType() != enumType {
				return res.WithError(fmt.Errorf("entry '%s' in list for '%s' has type '%s'", e.GetKey(), enumType, e.GetType()))
			}
		}
	}

	return res
}

// typeIsolationDelete verifies that deleting entries in one type does not affect other types
func (d *Requirement) typeIsolationDelete(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Type Isolation - Delete"}
	ctx := context.Background()

	// Snapshot counts before delete
	sizesBefore, err := actor.EnumList(ctx, typeSizes)
	if err != nil {
		return res.WithError(fmt.Errorf("list sizes before failed: %w", err))
	}
	colorsBefore, err := actor.EnumList(ctx, typeColors)
	if err != nil {
		return res.WithError(fmt.Errorf("list colors before failed: %w", err))
	}

	// Delete one priority entry
	err = actor.EnumDelete(ctx, typePriority, "low")
	if err != nil {
		return res.WithError(fmt.Errorf("delete priority 'low' failed: %w", err))
	}

	// Verify priority lost one entry
	prioAfter, err := actor.EnumList(ctx, typePriority)
	if err != nil {
		return res.WithError(fmt.Errorf("list priority after failed: %w", err))
	}
	if entryKeys(prioAfter)["low"] {
		return res.WithError(errors.New("'low' still present in priority after delete"))
	}

	// Verify sizes and colors are unchanged
	sizesAfter, err := actor.EnumList(ctx, typeSizes)
	if err != nil {
		return res.WithError(fmt.Errorf("list sizes after failed: %w", err))
	}
	if len(sizesAfter) != len(sizesBefore) {
		return res.WithError(fmt.Errorf("sizes count changed from %d to %d after deleting from priority", len(sizesBefore), len(sizesAfter)))
	}

	colorsAfter, err := actor.EnumList(ctx, typeColors)
	if err != nil {
		return res.WithError(fmt.Errorf("list colors after failed: %w", err))
	}
	if len(colorsAfter) != len(colorsBefore) {
		return res.WithError(fmt.Errorf("colors count changed from %d to %d after deleting from priority", len(colorsBefore), len(colorsAfter)))
	}

	return res
}

// typeIsolationReplace verifies that replacing one type does not affect other types
func (d *Requirement) typeIsolationReplace(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Type Isolation - Replace"}
	ctx := context.Background()

	// Snapshot other types before replace
	colorsBefore, err := actor.EnumList(ctx, typeColors)
	if err != nil {
		return res.WithError(fmt.Errorf("list colors before failed: %w", err))
	}
	prioBefore, err := actor.EnumList(ctx, typePriority)
	if err != nil {
		return res.WithError(fmt.Errorf("list priority before failed: %w", err))
	}

	// Replace sizes with entirely new entries
	newSizes := []*proto.EnumEntry{
		{Type: typeSizes, Key: "xs", Name: "Extra Small"},
		{Type: typeSizes, Key: "xxl", Name: "Extra Extra Large"},
	}
	err = actor.EnumReplace(ctx, typeSizes, newSizes)
	if err != nil {
		return res.WithError(fmt.Errorf("replace sizes failed: %w", err))
	}

	// Verify sizes were replaced
	sizesAfter, err := actor.EnumList(ctx, typeSizes)
	if err != nil {
		return res.WithError(fmt.Errorf("list sizes after failed: %w", err))
	}
	sizeKeys := entryKeys(sizesAfter)
	if !sizeKeys["xs"] || !sizeKeys["xxl"] {
		return res.WithError(errors.New("replaced size entries not found"))
	}
	if sizeKeys["s"] || sizeKeys["m"] || sizeKeys["l"] {
		return res.WithError(errors.New("old size entries still present after replace"))
	}

	// Verify colors and priority were not affected
	colorsAfter, err := actor.EnumList(ctx, typeColors)
	if err != nil {
		return res.WithError(fmt.Errorf("list colors after failed: %w", err))
	}
	if len(colorsAfter) != len(colorsBefore) {
		return res.WithError(fmt.Errorf("colors count changed from %d to %d after replacing sizes", len(colorsBefore), len(colorsAfter)))
	}

	prioAfter, err := actor.EnumList(ctx, typePriority)
	if err != nil {
		return res.WithError(fmt.Errorf("list priority after failed: %w", err))
	}
	if len(prioAfter) != len(prioBefore) {
		return res.WithError(fmt.Errorf("priority count changed from %d to %d after replacing sizes", len(prioBefore), len(prioAfter)))
	}

	return res
}

// sameKeyDifferentTypes verifies that the same key in two types stores independent data
func (d *Requirement) sameKeyDifferentTypes(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Same Key Different Types"}
	ctx := context.Background()

	typeA := typeIsolation + "-a"
	typeB := typeIsolation + "-b"

	// Cleanup at start and end
	defer func() {
		_ = actor.EnumDelete(ctx, typeA, "")
		_ = actor.EnumDelete(ctx, typeB, "")
	}()
	_ = actor.EnumDelete(ctx, typeA, "")
	_ = actor.EnumDelete(ctx, typeB, "")

	// Put same key "active" in both types with different values
	err := actor.EnumPut(ctx, typeA, "active", "Active User", "An active user status", nil)
	if err != nil {
		return res.WithError(fmt.Errorf("put typeA failed: %w", err))
	}
	err = actor.EnumPut(ctx, typeB, "active", "Active Subscription", "An active subscription status", nil)
	if err != nil {
		return res.WithError(fmt.Errorf("put typeB failed: %w", err))
	}

	// Get from typeA - should have typeA's values
	entryA, err := actor.EnumGet(ctx, typeA, "active")
	if err != nil {
		return res.WithError(fmt.Errorf("get typeA failed: %w", err))
	}
	if entryA.GetName() != "Active User" {
		return res.WithError(fmt.Errorf("typeA: expected name 'Active User', got '%s'", entryA.GetName()))
	}
	if entryA.GetDescription() != "An active user status" {
		return res.WithError(fmt.Errorf("typeA: expected description 'An active user status', got '%s'", entryA.GetDescription()))
	}

	// Get from typeB - should have typeB's values
	entryB, err := actor.EnumGet(ctx, typeB, "active")
	if err != nil {
		return res.WithError(fmt.Errorf("get typeB failed: %w", err))
	}
	if entryB.GetName() != "Active Subscription" {
		return res.WithError(fmt.Errorf("typeB: expected name 'Active Subscription', got '%s'", entryB.GetName()))
	}
	if entryB.GetDescription() != "An active subscription status" {
		return res.WithError(fmt.Errorf("typeB: expected description 'An active subscription status', got '%s'", entryB.GetDescription()))
	}

	// Delete from typeA should not affect typeB
	err = actor.EnumDelete(ctx, typeA, "active")
	if err != nil {
		return res.WithError(fmt.Errorf("delete typeA failed: %w", err))
	}

	// typeB should still have its entry
	entryB2, err := actor.EnumGet(ctx, typeB, "active")
	if err != nil {
		return res.WithError(fmt.Errorf("get typeB after typeA delete failed: %w", err))
	}
	if entryB2.GetName() != "Active Subscription" {
		return res.WithError(errors.New("typeB entry was affected by deleting same key from typeA"))
	}

	return res
}

func (d *Requirement) replace(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Replace"}
	ctx := context.Background()

	replaceType := typeSizes + "-replace"
	defer func() { _ = actor.EnumDelete(ctx, replaceType, "") }()

	// Put initial entries
	for _, entry := range []struct{ key, name string }{
		{"s", "Small"}, {"m", "Medium"},
	} {
		if err := actor.EnumPut(ctx, replaceType, entry.key, entry.name, "", nil); err != nil {
			return res.WithError(fmt.Errorf("initial put failed: %w", err))
		}
	}

	// Replace with a new set
	newEntries := []*proto.EnumEntry{
		{Type: replaceType, Key: "l", Name: "Large", Description: "Large size"},
		{Type: replaceType, Key: "xl", Name: "Extra Large", Description: "Extra large size"},
	}
	err := actor.EnumReplace(ctx, replaceType, newEntries)
	if err != nil {
		return res.WithError(fmt.Errorf("replace failed: %w", err))
	}

	entries, err := actor.EnumList(ctx, replaceType)
	if err != nil {
		return res.WithError(fmt.Errorf("list after replace failed: %w", err))
	}

	keys := entryKeys(entries)
	if !keys["l"] || !keys["xl"] {
		return res.WithError(errors.New("new entries not found after replace"))
	}
	if keys["s"] || keys["m"] {
		return res.WithError(errors.New("old entries still present after replace"))
	}

	return res
}

func (d *Requirement) deleteKey(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Delete Key"}
	ctx := context.Background()

	delType := typeColors + "-delkey"
	defer func() { _ = actor.EnumDelete(ctx, delType, "") }()

	err := actor.EnumPut(ctx, delType, "toremove", "To Remove", "", nil)
	if err != nil {
		return res.WithError(fmt.Errorf("put failed: %w", err))
	}
	err = actor.EnumPut(ctx, delType, "tokeep", "To Keep", "", nil)
	if err != nil {
		return res.WithError(fmt.Errorf("put failed: %w", err))
	}

	// Delete single key
	err = actor.EnumDelete(ctx, delType, "toremove")
	if err != nil {
		return res.WithError(fmt.Errorf("delete key failed: %w", err))
	}

	entries, err := actor.EnumList(ctx, delType)
	if err != nil {
		return res.WithError(fmt.Errorf("list after delete failed: %w", err))
	}

	keys := entryKeys(entries)
	if keys["toremove"] {
		return res.WithError(errors.New("'toremove' should have been deleted"))
	}
	if !keys["tokeep"] {
		return res.WithError(errors.New("'tokeep' should still exist"))
	}

	return res
}

func (d *Requirement) deleteType(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Delete Type"}
	ctx := context.Background()

	delType := typeColors + "-deltype"

	err := actor.EnumPut(ctx, delType, "a", "Alpha", "", nil)
	if err != nil {
		return res.WithError(fmt.Errorf("put failed: %w", err))
	}
	err = actor.EnumPut(ctx, delType, "b", "Beta", "", nil)
	if err != nil {
		return res.WithError(fmt.Errorf("put failed: %w", err))
	}

	// Delete entire type
	err = actor.EnumDelete(ctx, delType, "")
	if err != nil {
		return res.WithError(fmt.Errorf("delete type failed: %w", err))
	}

	entries, err := actor.EnumList(ctx, delType)
	if err != nil {
		return res.WithError(fmt.Errorf("list after delete type failed: %w", err))
	}

	if len(entries) != 0 {
		return res.WithError(fmt.Errorf("expected 0 entries after type delete, got %d", len(entries)))
	}

	return res
}

func (d *Requirement) listEmptyType(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "List Empty Type"}

	entries, err := actor.EnumList(context.Background(), "enum-test-nonexistent-type")
	if err != nil {
		return res.WithError(fmt.Errorf("list empty type failed: %w", err))
	}

	if len(entries) != 0 {
		return res.WithError(fmt.Errorf("expected 0 entries for unknown type, got %d", len(entries)))
	}

	return res
}

func entryKeys(entries []*proto.EnumEntry) map[string]bool {
	keys := make(map[string]bool, len(entries))
	for _, e := range entries {
		keys[e.GetKey()] = true
	}
	return keys
}
