package keystone

import (
	"context"
	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"net"
)

const bufSize = 1024 * 1024

var mockListener *bufconn.Listener

type MockServer struct {
	proto.UnimplementedKeystoneServer
	DefineFunc           func(context.Context, *proto.SchemaRequest) (*proto.Schema, error)
	MutateFunc           func(context.Context, *proto.MutateRequest) (*proto.MutateResponse, error)
	ReportTimeSeriesFunc func(context.Context, *proto.ReportTimeSeriesRequest) (*proto.MutateResponse, error)
	RetrieveFunc         func(context.Context, *proto.EntityRequest) (*proto.EntityResponse, error)
	FindFunc             func(context.Context, *proto.FindRequest) (*proto.FindResponse, error)
	ListFunc             func(context.Context, *proto.ListRequest) (*proto.ListResponse, error)
	GroupCountFunc       func(context.Context, *proto.GroupCountRequest) (*proto.GroupCountResponse, error)
	LogFunc              func(context.Context, *proto.LogRequest) (*proto.LogResponse, error)
	LogsFunc             func(context.Context, *proto.LogsRequest) (*proto.LogsResponse, error)
	EventsFunc           func(context.Context, *proto.EventRequest) (*proto.EventsResponse, error)
	DailyEntitiesFunc    func(context.Context, *proto.DailyEntityRequest) (*proto.DailyEntityResponse, error)
	SchemaStatisticsFunc func(context.Context, *proto.SchemaStatisticsRequest) (*proto.SchemaStatisticsResponse, error)
	ChartTimeSeriesFunc  func(context.Context, *proto.ChartTimeSeriesRequest) (*proto.ChartTimeSeriesResponse, error)
	ShareViewFunc        func(context.Context, *proto.ShareViewRequest) (*proto.SharedViewResponse, error)
	SharedViewsFunc      func(context.Context, *proto.SharedViewsRequest) (*proto.SharedViewsResponse, error)
	RateLimitFunc        func(context.Context, *proto.RateLimitRequest) (*proto.RateLimitResponse, error)
	StatusFunc           func(context.Context, *proto.Authorization) (*proto.StatusResponse, error)
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return mockListener.Dial()
}

// server.Serve(listener) / defer close

func MockConnection() (*Connection, *MockServer, *bufconn.Listener, *grpc.Server) {
	mockListener = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	m := &MockServer{}
	proto.RegisterKeystoneServer(s, m)
	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return NewConnection(proto.NewKeystoneClient(conn), "", "", ""), m, mockListener, s
}

func (m *MockServer) Define(ctx context.Context, req *proto.SchemaRequest) (*proto.Schema, error) {
	if m.DefineFunc == nil {
		return m.UnimplementedKeystoneServer.Define(ctx, req)
	}
	return m.DefineFunc(ctx, req)
}

func (m *MockServer) Mutate(ctx context.Context, req *proto.MutateRequest) (*proto.MutateResponse, error) {
	if m.MutateFunc == nil {
		return m.UnimplementedKeystoneServer.Mutate(ctx, req)
	}
	return m.MutateFunc(ctx, req)
}

func (m *MockServer) ReportTimeSeries(ctx context.Context, req *proto.ReportTimeSeriesRequest) (*proto.MutateResponse, error) {
	if m.ReportTimeSeriesFunc == nil {
		return m.UnimplementedKeystoneServer.ReportTimeSeries(ctx, req)
	}
	return m.ReportTimeSeriesFunc(ctx, req)
}

func (m *MockServer) Retrieve(ctx context.Context, req *proto.EntityRequest) (*proto.EntityResponse, error) {
	if m.RetrieveFunc == nil {
		return m.UnimplementedKeystoneServer.Retrieve(ctx, req)
	}
	return m.RetrieveFunc(ctx, req)
}

func (m *MockServer) Find(ctx context.Context, req *proto.FindRequest) (*proto.FindResponse, error) {
	if m.FindFunc == nil {
		return m.UnimplementedKeystoneServer.Find(ctx, req)
	}
	return m.FindFunc(ctx, req)
}

