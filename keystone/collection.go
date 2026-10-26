package keystone

import (
	"context"
	"fmt"
	"iter"
	"reflect"

	"github.com/keystonedb/sdk-go/proto"
	gproto "google.golang.org/protobuf/proto"
)

// EntityCollection owns typed query results and their cursor pagination state.
// It is intended for sequential use; concurrent access requires synchronization.
// Use Collection to construct one.
type EntityCollection[T any] struct {
	items    []T
	response *proto.QueryResponse
	request  *proto.QueryRequest
	actor    *Actor
}

// Collection creates an empty collection of T bound to actor. The example supplies only the
// type; its field values are not defaults or filters. T must be a struct or a
// pointer to a struct. Invalid types are reported by Query.
func Collection[T any](_ T, actor *Actor) *EntityCollection[T] {
	return &EntityCollection[T]{actor: actor}
}

// Items returns all entities loaded by Query and subsequent Next calls, in query
// order. The slice belongs to the collection and is replaced by a successful Query.
func (c *EntityCollection[T]) Items() []T { return c.items }

// All iterates over the loaded entities in query order. It does not fetch pages.
// The loaded slice is captured when iteration starts. Do not call Query or Next
// during iteration.
func (c *EntityCollection[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, item := range c.items {
			if !yield(item) {
				return
			}
		}
	}
}

// Response returns the most recent successful page, including totals, execution
// details. It is nil before the first successful Query.
// The response belongs to the collection and should be treated as read-only.
func (c *EntityCollection[T]) Response() *proto.QueryResponse { return c.response }

func (c *EntityCollection[T]) HasMore() bool { return c.response.GetHasMore() }

func (c *EntityCollection[T]) NextCursor() string { return c.response.GetNextCursor() }

// Latest replaces the collection with up to size entities, newest by creation
// time first. Size must be positive. Options configure projections and filters,
// as in Query; additional sorting breaks creation-time ties. Latest always starts
// at the first page and overrides any Cursor or Limit size option.
// Without Select or SelectSecure, only entity metadata is loaded.
func (c *EntityCollection[T]) Latest(ctx context.Context, size int32, options ...QueryOption) error {
	if size <= 0 {
		return fmt.Errorf("collection: latest size must be positive")
	}
	queryOptions := make([]QueryOption, 0, len(options)+3)
	queryOptions = append(queryOptions, SortDesc("_created"))
	queryOptions = append(queryOptions, options...)
	queryOptions = append(queryOptions, Limit(size), Cursor(""))
	return c.Query(ctx, queryOptions...)
}

// Query loads the first page and replaces the collection's results and query
// state on success. It uses Actor.Query options, including explicit Select or
// SelectSecure projections; without a projection only metadata is loaded.
// Errors leave the previous results and pagination state intact.
func (c *EntityCollection[T]) Query(ctx context.Context, options ...QueryOption) error {
	actor := c.actor
	t := reflect.TypeFor[T]()
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("collection: expected a struct or pointer to a struct, got %v", reflect.TypeFor[T]())
	}
	if actor == nil || actor.connection == nil {
		return fmt.Errorf("collection: actor must have a connection")
	}
	request, err := actor.queryRequest(QuickDefine(reflect.New(t).Interface()).Type, nil, options...)
	if err != nil {
		return err
	}
	if len(request.GroupBy) > 0 || len(request.Aggregates) > 0 {
		return fmt.Errorf("collection: aggregate queries require QueryPages and QueryRows")
	}
	request = gproto.Clone(request).(*proto.QueryRequest)
	response, items, err := c.load(ctx, actor, request)
	if err != nil {
		return err
	}
	c.actor, c.request, c.response, c.items = actor, request, response, items
	return nil
}

// Next appends the next page using the last successful Query's filters,
// projection, ordering and actor. It returns io.EOF before Query or when no more
// pages remain. Errors leave loaded results and the cursor intact for retry.
func (c *EntityCollection[T]) Next(ctx context.Context) error {
	request, err := nextQueryRequest(c.request, c.response)
	if err != nil {
		return err
	}
	response, items, err := c.load(ctx, c.actor, request)
	if err != nil {
		return err
	}
	c.items = append(c.items, items...)
	c.request, c.response = request, response
	return nil
}

func (c *EntityCollection[T]) load(ctx context.Context, actor *Actor, request *proto.QueryRequest) (*proto.QueryResponse, []T, error) {
	// Keep a snapshot independent of caller-owned option values and RPC changes.
	response, err := actor.connection.Query(ctx, gproto.Clone(request).(*proto.QueryRequest))
	if err != nil {
		return nil, nil, err
	}
	if response == nil {
		return nil, nil, fmt.Errorf("collection: nil query response")
	}
	if len(response.Rows) > 0 {
		return nil, nil, fmt.Errorf("collection: aggregate rows require QueryRows")
	}
	for _, entity := range response.Entities {
		if entity == nil {
			return nil, nil, ErrEntityIDMissing
		}
	}
	items, err := AsSlice[T](response.Entities...)
	if err != nil {
		return nil, nil, fmt.Errorf("collection: %w", err)
	}
	return response, items, nil
}
