package keystone

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/keystonedb/sdk-go/proto"
	pb "google.golang.org/protobuf/proto"
)

// DynamicRecordEntity is embedded in a typed struct backed by a schema-free
// Dynamic Record. Its fields use the usual keystone tags and value converters.
// Use Actor.GetDynamicRecord to load and Actor.Mutate to save. No registration
// is required. After a load/save, only changed fields are submitted by default.
// Instances are mutable and should not be shared across concurrent operations.
type DynamicRecordEntity struct {
	key      string
	exists   bool
	revision string
	baseline map[Property]*proto.Value
}

func NewDynamicRecordEntity(key string) DynamicRecordEntity              { return DynamicRecordEntity{key: key} }
func (d *DynamicRecordEntity) dynamicRecordEntity() *DynamicRecordEntity { return d }
func (d *DynamicRecordEntity) GetDynamicRecordKey() string               { return d.key }

// SetDynamicRecordKey selects a different record and resets tracking metadata.
// The containing struct's application fields remain unchanged.
func (d *DynamicRecordEntity) SetDynamicRecordKey(key string) {
	if d.key != key {
		*d = NewDynamicRecordEntity(key)
	}
}
func (d *DynamicRecordEntity) DynamicRecordExists() bool     { return d.exists }
func (d *DynamicRecordEntity) DynamicRecordRevision() string { return d.revision }

type dynamicRecordEntityProvider interface{ dynamicRecordEntity() *DynamicRecordEntity }

func dynamicRecordEntityBase(src any) (*DynamicRecordEntity, error) {
	v := reflect.ValueOf(src)
	if !v.IsValid() || v.Kind() != reflect.Pointer || v.IsNil() || v.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("dynamic record entity requires a non-nil pointer to a struct")
	}
	e, ok := src.(dynamicRecordEntityProvider)
	if !ok {
		return nil, fmt.Errorf("struct must embed DynamicRecordEntity")
	}
	d := e.dynamicRecordEntity()
	if d == nil {
		return nil, fmt.Errorf("DynamicRecordEntity must be initialized")
	}
	return d, nil
}

// GetDynamicRecord loads a typed entity by a known app key, without a schema.
// No selection loads all fields. Selected fields absent from storage are cleared;
// unselected fields are preserved. Inspect DynamicRecordExists for a missing key.
func (a *Actor) GetDynamicRecord(ctx context.Context, key string, dst any, properties ...string) error {
	return a.DynamicRecord(key).Load(ctx, dst, properties...)
}
func (r *DynamicRecordClient) Load(ctx context.Context, dst any, properties ...string) error {
	base, err := dynamicRecordEntityBase(dst)
	if err != nil {
		return err
	}
	response, err := r.Get(ctx, properties...)
	if err != nil {
		return err
	}
	if response == nil {
		return fmt.Errorf("dynamic record: nil response")
	}
	// Hydrate a copy so a conversion error cannot leave the destination's metadata
	// or fields partially updated. Unselected values remain on that copy.
	val := reflect.ValueOf(dst).Elem()
	copyPtr := reflect.New(val.Type())
	copyPtr.Elem().Set(val)
	if err := clearDynamicRecordFields(copyPtr.Elem(), "", properties); err != nil {
		return err
	}
	values := map[Property]*proto.Value{}
	for name, value := range response.GetProperties() {
		parts := strings.Split(name, ".")
		values[knownPrefixProperty(strings.Join(parts[:len(parts)-1], "."), parts[len(parts)-1])] = value
	}
	if err := UnmarshalProperties(values, copyPtr.Interface()); err != nil {
		return err
	}
	baseline, err := Marshal(copyPtr.Interface())
	if err != nil {
		return err
	}
	val.Set(copyPtr.Elem())
	// Reacquire embedded pointers after setting the copied struct.
	base, err = dynamicRecordEntityBase(dst)
	if err != nil {
		return err
	}
	base.key = r.key
	base.exists = response.Exists
	base.revision = response.Revision
	base.baseline = cloneDynamicRecordProperties(baseline)
	return nil
}