func (m *MockServer) List(ctx context.Context, req *proto.ListRequest) (*proto.ListResponse, error) {
	if m.ListFunc == nil {
		return m.UnimplementedKeystoneServer.List(ctx, req)
	}
	return m.ListFunc(ctx, req)
}

func (m *MockServer) GroupCount(ctx context.Context, req *proto.GroupCountRequest) (*proto.GroupCountResponse, error) {
	if m.GroupCountFunc == nil {
		return m.UnimplementedKeystoneServer.GroupCount(ctx, req)
	}
	return m.GroupCountFunc(ctx, req)
}

func (m *MockServer) Log(ctx context.Context, req *proto.LogRequest) (*proto.LogResponse, error) {
	if m.LogsFunc == nil {
		return m.UnimplementedKeystoneServer.Log(ctx, req)
	}
	return m.LogFunc(ctx, req)
}

func (m *MockServer) Logs(ctx context.Context, req *proto.LogsRequest) (*proto.LogsResponse, error) {
	if m.LogsFunc == nil {
		return m.UnimplementedKeystoneServer.Logs(ctx, req)
	}
	return m.LogsFunc(ctx, req)
}

func (m *MockServer) Events(ctx context.Context, req *proto.EventRequest) (*proto.EventsResponse, error) {
	if m.EventsFunc == nil {
		return m.UnimplementedKeystoneServer.Events(ctx, req)
	}
	return m.EventsFunc(ctx, req)
}
func (m *MockServer) DailyEntities(ctx context.Context, req *proto.DailyEntityRequest) (*proto.DailyEntityResponse, error) {
	if m.DailyEntitiesFunc == nil {
		return m.UnimplementedKeystoneServer.DailyEntities(ctx, req)
	}
	return m.DailyEntitiesFunc(ctx, req)
}
func (m *MockServer) SchemaStatistics(ctx context.Context, req *proto.SchemaStatisticsRequest) (*proto.SchemaStatisticsResponse, error) {
	if m.SchemaStatisticsFunc == nil {
		return m.UnimplementedKeystoneServer.SchemaStatistics(ctx, req)
	}
	return m.SchemaStatisticsFunc(ctx, req)
}

func (m *MockServer) ChartTimeSeries(ctx context.Context, req *proto.ChartTimeSeriesRequest) (*proto.ChartTimeSeriesResponse, error) {
	if m.ChartTimeSeriesFunc == nil {
		return m.UnimplementedKeystoneServer.ChartTimeSeries(ctx, req)
	}
	return m.ChartTimeSeriesFunc(ctx, req)
}

func (m *MockServer) ShareView(ctx context.Context, req *proto.ShareViewRequest) (*proto.SharedViewResponse, error) {
	if m.ShareViewFunc == nil {
		return m.UnimplementedKeystoneServer.ShareView(ctx, req)
	}
	return m.ShareViewFunc(ctx, req)
}

func (m *MockServer) ShareViews(ctx context.Context, req *proto.SharedViewsRequest) (*proto.SharedViewsResponse, error) {
	if m.SharedViewsFunc == nil {
		return m.UnimplementedKeystoneServer.SharedViews(ctx, req)
	}
	return m.SharedViewsFunc(ctx, req)
}

func (m *MockServer) RateLimit(ctx context.Context, req *proto.RateLimitRequest) (*proto.RateLimitResponse, error) {
	if m.SharedViewsFunc == nil {
		return m.UnimplementedKeystoneServer.RateLimit(ctx, req)
	}
	return m.RateLimitFunc(ctx, req)
}

func (m *MockServer) Status(ctx context.Context, req *proto.Authorization) (*proto.StatusResponse, error) {
	if m.StatusFunc == nil {
		return m.UnimplementedKeystoneServer.Status(ctx, req)
	}
	return m.StatusFunc(ctx, req)
}
