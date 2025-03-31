package keystone

import (
	"context"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/packaged/logger/v3/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"reflect"
	"sync"
	"time"
)

// Connection is a connection to a keystone server
type Connection struct {
	proto.KeystoneClient
	client        proto.KeystoneClient
	logger        *logger.Logger
	timeLogConfig *logger.TimedLogConfig
	appID         proto.VendorApp
	token         string
	typeRegister  map[reflect.Type]*TypeDefinition
	registerQueue map[reflect.Type]bool // true if the type is processing registration
}

func DefaultConnection(host, port, vendorID, appID, accessToken string) *Connection {
	ksGrpcConn, err := grpc.Dial(host+":"+port, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithIdleTimeout(time.Minute*5), grpc.WithConnectParams(grpc.ConnectParams{MinConnectTimeout: time.Second * 5}))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	return NewConnection(proto.NewKeystoneClient(ksGrpcConn), vendorID, appID, accessToken)
}

// NewConnection creates a new connection to a keystone server
func NewConnection(client proto.KeystoneClient, vendorID, appID, accessToken string) *Connection {
	return &Connection{
		timeLogConfig: &logger.TimedLogConfig{
			ErrorDuration: time.Minute,
			WarnDuration:  30 * time.Second,
			InfoDuration:  2 * time.Second,
			DebugDuration: 100 * time.Millisecond,
		},
		logger:        logger.I(),
		client:        client,
		appID:         proto.VendorApp{VendorId: vendorID, AppId: appID},
		token:         accessToken,
		typeRegister:  make(map[reflect.Type]*TypeDefinition),
		registerQueue: make(map[reflect.Type]bool),
	}
}

// DirectClient avoid using the direct client in case of changes
func (c *Connection) DirectClient() proto.KeystoneClient { return c.client }

func (c *Connection) authorization() *proto.Authorization {
	return &proto.Authorization{
		Source: &c.appID,
		Token:  c.token,
	}
}

// Actor returns an actor for the given workspace, remote IP, user ID, and user agent
func (c *Connection) Actor(workspaceID, remoteIP, userID, userAgent string) Actor {
	return Actor{
		connection:  c,
		workspaceID: workspaceID,
		user: &proto.User{
			UserAgent: userAgent,
			RemoteIp:  remoteIP,
			UserId:    userID,
			Client:    "Keystone go-SDK",
		},
	}
}

// RegisterTypes registers the given types with the connection, returning the number of new types registered
func (c *Connection) RegisterTypes(types ...interface{}) int {
	registered := 0
	for _, t := range types {
		tt := reflect.TypeOf(t)
		alreadyRegistered := false
		if tt.Kind() == reflect.Ptr {
			_, alreadyRegistered = c.registerType(t)
		} else {
			vp := reflect.New(tt)
			vp.Elem().Set(reflect.ValueOf(t))
			_, alreadyRegistered = c.registerType(vp.Interface())
		}
		if !alreadyRegistered {
			registered++
		}
	}
	return registered
}

// registerType returns true if the type is already registered
func (c *Connection) registerType(t interface{}) (TypeDefinition, bool) {
	typ := reflect.TypeOf(t)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	sDef, ok := c.typeRegister[typ]
	if !ok {
		newDef := Define(t)
		c.typeRegister[typ] = &newDef
		c.registerQueue[typ] = false
		return newDef, false
	}
	return *sDef, true
}

// SyncSchema syncs the schema with the server
func (c *Connection) SyncSchema() *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	wg.Add(len(c.registerQueue))
	go func() {
		for typ, processing := range c.registerQueue {
			if !processing {
				if toRegister, ok := c.typeRegister[typ]; ok {
					resp, err := c.Define(context.Background(), &proto.SchemaRequest{
						Authorization: c.authorization(),
						Schema:        toRegister.Schema(),
						//TODO: Handle views
						//Views:         toRegister.definition.Views,
					})

					if err == nil {
						c.typeRegister[typ].id = resp.GetId()
						c.typeRegister[typ].Name = resp.GetName()
						c.typeRegister[typ].Type = resp.GetType()
						//TODO: Do we need this back/replaced?
						//c.typeRegister[typ].Properties = resp.GetProperties()
						c.typeRegister[typ].Options = resp.GetOptions()
						c.typeRegister[typ].Singular = resp.GetSingular()
						c.typeRegister[typ].Plural = resp.GetPlural()
					} else {
						c.logger.Error("Failed to define schema", zap.Error(err))
					}

				}
				wg.Done()
			}
		}
	}()
	return wg
}

