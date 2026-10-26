package keystone

import (
	"context"
	"io"
	"testing"

	"github.com/keystonedb/sdk-go/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	gproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestAggregateOptionsAndRows(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	options := []QueryOption{GroupBy("category", "category"), Count("count"), CountProperty("price", "priced"), CountDistinct("name", "names"), Sum("price", "sum"), Min("price", "min"), Max("price", "max"), Avg("price", "avg"), Aggregate(proto.QueryAggregate_SUM, "price", "distinct_sum", true), Having(And(WhereGreaterThan("count", 1), Or(IsNull("category"), WhereEquals("category", "a")))), Explain()}
	var first *proto.QueryRequest
	mock.QueryFunc = func(_ context.Context, req *proto.QueryRequest) (*proto.QueryResponse, error) {
		require.Equal(t, []string{"category"}, req.GroupBy)
		require.Len(t, req.Aggregates, 8)
		require.True(t, req.Aggregates[2].Distinct)
		require.True(t, req.Aggregates[7].Distinct)
		require.True(t, req.Explain)
		require.Len(t, req.Having, 1)
		require.Len(t, req.Having[0].Nested, 2)
		require.Empty(t, req.Where)
		if first == nil {
			first = gproto.Clone(req).(*proto.QueryRequest)
		} else {
			require.True(t, gproto.Equal(first, req))
		}
		req.Having[0].Nested[0].Values[0].Int = 999
		return &proto.QueryResponse{Rows: []*proto.QueryRow{{Values: map[string]*proto.Value{"category": {Text: "a"}, "count": {Int: 2}, "avg": {Float: 1.5}}}}}, nil
	}
	for range 2 {
		response, err := actor.Query(context.Background(), "customer", options...)
		require.NoError(t, err)
		type row struct {
			Category string
			Count    int64
			Avg      float64
		}
		rows, err := QueryRows[row](response.Rows...)
		require.NoError(t, err)
		require.Equal(t, []row{{"a", 2, 1.5}}, rows)
		pointers, err := QueryRows[*row](response.Rows...)
		require.NoError(t, err)
		require.Equal(t, rows[0], *pointers[0])
	}
	for _, option := range []QueryOption{Having(Limit(1)), Having(nil), Having(And(WhereEquals("count", 1), Select("name"))), ResultSet("")} {
		_, err := actor.Query(context.Background(), "customer", option)
		require.Error(t, err)
	}
	_, err := QueryRows[int]()
	require.Error(t, err)
	_, err = QueryRows[struct{}](nil)
	require.Error(t, err)
}

func TestQueryPagerSavedResultsAndRetry(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	ctx := context.Background()
	expires := timestamppb.Now()
	calls := 0
	failure := status.Error(codes.Unavailable, "temporary")
	mock.QueryFunc = func(_ context.Context, req *proto.QueryRequest) (*proto.QueryResponse, error) {
		calls++
		if calls == 1 {
			require.True(t, req.SaveResultSet)
			require.Equal(t, "customer", req.From.Key)
			require.Len(t, req.Aggregates, 1)
			req.Limit = 777 // RPC implementations cannot mutate the stored request.
			return &proto.QueryResponse{Rows: []*proto.QueryRow{{Values: map[string]*proto.Value{"count": {Int: 1}}}}, HasMore: true, NextCursor: "next", ResultSet: "saved", ResultSetExpires: expires}, nil
		}
		want := &proto.QueryRequest{Authorization: actor.Authorization(), ResultSet: "saved", Cursor: "next", Limit: 1, IncludeTotal: true, Explain: true}
		require.True(t, gproto.Equal(want, req), "got %v", req)
		if calls == 2 {
			return nil, failure
		}
		return &proto.QueryResponse{ResultSet: "saved", ResultSetExpires: expires, Rows: []*proto.QueryRow{{Values: map[string]*proto.Value{"count": {Int: 2}}}}}, nil
	}
	pager, err := actor.QueryPages(ctx, "customer", Count("count"), SaveResultSet(), Limit(1), IncludeTotal(), Explain())
	require.NoError(t, err)
	original := pager.Response()
	require.Equal(t, codes.Unavailable, status.Code(pager.Next(ctx)))
	require.Same(t, original, pager.Response())
	require.NoError(t, pager.Next(ctx))
	require.Equal(t, int64(2), pager.Response().Rows[0].Values["count"].Int)
	require.ErrorIs(t, pager.Next(ctx), io.EOF)
	require.Equal(t, 3, calls)
}

