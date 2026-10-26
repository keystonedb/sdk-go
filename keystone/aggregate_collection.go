package keystone

import (
	"context"
	"fmt"
	"iter"
	"reflect"

	"github.com/keystonedb/sdk-go/proto"
	gproto "google.golang.org/protobuf/proto"
)

// AggregateRow contains group keys and aggregate aliases. Values preserve their
// protobuf types and SQL NULL flags; a missing alias returns nil from Value.
// Rows belong to their collection and should be treated as read-only.
type AggregateRow map[string]*proto.Value

func (r AggregateRow) Value(alias string) *proto.Value { return r[alias] }

// AggregateResults owns aggregate rows and their cursor state. Next appends
// pages, Query replaces results on success, and All iterates loaded rows only.
// Like EntityCollection, it requires sequential access.
type AggregateResults struct {
	actor      *Actor
	entityType string
	request    *proto.QueryRequest
	response   *proto.QueryResponse
	items      []AggregateRow
}

// AggregateCollection binds an aggregate collection to an entity type and actor.
// No result struct is needed: group keys and aliases are available on each row.
func AggregateCollection(entityType string, actor *Actor) *AggregateResults {
	return &AggregateResults{actor: actor, entityType: entityType}
}

// Aggregate queries this collection's entity type and returns a separate
// aggregate collection. It does not change the entity collection or inherit its
// previous query's filters; supply all filters and aggregate options explicitly.
func (c *EntityCollection[T]) Aggregate(ctx context.Context, options ...QueryOption) (*AggregateResults, error) {
	t := reflect.TypeFor[T]()
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("collection: expected a struct or pointer to a struct, got %v", reflect.TypeFor[T]())
	}
	results := AggregateCollection(QuickDefine(reflect.New(t).Interface()).Type, c.actor)
	if err := results.Query(ctx, options...); err != nil {
		return nil, err
	}
	return results, nil
}

func (c *AggregateResults) Items() []AggregateRow          { return c.items }
func (c *AggregateResults) Response() *proto.QueryResponse { return c.response }
func (c *AggregateResults) HasMore() bool                  { return c.response.GetHasMore() }
func (c *AggregateResults) NextCursor() string             { return c.response.GetNextCursor() }

// Value returns the named value from the first loaded row. It returns nil for
// an empty collection or missing alias. SQL NULL retains its IsNull flag.
// Next appends rows, so it does not change which row Value reads.
func (c *AggregateResults) Value(alias string) *proto.Value {
	if c == nil || len(c.items) == 0 {
		return nil
	}
	return c.items[0].Value(alias)
}

// IntValue returns the first row's integer value, or zero if absent.
func (c *AggregateResults) IntValue(alias string) int64 { return c.Value(alias).GetInt() }

// FloatValue returns the first row's floating-point value, or zero if absent.
func (c *AggregateResults) FloatValue(alias string) float64 { return c.Value(alias).GetFloat() }

// BoolValue returns the first row's boolean value, or false if absent.
func (c *AggregateResults) BoolValue(alias string) bool { return c.Value(alias).GetBool() }

// TextValue returns the first row's text value, or an empty string if absent.
func (c *AggregateResults) TextValue(alias string) string { return c.Value(alias).GetText() }

// StringValue is an alias for TextValue.
func (c *AggregateResults) StringValue(alias string) string { return c.TextValue(alias) }

// All iterates over loaded rows without fetching more pages. Do not call Query
// or Next during iteration.
func (c *AggregateResults) All() iter.Seq[AggregateRow] {
	return func(yield func(AggregateRow) bool) {
		for _, row := range c.items {
			if !yield(row) {
				return
			}
		}
	}
}

// Query loads the first page, replacing existing rows and pagination state only
// after a successful fetch and validation. Supply GroupBy or an aggregate such
// as Count or Sum, or use ResultSet to reopen a saved aggregate query.
func (c *AggregateResults) Query(ctx context.Context, options ...QueryOption) error {
	if c.actor == nil || c.actor.connection == nil {
		return fmt.Errorf("aggregate collection: actor must have a connection")
	}
	request, err := c.actor.queryRequest(c.entityType, nil, options...)
	if err != nil {
		return err
	}
	if len(request.GroupBy) == 0 && len(request.Aggregates) == 0 && request.ResultSet == "" {
		return fmt.Errorf("aggregate collection: GroupBy or an aggregate is required")
	}
	response, rows, err := c.load(ctx, request)
	if err != nil {
		return err
	}
	c.request, c.response, c.items = request, response, rows
	return nil
}

// Next appends the next live or saved aggregate page. It returns io.EOF before
// Query or at exhaustion. Errors leave rows and cursor intact for retry.
func (c *AggregateResults) Next(ctx context.Context) error {
	request, err := nextQueryRequest(c.request, c.response)
	if err != nil {
		return err
	}
	response, rows, err := c.load(ctx, request)
	if err != nil {
		return err
	}
	c.items = append(c.items, rows...)
	c.request, c.response = request, response
	return nil
}

func (c *AggregateResults) load(ctx context.Context, request *proto.QueryRequest) (*proto.QueryResponse, []AggregateRow, error) {
	response, err := c.actor.connection.Query(ctx, gproto.Clone(request).(*proto.QueryRequest))
	if err != nil {
		return nil, nil, err
	}
	if response == nil {
		return nil, nil, fmt.Errorf("aggregate collection: nil query response")
	}
	if len(response.Entities) != 0 {
		return nil, nil, fmt.Errorf("aggregate collection: query returned entities instead of aggregate rows")
	}
	rows := make([]AggregateRow, 0, len(response.Rows))
	for _, row := range response.Rows {
		if row == nil {
			return nil, nil, fmt.Errorf("aggregate collection: nil aggregate row")
		}
		rows = append(rows, AggregateRow(row.Values))
	}
	return response, rows, nil
}