func (c *Connection) Define(ctx context.Context, in *proto.SchemaRequest, opts ...grpc.CallOption) (*proto.Schema, error) {
	tl := c.timeLogConfig.NewLog("Define", zap.String("schema", in.GetSchema().GetType()))
	resp, err := c.client.Define(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return resp, err
}

func (c *Connection) Mutate(ctx context.Context, in *proto.MutateRequest, opts ...grpc.CallOption) (*proto.MutateResponse, error) {
	tl := c.timeLogConfig.NewLog("Mutate", zap.String("EntityId", in.GetEntityId()))
	resp, err := c.client.Mutate(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return resp, err
}

func (c *Connection) ReportTimeSeries(ctx context.Context, in *proto.ReportTimeSeriesRequest, opts ...grpc.CallOption) (*proto.MutateResponse, error) {
	tl := c.timeLogConfig.NewLog("ReportTimeSeries", zap.String("EntityId", in.GetEntityId()))
	resp, err := c.client.ReportTimeSeries(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return resp, err
}

func (c *Connection) Retrieve(ctx context.Context, in *proto.EntityRequest, opts ...grpc.CallOption) (*proto.EntityResponse, error) {
	tl := c.timeLogConfig.NewLog("Retrieve", zap.String("EntityId", in.GetEntityId()))
	resp, err := c.client.Retrieve(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return resp, err
}

func (c *Connection) Log(ctx context.Context, in *proto.LogRequest, opts ...grpc.CallOption) (*proto.LogResponse, error) {
	tl := c.timeLogConfig.NewLog("Logs", zap.String("EntityId", in.GetEntityId()))
	resp, err := c.client.Log(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return resp, err
}

func (c *Connection) Logs(ctx context.Context, in *proto.LogsRequest, opts ...grpc.CallOption) (*proto.LogsResponse, error) {
	tl := c.timeLogConfig.NewLog("Logs", zap.String("EntityId", in.GetEntityId()))
	resp, err := c.client.Logs(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return resp, err
}

func (c *Connection) Events(ctx context.Context, in *proto.EventRequest, opts ...grpc.CallOption) (*proto.EventsResponse, error) {
	tl := c.timeLogConfig.NewLog("Events", zap.String("EntityId", in.GetEntityId()))
	resp, err := c.client.Events(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return resp, err
}

func (c *Connection) Find(ctx context.Context, in *proto.FindRequest, opts ...grpc.CallOption) (*proto.FindResponse, error) {
	tl := c.timeLogConfig.NewLog("Find", zap.String("schema", in.GetSchema().GetKey()))
	resp, err := c.client.Find(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return resp, err
}

func (c *Connection) List(ctx context.Context, in *proto.ListRequest, opts ...grpc.CallOption) (*proto.ListResponse, error) {
	tl := c.timeLogConfig.NewLog("List", zap.String("schema", in.GetSchema().GetKey()))
	resp, err := c.client.List(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return resp, err
}

func (c *Connection) GroupCount(ctx context.Context, in *proto.GroupCountRequest, opts ...grpc.CallOption) (*proto.GroupCountResponse, error) {
	tl := c.timeLogConfig.NewLog("GroupCount", zap.String("schema", in.GetSchema().GetKey()))
	resp, err := c.client.GroupCount(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return resp, err
}

func (c *Connection) DailyEntities(ctx context.Context, in *proto.DailyEntityRequest, opts ...grpc.CallOption) (*proto.DailyEntityResponse, error) {
	tl := c.timeLogConfig.NewLog("DailyEntities", zap.String("schema", in.GetSchema().GetKey()))
	resp, err := c.client.DailyEntities(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return resp, err
}
func (c *Connection) SchemaStatistics(ctx context.Context, in *proto.SchemaStatisticsRequest, opts ...grpc.CallOption) (*proto.SchemaStatisticsResponse, error) {
	tl := c.timeLogConfig.NewLog("SchemaStatistics", zap.String("schema", in.GetSchema().GetKey()))
	resp, err := c.client.SchemaStatistics(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return resp, err
}

func (c *Connection) ChartTimeSeries(ctx context.Context, in *proto.ChartTimeSeriesRequest, opts ...grpc.CallOption) (*proto.ChartTimeSeriesResponse, error) {
	tl := c.timeLogConfig.NewLog("ChartTimeSeries", zap.String("schema", in.GetSchema().GetKey()))
	resp, err := c.client.ChartTimeSeries(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return resp, err
}

func (c *Connection) ShareView(ctx context.Context, in *proto.ShareViewRequest, opts ...grpc.CallOption) (*proto.SharedViewResponse, error) {
	tl := c.timeLogConfig.NewLog("ShareView", zap.String("EntityId", in.GetEntityId()))
	resp, err := c.client.ShareView(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return resp, err
}

func (c *Connection) SharedViews(ctx context.Context, in *proto.SharedViewsRequest, opts ...grpc.CallOption) (*proto.SharedViewsResponse, error) {
	tl := c.timeLogConfig.NewLog("ShareViews", zap.String("EntityId", in.GetEntityId()))
	resp, err := c.client.SharedViews(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return resp, err
}

func (c *Connection) RateLimit(ctx context.Context, in *proto.RateLimitRequest, opts ...grpc.CallOption) (*proto.RateLimitResponse, error) {
	tl := c.timeLogConfig.NewLog("RateLimit", zap.String("Key", in.GetKey()))
	resp, err := c.client.RateLimit(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return resp, err
}

func (c *Connection) AKVGet(ctx context.Context, in *proto.AKVGetRequest, opts ...grpc.CallOption) (*proto.AKVGetResponse, error) {
	tl := c.timeLogConfig.NewLog("AKVGet", zap.String("App", in.GetAuthorization().GetSource().String()))
	resp, err := c.client.AKVGet(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return resp, err
}
func (c *Connection) AKVPut(ctx context.Context, in *proto.AKVPutRequest, opts ...grpc.CallOption) (*proto.GenericResponse, error) {
	tl := c.timeLogConfig.NewLog("AKVPut", zap.String("App", in.GetAuthorization().GetSource().String()))
	resp, err := c.client.AKVPut(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return resp, err
}
func (c *Connection) AKVDel(ctx context.Context, in *proto.AKVDelRequest, opts ...grpc.CallOption) (*proto.GenericResponse, error) {
	tl := c.timeLogConfig.NewLog("AKVDel", zap.String("App", in.GetAuthorization().GetSource().String()))
	resp, err := c.client.AKVDel(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return resp, err
}

func (c *Connection) PiiToken(ctx context.Context, in *proto.PiiTokenRequest, opts ...grpc.CallOption) (*proto.PiiTokenResponse, error) {
	tl := c.timeLogConfig.NewLog("PiiToken", zap.String("App", in.GetAuthorization().GetSource().String()))
	resp, err := c.client.PiiToken(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return resp, err
}

func (c *Connection) PiiAnonymize(ctx context.Context, in *proto.PiiAnonymizeRequest, opts ...grpc.CallOption) (*proto.PiiAnonymizeResponse, error) {
	tl := c.timeLogConfig.NewLog("PiiAnonymize", zap.String("App", in.GetAuthorization().GetSource().String()))
	resp, err := c.client.PiiAnonymize(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return resp, err
}

func (c *Connection) IID(ctx context.Context, in *proto.IIDCreateRequest, opts ...grpc.CallOption) (*proto.IIDResponse, error) {
	tl := c.timeLogConfig.NewLog("IID", zap.String("App", in.GetAuthorization().GetSource().String()))
	resp, err := c.client.IID(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return resp, err
}

func (c *Connection) IIDLookup(ctx context.Context, in *proto.IIDRequest, opts ...grpc.CallOption) (*proto.IIDsResponse, error) {
	tl := c.timeLogConfig.NewLog("IIDLookup", zap.String("App", in.GetAuthorization().GetSource().String()))
	resp, err := c.client.IIDLookup(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return resp, err
}

func (c *Connection) EventStream(ctx context.Context, in *proto.EventStreamRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[proto.EventStreamResponse], error) {
	tl := c.timeLogConfig.NewLog("EventStream", zap.String("App", in.GetAuthorization().GetSource().String()))
	resp, err := c.client.EventStream(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return resp, err
}