func TestCollectionSavedResults(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	calls := 0
	mock.QueryFunc = func(_ context.Context, req *proto.QueryRequest) (*proto.QueryResponse, error) {
		calls++
		if calls == 1 {
			require.True(t, req.SaveResultSet)
			page := collectionPage("one", "Alice", "next")
			page.ResultSet = "saved"
			return page, nil
		}
		require.Nil(t, req.From)
		require.Empty(t, req.Select)
		require.Empty(t, req.Where)
		require.False(t, req.SaveResultSet)
		require.Equal(t, "saved", req.ResultSet)
		require.Equal(t, "next", req.Cursor)
		return collectionPage("two", "Bob", ""), nil
	}
	collection := Collection(collectionCustomer{}, actor)
	require.NoError(t, collection.Query(context.Background(), Select("name"), WhereEquals("paid", true), SaveResultSet()))
	require.NoError(t, collection.Next(context.Background()))
	require.Len(t, collection.Items(), 2)
	require.Error(t, collection.Query(context.Background(), Count("count")))
}

func TestResultSetReadsAndLivePager(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	ctx := context.Background()
	mock.QueryFunc = func(_ context.Context, req *proto.QueryRequest) (*proto.QueryResponse, error) {
		require.Nil(t, req.From)
		require.Equal(t, "saved", req.ResultSet)
		require.Equal(t, "cursor", req.Cursor)
		return &proto.QueryResponse{}, nil
	}
	_, err := actor.Query(ctx, "ignored", ResultSet("saved"), Cursor("cursor"))
	require.NoError(t, err)
	request := &proto.QueryRequest{ResultSet: "saved", Cursor: "cursor"}
	_, err = actor.QueryWithRequest(ctx, "", request)
	require.NoError(t, err)
	require.Nil(t, request.Authorization)
	calls := 0
	mock.QueryFunc = func(_ context.Context, req *proto.QueryRequest) (*proto.QueryResponse, error) {
		calls++
		require.Equal(t, "customer", req.From.Key)
		require.Equal(t, []string{"name"}, req.Select)
		if calls == 1 {
			return &proto.QueryResponse{HasMore: true, NextCursor: "live"}, nil
		}
		require.Equal(t, "live", req.Cursor)
		return &proto.QueryResponse{}, nil
	}
	pager, err := actor.QueryPages(ctx, "customer", Select("name"))
	require.NoError(t, err)
	require.NoError(t, pager.Next(ctx))
	require.ErrorIs(t, pager.Next(ctx), io.EOF)
	require.ErrorIs(t, (&QueryPager{}).Next(ctx), io.EOF)
	pager.response = &proto.QueryResponse{HasMore: true}
	require.Error(t, pager.Next(ctx))
	mock.QueryFunc = func(context.Context, *proto.QueryRequest) (*proto.QueryResponse, error) { return nil, nil }
	_, err = actor.QueryPages(ctx, "customer")
	require.NoError(t, err) // gRPC encodes a nil response as an empty message.
}

func TestQueryRowsNullAndZero(t *testing.T) {
	type row struct {
		Name     string
		Count    int64
		Optional *int64
		Missing  *int64
		Flag     bool
	}
	rows, err := QueryRows[row](&proto.QueryRow{Values: map[string]*proto.Value{
		"name": {Text: ""}, "count": {Int: 0}, "optional": {IsNull: true}, "missing": nil, "flag": {Bool: false},
	}})
	require.NoError(t, err)
	require.Equal(t, []row{{}}, rows)
	_, err = (&Actor{}).QueryPages(context.Background(), "customer")
	require.Error(t, err)
}
