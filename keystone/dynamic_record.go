package keystone

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/keystonedb/sdk-go/proto"
	pb "google.golang.org/protobuf/proto"
)

// DynamicRecordPatch sets or removes named properties, preserving all others.
// Set accepts SDK-supported Go values or *proto.Value for explicit dynamic types.
type DynamicRecordPatch struct {
	Set     map[string]any
	Remove  []string
	Comment string
}
type DynamicRecordHistoryOptions struct {
	Limit          int32
	BeforeRevision string
}

// DynamicRecord selects a schema-free record without making a network request.
// The record is isolated by the actor's workspace, vendor and app.
func (a *Actor) DynamicRecord(key string) *DynamicRecordClient {
	return &DynamicRecordClient{actor: a, key: key}
}

type DynamicRecordClient struct {
	actor *Actor
	key   string
}

func (r *DynamicRecordClient) authorization() (*proto.Authorization, error) {
	if r == nil || r.actor == nil || r.actor.connection == nil {
		return nil, fmt.Errorf("dynamic record: actor must have a connection")
	}
	if len(r.key) == 0 || len(r.key) > 512 || strings.TrimSpace(r.key) != r.key {
		return nil, fmt.Errorf("dynamic record: key must contain 1–512 bytes without surrounding whitespace")
	}
	return pb.Clone(r.actor.Authorization()).(*proto.Authorization), nil
}

// Get reads selected properties; no selection reads all. Missing properties are
// omitted from Properties. Exists distinguishes missing records from empty ones.
func (r *DynamicRecordClient) Get(ctx context.Context, properties ...string) (*proto.DynamicRecord, error) {
	auth, err := r.authorization()
	if err != nil {
		return nil, err
	}
	return r.actor.connection.GetDynamicRecord(ctx, &proto.GetDynamicRecordRequest{Authorization: auth, Key: r.key, Properties: properties})
}

// Set replaces one property, without reading the current record.
func (r *DynamicRecordClient) Set(ctx context.Context, name string, value any) (*proto.DynamicRecordMutationResponse, error) {
	return r.Mutate(ctx, DynamicRecordPatch{Set: map[string]any{name: value}})
}
func (r *DynamicRecordClient) Remove(ctx context.Context, names ...string) (*proto.DynamicRecordMutationResponse, error) {
	return r.Mutate(ctx, DynamicRecordPatch{Remove: names})
}

// Mutate creates the record if absent and atomically patches the supplied names.
// Retries after an ambiguous transport failure can produce multiple history entries.
func (r *DynamicRecordClient) Mutate(ctx context.Context, patch DynamicRecordPatch) (*proto.DynamicRecordMutationResponse, error) {
	auth, err := r.authorization()
	if err != nil {
		return nil, err
	}
	values := make(map[string]*proto.Value, len(patch.Set))
	for name, value := range patch.Set {
		v, err := dynamicRecordValue(value)
		if err != nil {
			return nil, fmt.Errorf("dynamic record property %q: %w", name, err)
		}
		values[name] = v
	}
	return r.actor.connection.MutateDynamicRecord(ctx, &proto.MutateDynamicRecordRequest{Authorization: auth, Key: r.key, Set: values, Remove: patch.Remove, Comment: patch.Comment})
}
func (r *DynamicRecordClient) Delete(ctx context.Context, comment string) (*proto.DynamicRecordMutationResponse, error) {
	auth, err := r.authorization()
	if err != nil {
		return nil, err
	}
	return r.actor.connection.DeleteDynamicRecord(ctx, &proto.DeleteDynamicRecordRequest{Authorization: auth, Key: r.key, Comment: comment})
}
func (r *DynamicRecordClient) History(ctx context.Context, options DynamicRecordHistoryOptions) (*proto.ListDynamicRecordHistoryResponse, error) {
	auth, err := r.authorization()
	if err != nil {
		return nil, err
	}
	return r.actor.connection.ListDynamicRecordHistory(ctx, &proto.ListDynamicRecordHistoryRequest{Authorization: auth, Key: r.key, Limit: options.Limit, BeforeRevision: options.BeforeRevision})
}
func dynamicRecordValue(value any) (*proto.Value, error) {
	if v, ok := value.(*proto.Value); ok {
		if v == nil {
			return nil, fmt.Errorf("nil value")
		}
		return pb.Clone(v).(*proto.Value), nil
	}
	v := reflect.ValueOf(value)
	for v.IsValid() && (v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface) {
		if v.IsNil() {
			return nil, fmt.Errorf("nil value")
		}
		v = v.Elem()
	}
	if !v.IsValid() {
		return nil, fmt.Errorf("nil value; use Remove to delete a property")
	}
	ref := GetReflector(v.Type(), v)
	if ref == nil {
		return nil, fmt.Errorf("unsupported value type %s", v.Type())
	}
	out, err := ref.ToProto(v)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, fmt.Errorf("unsupported value")
	}
	out = pb.Clone(out).(*proto.Value)
	out.KnownType = ref.PropertyDefinition().DataType
	return out, nil
}
