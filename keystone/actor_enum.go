package keystone

import (
	"context"
	"errors"

	"github.com/keystonedb/sdk-go/proto"
)

// EnumPut creates or updates a single enum entry
func (a *Actor) EnumPut(ctx context.Context, enumType, key, name, description string, metadata map[string]string) error {
	if a == nil || a.connection == nil {
		return errors.New("actor or connection is nil")
	}
	resp, err := a.connection.EnumPut(ctx, &proto.EnumPutRequest{
		Authorization: a.Authorization(),
		Enum: &proto.EnumEntry{
			Type: enumType, Key: key, Name: name, Description: description, Metadata: metadata,
		},
	})
	if err != nil {
		return err
	}
	if !resp.GetSuccess() {
		return errors.New(resp.GetErrorMessage())
	}
	return nil
}

// EnumGet retrieves a single enum entry by type and key
func (a *Actor) EnumGet(ctx context.Context, enumType, key string) (*proto.EnumEntry, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}
	resp, err := a.connection.EnumGet(ctx, &proto.EnumGetRequest{
		Authorization: a.Authorization(), Type: enumType, Key: key,
	})
	if err != nil {
		return nil, err
	}
	if !resp.GetSummary().GetSuccess() {
		return nil, errors.New(resp.GetSummary().GetErrorMessage())
	}
	return resp.GetEnum(), nil
}

// EnumDelete deletes enum entries. If key is empty, deletes the entire type.
func (a *Actor) EnumDelete(ctx context.Context, enumType, key string) error {
	if a == nil || a.connection == nil {
		return errors.New("actor or connection is nil")
	}
	resp, err := a.connection.EnumDelete(ctx, &proto.EnumDeleteRequest{
		Authorization: a.Authorization(), Type: enumType, Key: key,
	})
	if err != nil {
		return err
	}
	if !resp.GetSuccess() {
		return errors.New(resp.GetErrorMessage())
	}
	return nil
}

// EnumList returns all enum entries for a given type
func (a *Actor) EnumList(ctx context.Context, enumType string) ([]*proto.EnumEntry, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}
	resp, err := a.connection.EnumList(ctx, &proto.EnumListRequest{
		Authorization: a.Authorization(), Type: enumType,
	})
	if err != nil {
		return nil, err
	}
	if !resp.GetSummary().GetSuccess() {
		return nil, errors.New(resp.GetSummary().GetErrorMessage())
	}
	return resp.GetEnums(), nil
}

// EnumReplace replaces all entries for a type with the provided entries (diff-based on server)
func (a *Actor) EnumReplace(ctx context.Context, enumType string, entries []*proto.EnumEntry) error {
	if a == nil || a.connection == nil {
		return errors.New("actor or connection is nil")
	}
	resp, err := a.connection.EnumReplace(ctx, &proto.EnumReplaceRequest{
		Authorization: a.Authorization(), Type: enumType, Enums: entries,
	})
	if err != nil {
		return err
	}
	if !resp.GetSuccess() {
		return errors.New(resp.GetErrorMessage())
	}
	return nil
}
