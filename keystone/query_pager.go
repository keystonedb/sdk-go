package keystone

import (
	"context"
	"fmt"
	"io"
	"reflect"

	"github.com/keystonedb/sdk-go/proto"
	gproto "google.golang.org/protobuf/proto"
)

// QueryPager retains the latest entity or aggregate page and its query controls.
// Use sequentially. Response belongs to the pager and must be treated as read-only.
// Next replaces the page, so memory use does not grow with the number of pages.
type QueryPager struct {
	actor    *Actor
	request  *proto.QueryRequest
	response *proto.QueryResponse
}

// QueryPages executes the first page of an entity or aggregate query.
func (a *Actor) QueryPages(ctx context.Context, entityType string, options ...QueryOption) (*QueryPager, error) {
	return a.QueryPagesWithRequest(ctx, entityType, nil, options...)
}

// QueryPagesWithRequest is QueryPages with advanced protobuf settings.
// Requests and options are snapshotted before the first RPC.
func (a *Actor) QueryPagesWithRequest(ctx context.Context, entityType string, request *proto.QueryRequest, options ...QueryOption) (*QueryPager, error) {
	if a == nil || a.connection == nil {
		return nil, fmt.Errorf("query: actor must have a connection")
	}
	req, err := a.queryRequest(entityType, request, options...)
	if err != nil {
		return nil, err
	}
	response, err := a.connection.Query(ctx, gproto.Clone(req).(*proto.QueryRequest))
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, fmt.Errorf("query: nil response")
	}
	return &QueryPager{actor: a, request: req, response: response}, nil
}

func (p *QueryPager) Response() *proto.QueryResponse { return p.response }
func (p *QueryPager) HasMore() bool                  { return p.response.GetHasMore() }
func (p *QueryPager) NextCursor() string             { return p.response.GetNextCursor() }

// Next fetches the next page, returning io.EOF when exhausted. Failures retain
// the previous page and cursor for retry. Saved queries continue by token without
// resending their filters or creating another result set.
func (p *QueryPager) Next(ctx context.Context) error {
	req, err := nextQueryRequest(p.request, p.response)
	if err != nil {
		return err
	}
	response, err := p.actor.connection.Query(ctx, gproto.Clone(req).(*proto.QueryRequest))
	if err != nil {
		return err
	}
	if response == nil {
		return fmt.Errorf("query: nil response")
	}
	p.request, p.response = req, response
	return nil
}

func nextQueryRequest(request *proto.QueryRequest, response *proto.QueryResponse) (*proto.QueryRequest, error) {
	if !response.GetHasMore() {
		return nil, io.EOF
	}
	if response.GetNextCursor() == "" {
		return nil, fmt.Errorf("query: response has more results but no next cursor")
	}
	if request == nil {
		return nil, fmt.Errorf("query: missing original request")
	}
	token := response.GetResultSet()
	if token == "" {
		token = request.ResultSet
	}
	if token != "" {
		return &proto.QueryRequest{Authorization: gproto.Clone(request.Authorization).(*proto.Authorization), ResultSet: token, Cursor: response.NextCursor, Limit: request.Limit, IncludeTotal: request.IncludeTotal, Explain: request.Explain}, nil
	}
	if request.SaveResultSet {
		return nil, fmt.Errorf("query: saved response has no result set token")
	}
	req := gproto.Clone(request).(*proto.QueryRequest)
	req.Cursor = response.NextCursor
	return req, nil
}

// QueryRows decodes aggregate rows into structs or pointers to structs using
// Keystone property names and tags. Raw Values remain available on QueryResponse
// to distinguish SQL NULL from zero values or read dynamically named aliases.
func QueryRows[T any](rows ...*proto.QueryRow) ([]T, error) {
	t := reflect.TypeFor[T]()
	pointer := t.Kind() == reflect.Pointer
	if pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("query: row type must be a struct or pointer to a struct")
	}
	result := make([]T, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			return nil, fmt.Errorf("query: nil aggregate row")
		}
		data := make(map[Property]*proto.Value, len(row.Values))
		for name, value := range row.Values {
			if value == nil || value.IsNull {
				continue
			}
			data[NewProperty(name)] = value
		}
		value := reflect.New(t)
		if err := UnmarshalProperties(data, value.Interface()); err != nil {
			return nil, fmt.Errorf("query: decode row: %w", err)
		}
		if pointer {
			result = append(result, value.Interface().(T))
		} else {
			result = append(result, value.Elem().Interface().(T))
		}
	}
	return result, nil
}
