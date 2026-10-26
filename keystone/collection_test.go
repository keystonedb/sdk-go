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

type collectionCustomer struct {
	BaseEntity
	Name string
	Paid bool
}

func (*collectionCustomer) GetKeystoneDefinition() TypeDefinition {
	return TypeDefinition{Type: "customer"}
}

func collectionPage(id, name, cursor string) *proto.QueryResponse {
	return &proto.QueryResponse{
		Entities:   []*proto.EntityResponse{{Entity: &proto.Entity{EntityId: id}, Exists: Pointer(true), Properties: []*proto.EntityProperty{{Property: "name", Value: &proto.Value{Text: name}}}}},
		NextCursor: cursor, HasMore: cursor != "", TotalResults: Pointer(int64(2)), Execution: &proto.QueryExecution{Strategy: "test"},
	}
}

func TestCollectionQueryAndPagination(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	ctx := context.Background()
	customers := Collection(collectionCustomer{Name: "ignored"}, actor)
	require.Nil(t, customers.Response())
	require.ErrorIs(t, customers.Next(ctx), io.EOF)
	var firstRequest *proto.QueryRequest
	calls := 0
	mock.QueryFunc = func(_ context.Context, req *proto.QueryRequest) (*proto.QueryResponse, error) {
		calls++
		require.Equal(t, "customer", req.From.Key)
		require.True(t, gproto.Equal(actor.Authorization(), req.Authorization))
		if calls == 1 {
			firstRequest = gproto.Clone(req).(*proto.QueryRequest)
			return collectionPage("customer-1", "Alice", "page-2"), nil
		}
		want := gproto.Clone(firstRequest).(*proto.QueryRequest)
		want.Cursor = "page-2"
		require.True(t, gproto.Equal(want, req), "pagination changed query: %v", req)
		return collectionPage("customer-2", "Bob", ""), nil
	}
	require.NoError(t, customers.Query(ctx, Select("name"), WhereEquals("name", "Alice"), SortAsc("name"), Limit(1), IncludeTotal()))
	require.Len(t, customers.Items(), 1)
	require.Equal(t, "Alice", customers.Items()[0].Name)
	require.Equal(t, ID("customer-1"), customers.Items()[0].GetKeystoneID())
	require.True(t, customers.Items()[0].ExistsInKeystone())
	require.True(t, customers.Items()[0].HasWatcher())
	require.Equal(t, int64(2), customers.Response().GetTotalResults())
	require.Equal(t, "test", customers.Response().Execution.Strategy)
	require.True(t, customers.HasMore())
	require.NoError(t, customers.Next(ctx))
	require.Len(t, customers.Items(), 2)
	require.Equal(t, "Bob", customers.Items()[1].Name)
	require.False(t, customers.HasMore())
	require.ErrorIs(t, customers.Next(ctx), io.EOF)
	require.Equal(t, 2, calls)
}

func TestCollectionPreservesStateOnFailureAndCanReset(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	ctx := context.Background()
	customers := Collection(collectionCustomer{}, actor)
	mock.QueryFunc = func(context.Context, *proto.QueryRequest) (*proto.QueryResponse, error) {
		return collectionPage("customer-1", "Alice", "page-2"), nil
	}
	require.NoError(t, customers.Query(ctx, Select("name")))
	original := customers.Response()
	transportErr := status.Error(codes.Unavailable, "unavailable")
	mock.QueryFunc = func(context.Context, *proto.QueryRequest) (*proto.QueryResponse, error) { return nil, transportErr }
	require.Equal(t, codes.Unavailable, status.Code(customers.Next(ctx)))
	require.Equal(t, codes.Unavailable, status.Code(customers.Query(ctx)))
	require.Same(t, original, customers.Response())
	for _, invalid := range []*proto.EntityResponse{nil, {}, {Entity: &proto.Entity{}}} {
		mock.QueryFunc = func(context.Context, *proto.QueryRequest) (*proto.QueryResponse, error) {
			return &proto.QueryResponse{Entities: []*proto.EntityResponse{invalid}}, nil
		}
		require.ErrorIs(t, customers.Next(ctx), ErrEntityIDMissing)
		require.Same(t, original, customers.Response())
		require.Len(t, customers.Items(), 1)
	}
	mock.QueryFunc = func(_ context.Context, req *proto.QueryRequest) (*proto.QueryResponse, error) {
		require.Equal(t, "page-2", req.Cursor)
		return collectionPage("customer-2", "Bob", ""), nil
	}
	require.NoError(t, customers.Next(ctx))
	require.Len(t, customers.Items(), 2)
	mock.QueryFunc = func(_ context.Context, req *proto.QueryRequest) (*proto.QueryResponse, error) {
		require.Empty(t, req.Select)
		require.Empty(t, req.Cursor)
		require.Empty(t, req.Where)
		return &proto.QueryResponse{}, nil
	}
	require.NoError(t, customers.Query(ctx))
	require.Empty(t, customers.Items())
	require.False(t, customers.HasMore())
}

