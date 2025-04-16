package keystone

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/packaged/logger/v3/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"sync"
	"time"
)

type LogStream struct {
	actor  *Actor
	auth   *proto.Authorization
	client grpc.BidiStreamingClient[proto.LogRequest, proto.LogResponse]
	logs   chan *LogBatch
	wg     *sync.WaitGroup
	stop   bool
}

func NewLogStream(actor *Actor) (*LogStream, error) {
	client, err := actor.Connection().Log(context.Background())
	if err != nil {
		return nil, err
	}
	return &LogStream{
		actor: actor, client: client,
		auth: actor.Authorization(),
		wg:   new(sync.WaitGroup),
		logs: make(chan *LogBatch, 1000),
	}, nil
}

func (b *LogStream) Start() error {
	processed := 0
	defer func() {
		logger.I().Info("Processed", zap.Int("processed", processed))
	}()
	for {
		select {
		case batch, ok := <-b.logs:
			if !ok {
				return nil
			}
			processed++
			if err := b.client.Send(&proto.LogRequest{
				Authorization: b.auth,
				EntityId:      batch.entityID.String(),
				Logs:          batch.logs,
				BatchId:       batch.batchID,
			}); err != nil {
				logger.I().Error("failed to send batch log", zap.Error(err), zap.Int("size", len(b.logs)))
			} else {
				batch.Stored()
			}

			b.wg.Done()
		}
	}
}

func (b *LogStream) Stop() error {
	// Stop accepting new logs
	b.stop = true
	// wait for the queue to clear
	b.wg.Wait()
	b.client.Context().Done()
	close(b.logs)

	logger.I().Info("Closing Log", zap.Int("size", len(b.logs)))
	if err := b.client.CloseSend(); err != nil {
		return err
	}
	if err := b.client.Context().Err(); err != nil {
		return err
	}
	return nil
}

func (b *LogStream) LogBatch(log *LogBatch) error {
	if !b.stop {
		b.wg.Add(1)
		b.logs <- log
		return nil
	}
	return errors.New("log batch already stopped")
}

type LogBatch struct {
	entityID ID
	batchID  string
	logs     []*proto.EntityLog
}

func NewEntityLog(entityID ID) *LogBatch {
	return &LogBatch{
		entityID: entityID,
		batchID:  uuid.New().String(),
		logs:     []*proto.EntityLog{},
	}
}

func (e *LogBatch) Stored() {
	e.logs = nil
}

// Debug logs a debug message
func (e *LogBatch) Debug(message, reference, actor, traceID string, data map[string]string) {
	e.Log(proto.LogLevel_Debug, message, reference, actor, traceID, time.Now(), data)
}

// Info logs an info message
func (e *LogBatch) Info(message, reference, actor, traceID string, data map[string]string) {
	e.Log(proto.LogLevel_Info, message, reference, actor, traceID, time.Now(), data)
}

// Notice logs a notice message
func (e *LogBatch) Notice(message, reference, actor, traceID string, data map[string]string) {
	e.Log(proto.LogLevel_Notice, message, reference, actor, traceID, time.Now(), data)
}

// Warn logs a warning message
func (e *LogBatch) Warn(message, reference, actor, traceID string, data map[string]string) {
	e.Log(proto.LogLevel_Warn, message, reference, actor, traceID, time.Now(), data)
}

// Error logs an error message
func (e *LogBatch) Error(message, reference, actor, traceID string, data map[string]string) {
	e.Log(proto.LogLevel_Error, message, reference, actor, traceID, time.Now(), data)
}

// Critical logs a critical message
func (e *LogBatch) Critical(message, reference, actor, traceID string, data map[string]string) {
	e.Log(proto.LogLevel_Critical, message, reference, actor, traceID, time.Now(), data)
}

// Alert logs an alert message
func (e *LogBatch) Alert(message, reference, actor, traceID string, data map[string]string) {
	e.Log(proto.LogLevel_Alert, message, reference, actor, traceID, time.Now(), data)
}

// Fatal logs a fatal message
func (e *LogBatch) Fatal(message, reference, actor, traceID string, data map[string]string) {
	e.Log(proto.LogLevel_Fatal, message, reference, actor, traceID, time.Now(), data)
}

// Log logs a message
func (e *LogBatch) Log(level proto.LogLevel, message, reference, actor, traceID string, logTime time.Time, data map[string]string) {
	if e.logs == nil {
		e.logs = []*proto.EntityLog{}
	}

	e.logs = append(e.logs, &proto.EntityLog{
		Actor:     actor,
		Level:     level,
		Message:   message,
		Reference: reference,
		TraceId:   traceID,
		Time:      timestamppb.New(logTime),
		Data:      data,
	})
}
