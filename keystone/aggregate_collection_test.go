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
)

func aggregatePage(category string, count int64, cursor, token string) *proto.QueryResponse {
	return &proto.QueryResponse{Rows: []*proto.QueryRow{{Values: map[string]*proto.Value{"category": {Text: category}, "count": {Int: count}, "sum": {IsNull: true}}}}, HasMore: cursor != "", NextCursor: cursor, ResultSet: token}
}

func TestAggregateCollectionPages(t *testing.T) {
	for _, saved := range []bool{false, true} {
		t.Run(map[bool]string{false: "live", true: "saved"}[saved], func(t *testing.T) {
			actor, mock, cleanup := newQueryIndexTestActor(t)
			defer cleanup()
			source := Collection((*collectionCustomer)(nil), actor)
			source.items = []*collectionCustomer{{Name: "existing"}}
			var original *proto.QueryRequest
			calls := 0
			token := ""
			if saved {
				token = "saved-token"
			}
			mock.QueryFunc = func(_ context.Context, req *proto.QueryRequest) (*proto.QueryResponse, error) {
				calls++
				if calls == 1 {
					require.Equal(t, "customer", req.From.Key)
					require.Equal(t, []string{"category"}, req.GroupBy)
					require.Len(t, req.Aggregates, 1)
					require.Len(t, req.Where, 1)
					require.Equal(t, saved, req.SaveResultSet)
					original = gproto.Clone(req).(*proto.QueryRequest)
					req.Limit = 999
					return aggregatePage("a", 2, "next", token), nil
				}
				want := gproto.Clone(original).(*proto.QueryRequest)
				want.Cursor = "next"
				if saved {
					want = &proto.QueryRequest{Authorization: actor.Authorization(), ResultSet: token, Cursor: "next", Limit: 1, IncludeTotal: true}
				}
				require.True(t, gproto.Equal(want, req), "unexpected next request: %v", req)
				return aggregatePage("b", 3, "", token), nil
			}
			options := []QueryOption{GroupBy("category"), Count("count"), WhereEquals("paid", true), Limit(1), IncludeTotal()}
			if saved {
				options = append(options, SaveResultSet())
			}
			rows, err := source.Aggregate(context.Background(), options...)
			require.NoError(t, err)
			require.True(t, rows.HasMore())
			require.Equal(t, "next", rows.NextCursor())
			require.Len(t, rows.Items(), 1)
			require.Equal(t, int64(2), rows.Items()[0].Value("count").GetInt())
			require.True(t, rows.Items()[0].Value("sum").IsNull)
			require.Nil(t, rows.Items()[0].Value("missing"))
			require.NoError(t, rows.Next(context.Background()))
			require.Len(t, rows.Items(), 2)
			require.ErrorIs(t, rows.Next(context.Background()), io.EOF)
			var categories []string
			for row := range rows.All() {
				categories = append(categories, row.Value("category").GetText())
			}
			require.Equal(t, []string{"a", "b"}, categories)
			yielded := 0
			for range rows.All() {
				yielded++
				break
			}
			require.Equal(t, 1, yielded)
			require.Equal(t, "existing", source.Items()[0].Name)
			require.Nil(t, source.Response())
		})
	}
}

func TestAggregateCollectionPreservesStateAndResets(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	ctx := context.Background()
	rows := AggregateCollection("customer", actor)
	require.Nil(t, rows.Response())
	require.ErrorIs(t, rows.Next(ctx), io.EOF)
	for range rows.All() {
		t.Fatal("empty collection yielded a row")
	}
	mock.QueryFunc = func(context.Context, *proto.QueryRequest) (*proto.QueryResponse, error) {
		return aggregatePage("a", 2, "next", "saved"), nil
	}
	require.NoError(t, rows.Query(ctx, GroupBy("category"), Count("count"), SaveResultSet()))
	original := rows.Response()
	mock.QueryFunc = func(context.Context, *proto.QueryRequest) (*proto.QueryResponse, error) {
		return nil, status.Error(codes.Unavailable, "retry")
	}
	require.Equal(t, codes.Unavailable, status.Code(rows.Next(ctx)))
	require.Equal(t, codes.Unavailable, status.Code(rows.Query(ctx, Count("count"))))
	require.Same(t, original, rows.Response())
	require.Equal(t, "next", rows.NextCursor())
	mock.QueryFunc = func(context.Context, *proto.QueryRequest) (*proto.QueryResponse, error) {
		return &proto.QueryResponse{Entities: []*proto.EntityResponse{{Entity: &proto.Entity{EntityId: "customer"}}}}, nil
	}
	require.Error(t, rows.Next(ctx))
	require.Same(t, original, rows.Response())
	mock.QueryFunc = func(_ context.Context, req *proto.QueryRequest) (*proto.QueryResponse, error) {
		require.Equal(t, "saved", req.ResultSet)
		require.Equal(t, "next", req.Cursor)
		return aggregatePage("b", 3, "", "saved"), nil
	}
	require.NoError(t, rows.Next(ctx))
	require.Len(t, rows.Items(), 2)
	mock.QueryFunc = func(_ context.Context, req *proto.QueryRequest) (*proto.QueryResponse, error) {
		require.Empty(t, req.ResultSet)
		require.Empty(t, req.Cursor)
		require.Empty(t, req.GroupBy)
		return &proto.QueryResponse{}, nil
	}
	require.NoError(t, rows.Query(ctx, Count("count")))
	require.Empty(t, rows.Items())
	require.False(t, rows.HasMore())
}

func TestAggregateCollectionValidation(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	ctx := context.Background()
	mock.QueryFunc = func(context.Context, *proto.QueryRequest) (*proto.QueryResponse, error) {
		t.Fatal("invalid collection query reached server")
		return nil, nil
	}
	require.Error(t, AggregateCollection("customer", nil).Query(ctx, Count("count")))
	require.Error(t, AggregateCollection("customer", &Actor{}).Query(ctx, Count("count")))
	require.Error(t, AggregateCollection("customer", actor).Query(ctx))
	require.Error(t, AggregateCollection("customer", actor).Query(ctx, Count("count"), nil))
	_, err := Collection(1, actor).Aggregate(ctx, Count("count"))
	require.Error(t, err)
	_, err = Collection[any](nil, actor).Aggregate(ctx, Count("count"))
	require.Error(t, err)
	_, err = Collection(collectionCustomer{}, nil).Aggregate(ctx, Count("count"))
	require.Error(t, err)
	mock.QueryFunc = func(_ context.Context, req *proto.QueryRequest) (*proto.QueryResponse, error) {
		require.Nil(t, req.From)
		require.Equal(t, "saved", req.ResultSet)
		return aggregatePage("a", 2, "", "saved"), nil
	}
	rows := AggregateCollection("customer", actor)
	require.NoError(t, rows.Query(ctx, ResultSet("saved")))
	require.Len(t, rows.Items(), 1)
}
