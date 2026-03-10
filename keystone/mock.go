package keystone

import (
	"context"
	"net"

	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var mockListener *bufconn.Listener

type MockServer struct {
	proto.UnimplementedKeystoneServer
	DefineFunc           func(context.Context, *proto.SchemaRequest) (*proto.Schema, error)
	MutateFunc           func(context.Context, *proto.MutateRequest) (*proto.MutateResponse, error)
	DestroyFunc          func(context.Context, *proto.DestroyRequest) (*proto.DestroyResponse, error)
	ReportTimeSeriesFunc func(context.Context, *proto.ReportTimeSeriesRequest) (*proto.MutateResponse, error)
	SnapshotReportFunc   func(context.Context, *proto.SnapshotReportRequest) (*proto.MutateResponse, error)
	RetrieveFunc         func(context.Context, *proto.EntityRequest) (*proto.EntityResponse, error)
	FindFunc             func(context.Context, *proto.FindRequest) (*proto.FindResponse, error)
	ListFunc             func(context.Context, *proto.ListRequest) (*proto.ListResponse, error)
	QueryIndexFunc       func(context.Context, *proto.QueryIndexRequest) (*proto.QueryIndexResponse, error)
	LookupFunc           func(context.Context, *proto.LookupRequest) (*proto.LookupResponse, error)
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
	IIDFunc              func(context.Context, *proto.IIDCreateRequest) (*proto.IIDResponse, error)
	IIDLookupFunc        func(context.Context, *proto.IIDRequest) (*proto.IIDsResponse, error)
	PiiTokenFunc         func(context.Context, *proto.PiiTokenRequest) (*proto.PiiTokenResponse, error)
	PiiAnonymizeFunc     func(context.Context, *proto.PiiAnonymizeRequest) (*proto.PiiAnonymizeResponse, error)
	SquidFunc            func(context.Context, *proto.SquidRequest) (*proto.SquidResponse, error)
	SquidRecoverFunc     func(context.Context, *proto.SquidRecoverRequest) (*proto.SquidResponse, error)
	AKVGetFunc           func(context.Context, *proto.AKVGetRequest) (*proto.AKVGetResponse, error)
	AKVPutFunc           func(context.Context, *proto.AKVPutRequest) (*proto.GenericResponse, error)
	AKVDelFunc           func(context.Context, *proto.AKVDelRequest) (*proto.GenericResponse, error)
	PushTaskFunc         func(context.Context, *proto.PushTaskRequest) (*proto.GenericResponse, error)
	EnumPutFunc          func(context.Context, *proto.EnumPutRequest) (*proto.GenericResponse, error)
	EnumGetFunc          func(context.Context, *proto.EnumGetRequest) (*proto.EnumGetResponse, error)
	EnumDeleteFunc       func(context.Context, *proto.EnumDeleteRequest) (*proto.GenericResponse, error)
	EnumListFunc         func(context.Context, *proto.EnumListRequest) (*proto.EnumListResponse, error)
	EnumReplaceFunc      func(context.Context, *proto.EnumReplaceRequest) (*proto.GenericResponse, error)
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
	if m.LogFunc == nil {
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
	if m.RateLimitFunc == nil {
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

func (m *MockServer) Destroy(ctx context.Context, req *proto.DestroyRequest) (*proto.DestroyResponse, error) {
	if m.DestroyFunc == nil {
		return m.UnimplementedKeystoneServer.Destroy(ctx, req)
	}
	return m.DestroyFunc(ctx, req)
}

func (m *MockServer) SnapshotReport(ctx context.Context, req *proto.SnapshotReportRequest) (*proto.MutateResponse, error) {
	if m.SnapshotReportFunc == nil {
		return m.UnimplementedKeystoneServer.SnapshotReport(ctx, req)
	}
	return m.SnapshotReportFunc(ctx, req)
}

func (m *MockServer) QueryIndex(ctx context.Context, req *proto.QueryIndexRequest) (*proto.QueryIndexResponse, error) {
	if m.QueryIndexFunc == nil {
		return m.UnimplementedKeystoneServer.QueryIndex(ctx, req)
	}
	return m.QueryIndexFunc(ctx, req)
}

func (m *MockServer) Lookup(ctx context.Context, req *proto.LookupRequest) (*proto.LookupResponse, error) {
	if m.LookupFunc == nil {
		return m.UnimplementedKeystoneServer.Lookup(ctx, req)
	}
	return m.LookupFunc(ctx, req)
}

func (m *MockServer) IID(ctx context.Context, req *proto.IIDCreateRequest) (*proto.IIDResponse, error) {
	if m.IIDFunc == nil {
		return m.UnimplementedKeystoneServer.IID(ctx, req)
	}
	return m.IIDFunc(ctx, req)
}

func (m *MockServer) IIDLookup(ctx context.Context, req *proto.IIDRequest) (*proto.IIDsResponse, error) {
	if m.IIDLookupFunc == nil {
		return m.UnimplementedKeystoneServer.IIDLookup(ctx, req)
	}
	return m.IIDLookupFunc(ctx, req)
}

func (m *MockServer) PiiToken(ctx context.Context, req *proto.PiiTokenRequest) (*proto.PiiTokenResponse, error) {
	if m.PiiTokenFunc == nil {
		return m.UnimplementedKeystoneServer.PiiToken(ctx, req)
	}
	return m.PiiTokenFunc(ctx, req)
}

func (m *MockServer) PiiAnonymize(ctx context.Context, req *proto.PiiAnonymizeRequest) (*proto.PiiAnonymizeResponse, error) {
	if m.PiiAnonymizeFunc == nil {
		return m.UnimplementedKeystoneServer.PiiAnonymize(ctx, req)
	}
	return m.PiiAnonymizeFunc(ctx, req)
}

func (m *MockServer) SQUID(ctx context.Context, req *proto.SquidRequest) (*proto.SquidResponse, error) {
	if m.SquidFunc == nil {
		return m.UnimplementedKeystoneServer.SQUID(ctx, req)
	}
	return m.SquidFunc(ctx, req)
}

func (m *MockServer) SQUIDRecover(ctx context.Context, req *proto.SquidRecoverRequest) (*proto.SquidResponse, error) {
	if m.SquidRecoverFunc == nil {
		return m.UnimplementedKeystoneServer.SQUIDRecover(ctx, req)
	}
	return m.SquidRecoverFunc(ctx, req)
}

func (m *MockServer) AKVGet(ctx context.Context, req *proto.AKVGetRequest) (*proto.AKVGetResponse, error) {
	if m.AKVGetFunc == nil {
		return m.UnimplementedKeystoneServer.AKVGet(ctx, req)
	}
	return m.AKVGetFunc(ctx, req)
}

func (m *MockServer) AKVPut(ctx context.Context, req *proto.AKVPutRequest) (*proto.GenericResponse, error) {
	if m.AKVPutFunc == nil {
		return m.UnimplementedKeystoneServer.AKVPut(ctx, req)
	}
	return m.AKVPutFunc(ctx, req)
}

func (m *MockServer) AKVDel(ctx context.Context, req *proto.AKVDelRequest) (*proto.GenericResponse, error) {
	if m.AKVDelFunc == nil {
		return m.UnimplementedKeystoneServer.AKVDel(ctx, req)
	}
	return m.AKVDelFunc(ctx, req)
}

func (m *MockServer) PushTask(ctx context.Context, req *proto.PushTaskRequest) (*proto.GenericResponse, error) {
	if m.PushTaskFunc == nil {
		return m.UnimplementedKeystoneServer.PushTask(ctx, req)
	}
	return m.PushTaskFunc(ctx, req)
}

func (m *MockServer) EnumPut(ctx context.Context, req *proto.EnumPutRequest) (*proto.GenericResponse, error) {
	if m.EnumPutFunc == nil {
		return m.UnimplementedKeystoneServer.EnumPut(ctx, req)
	}
	return m.EnumPutFunc(ctx, req)
}

func (m *MockServer) EnumGet(ctx context.Context, req *proto.EnumGetRequest) (*proto.EnumGetResponse, error) {
	if m.EnumGetFunc == nil {
		return m.UnimplementedKeystoneServer.EnumGet(ctx, req)
	}
	return m.EnumGetFunc(ctx, req)
}

func (m *MockServer) EnumDelete(ctx context.Context, req *proto.EnumDeleteRequest) (*proto.GenericResponse, error) {
	if m.EnumDeleteFunc == nil {
		return m.UnimplementedKeystoneServer.EnumDelete(ctx, req)
	}
	return m.EnumDeleteFunc(ctx, req)
}

func (m *MockServer) EnumList(ctx context.Context, req *proto.EnumListRequest) (*proto.EnumListResponse, error) {
	if m.EnumListFunc == nil {
		return m.UnimplementedKeystoneServer.EnumList(ctx, req)
	}
	return m.EnumListFunc(ctx, req)
}

func (m *MockServer) EnumReplace(ctx context.Context, req *proto.EnumReplaceRequest) (*proto.GenericResponse, error) {
	if m.EnumReplaceFunc == nil {
		return m.UnimplementedKeystoneServer.EnumReplace(ctx, req)
	}
	return m.EnumReplaceFunc(ctx, req)
}