func TestCollectionPointerAndValidation(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	ctx := context.Background()
	mock.QueryFunc = func(context.Context, *proto.QueryRequest) (*proto.QueryResponse, error) {
		return collectionPage("customer-1", "Alice", ""), nil
	}
	customers := Collection((*collectionCustomer)(nil), actor)
	require.NoError(t, customers.Query(ctx, Select("name")))
	require.IsType(t, &collectionCustomer{}, customers.Items()[0])
	require.Equal(t, "Alice", customers.Items()[0].Name)
	require.Error(t, Collection(1, actor).Query(ctx))
	require.Error(t, Collection[any](nil, actor).Query(ctx))
	require.Error(t, Collection(collectionCustomer{}, nil).Query(ctx))
	require.Error(t, Collection(collectionCustomer{}, &Actor{}).Query(ctx))
	require.Error(t, customers.Query(ctx, Limit(1, 1)))
	require.Error(t, customers.Query(ctx, nil))
}

func TestCollectionSnapshotsQueryOptions(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	customers := Collection(collectionCustomer{}, actor)
	ids := []string{"customer-1", "customer-2"}
	calls := 0
	mock.QueryFunc = func(_ context.Context, req *proto.QueryRequest) (*proto.QueryResponse, error) {
		calls++
		require.Equal(t, []string{"customer-1", "customer-2"}, req.EntityIds.Values)
		if calls == 1 {
			return collectionPage("customer-1", "Alice", "next"), nil
		}
		return collectionPage("customer-2", "Bob", ""), nil
	}
	require.NoError(t, customers.Query(context.Background(), WithEntityIDs(ids)))
	ids[0] = "changed"
	require.NoError(t, customers.Next(context.Background()))
}

func TestCollectionRange(t *testing.T) {
	customers := Collection(collectionCustomer{}, nil)
	for range customers.All() {
		t.Fatal("empty collection yielded an item")
	}
	customers.items = []collectionCustomer{{Name: "Alice"}, {Name: "Bob"}, {Name: "Charlie"}}
	var names []string
	for customer := range customers.All() {
		names = append(names, customer.Name)
	}
	require.Equal(t, []string{"Alice", "Bob", "Charlie"}, names)
	names = nil
	for customer := range customers.All() {
		names = append(names, customer.Name)
		break
	}
	require.Equal(t, []string{"Alice"}, names)
}

func TestCollectionLatest(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	customers := Collection(collectionCustomer{}, actor)
	customers.items = []collectionCustomer{{Name: "previous"}}
	calls := 0
	mock.QueryFunc = func(_ context.Context, req *proto.QueryRequest) (*proto.QueryResponse, error) {
		calls++
		require.Equal(t, "customer", req.From.Key)
		require.True(t, gproto.Equal(actor.Authorization(), req.Authorization))
		require.Equal(t, int32(2), req.Limit)
		require.Equal(t, []string{"name"}, req.Select)
		require.Len(t, req.Where, 1)
		require.Equal(t, "name", req.Where[0].Property)
		require.Equal(t, []*proto.PropertySort{{Property: "_created", Descending: true}, {Property: "name"}}, req.OrderBy)
		if calls == 1 {
			require.Empty(t, req.Cursor)
			return collectionPage("customer-new", "newest", "older"), nil
		}
		require.Equal(t, "older", req.Cursor)
		return collectionPage("customer-old", "older", ""), nil
	}
	ctx := context.Background()
	for _, size := range []int32{0, -1} {
		require.Error(t, customers.Latest(ctx, size))
		require.Equal(t, "previous", customers.Items()[0].Name)
	}
	require.Zero(t, calls)
	require.NoError(t, customers.Latest(ctx, 2, Select("name"), WhereEquals("name", "example"), SortAsc("name"), Limit(99), Cursor("ignored")))
	require.Len(t, customers.Items(), 1)
	require.Equal(t, "newest", customers.Items()[0].Name)
	require.NoError(t, customers.Next(ctx))
	require.Len(t, customers.Items(), 2)
	require.Equal(t, "older", customers.Items()[1].Name)
	require.Equal(t, 2, calls)
}

type CustomerRepository struct {
	*EntityCollection[collectionCustomer]
}

func NewCustomerRepository(actor *Actor) *CustomerRepository {
	return &CustomerRepository{
		EntityCollection: Collection(collectionCustomer{}, actor),
	}
}

func (r *CustomerRepository) Paid(
	ctx context.Context,
	options ...QueryOption,
) error {
	opts := []QueryOption{
		WhereEquals("paid", true),
	}
	return r.Query(ctx, append(opts, options...)...)
}

func TestCustomerRepository(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	customers := NewCustomerRepository(actor)
	ctx := context.Background()
	mock.QueryFunc = func(_ context.Context, req *proto.QueryRequest) (*proto.QueryResponse, error) {
		require.Equal(t, "customer", req.From.Key)
		require.True(t, gproto.Equal(actor.Authorization(), req.Authorization))
		require.Equal(t, []string{"name", "paid"}, req.Select)
		require.Len(t, req.Where, 1)
		require.Equal(t, "paid", req.Where[0].Property)
		require.Equal(t, proto.Operator_Equal, req.Where[0].Operator)
		require.Len(t, req.Where[0].Values, 1)
		require.True(t, req.Where[0].Values[0].GetBool())
		response := collectionPage("customer-1", "Alice", "")
		response.Entities[0].Properties = append(response.Entities[0].Properties,
			&proto.EntityProperty{Property: "paid", Value: &proto.Value{Bool: true}})
		return response, nil
	}
	require.NoError(t, customers.Paid(ctx, Select("name", "paid")))

	var names []string
	for customer := range customers.All() {
		names = append(names, customer.Name)
		require.True(t, customer.Paid)
		require.Equal(t, ID("customer-1"), customer.GetKeystoneID())
	}
	require.Equal(t, []string{"Alice"}, names)
}