// MutateDynamicRecord saves a typed entity. The key comes from its embedded
// DynamicRecordEntity. Supports MutateProperties and WithMutationComment only;
// entity/schema operations are rejected rather than silently ignored.
func (a *Actor) MutateDynamicRecord(ctx context.Context, src any, options ...MutateOption) error {
	base, err := dynamicRecordEntityBase(src)
	if err != nil {
		return err
	}
	current, err := Marshal(src)
	if err != nil {
		return err
	}
	var selected []string
	var comment string
	for _, option := range options {
		switch opt := option.(type) {
		case mutateProperties:
			selected = append(selected, opt.Property...)
		case withMutationComment:
			comment = opt.Comment
		default:
			return fmt.Errorf("dynamic records do not support mutation option %T", option)
		}
	}
	matches := func(name string) bool { return dynamicRecordPropertySelected(name, selected) }
	patch := DynamicRecordPatch{Set: map[string]any{}, Comment: comment}
	for prop, value := range current {
		if !matches(prop.Name()) {
			continue
		}
		if len(selected) == 0 && base.baseline != nil && pb.Equal(base.baseline[prop], value) {
			continue
		}
		patch.Set[prop.Name()] = value
	}
	for prop := range base.baseline {
		if _, present := current[prop]; !present && matches(prop.Name()) {
			patch.Remove = append(patch.Remove, prop.Name())
		}
	}
	// Explicit selection of an omitted nil/zero ValueMarshaler clears its property.
	for _, name := range selected {
		found := false
		for prop := range current {
			if prop.Name() == name || strings.HasPrefix(prop.Name(), name+".") {
				found = true
				break
			}
		}
		if !found {
			if !dynamicRecordHasField(reflect.TypeOf(src).Elem(), "", name) {
				return fmt.Errorf("unknown dynamic record property %q", name)
			}
			present := false
			for _, removed := range patch.Remove {
				if removed == name {
					present = true
				}
			}
			if !present {
				patch.Remove = append(patch.Remove, name)
			}
		}
	}
	// Validate actor/key even if there are no changes.
	client := a.DynamicRecord(base.key)
	if _, err := client.authorization(); err != nil {
		return err
	}
	if len(patch.Set) == 0 && len(patch.Remove) == 0 {
		return nil
	}
	response, err := client.Mutate(ctx, patch)
	if err != nil {
		return err
	}
	if response == nil {
		return fmt.Errorf("dynamic record: nil mutation response")
	}
	if base.baseline == nil {
		base.baseline = map[Property]*proto.Value{}
	}
	for prop, value := range current {
		if _, written := patch.Set[prop.Name()]; written {
			base.baseline[prop] = pb.Clone(value).(*proto.Value)
		}
	}
	for prop := range base.baseline {
		for _, name := range patch.Remove {
			if prop.Name() == name {
				delete(base.baseline, prop)
			}
		}
	}
	base.exists = true
	base.revision = response.Revision
	return nil
}
func dynamicRecordPropertySelected(name string, selection []string) bool {
	if len(selection) == 0 {
		return true
	}
	for _, prefix := range selection {
		if name == prefix || strings.HasPrefix(name, prefix+".") {
			return true
		}
	}
	return false
}
func cloneDynamicRecordProperties(values map[Property]*proto.Value) map[Property]*proto.Value {
	out := make(map[Property]*proto.Value, len(values))
	for key, value := range values {
		out[key] = pb.Clone(value).(*proto.Value)
	}
	return out
}
func dynamicRecordHasField(t reflect.Type, prefix, name string) bool {
	for _, field := range reflect.VisibleFields(t) {
		if field.Anonymous || !field.IsExported() {
			continue
		}
		prop, _ := ReflectProperty(field, "")
		if prop.HydrateOnly() {
			continue
		}
		full := prefix + prop.Name()
		if full == name {
			return true
		}
		ft := field.Type
		if ft.Kind() == reflect.Pointer {
			ft = ft.Elem()
		}
		if strings.HasPrefix(name, full+".") && ft.Kind() == reflect.Struct && dynamicRecordHasField(ft, full+".", name) {
			return true
		}
	}
	return false
}

// Clone nested structs before clearing/hydrating so pointer fields on the original
// are preserved if decoding fails. ValueMarshaler pointers are replaced on clear.
func clearDynamicRecordFields(v reflect.Value, prefix string, selection []string) error {
	for _, field := range reflect.VisibleFields(v.Type()) {
		if field.Anonymous || !field.IsExported() {
			continue
		}
		prop, _ := ReflectProperty(field, "")
		if prop.HydrateOnly() {
			continue
		}
		name := prefix + prop.Name()
		current := v.FieldByIndex(field.Index)
		if dynamicRecordPropertySelected(name, selection) {
			current.SetZero()
			continue
		}
		child := false
		for _, s := range selection {
			if strings.HasPrefix(s, name+".") {
				child = true
				break
			}
		}
		if !child {
			continue
		}
		if current.Kind() == reflect.Pointer && current.Type().Elem().Kind() == reflect.Struct {
			next := reflect.New(current.Type().Elem())
			if !current.IsNil() {
				next.Elem().Set(current.Elem())
			}
			current.Set(next)
			if err := clearDynamicRecordFields(next.Elem(), name+".", selection); err != nil {
				return err
			}
		} else if current.Kind() == reflect.Struct {
			if err := clearDynamicRecordFields(current, name+".", selection); err != nil {
				return err
			}
		}
	}
	return nil
}

func dynamicRecordEntityType(t reflect.Type) bool {
	if t == nil {
		return false
	}
	marker := reflect.TypeOf((*dynamicRecordEntityProvider)(nil)).Elem()
	if t.Implements(marker) {
		return true
	}
	return t.Kind() == reflect.Struct && reflect.PointerTo(t).Implements(marker)
}
