package keystone

import (
	"time"

	"github.com/keystonedb/sdk-go/proto"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// LogProvider is an interface for entities that can have logs
type LogProvider interface {
	ClearLogs() error
	GetLogs() []*proto.EntityLog
}

// EmbeddedLogs is a struct that implements LogProvider
type EmbeddedLogs struct {
	ksEntityLogs []*proto.EntityLog
}

// ClearLogs clears the logs
func (e *EmbeddedLogs) ClearLogs() error {
	e.ksEntityLogs = []*proto.EntityLog{}
	return nil
}

// GetLogs returns the logs
func (e *EmbeddedLogs) GetLogs() []*proto.EntityLog {
	return e.ksEntityLogs
}

// LogDebug logs a debug message
func (e *EmbeddedLogs) LogDebug(message, reference, actor, traceID string, data map[string]string) {
	e.Log(proto.LogLevel_Debug, message, reference, actor, traceID, time.Now(), data)
}

// LogInfo logs an info message
func (e *EmbeddedLogs) LogInfo(message, reference, actor, traceID string, data map[string]string) {
	e.Log(proto.LogLevel_Info, message, reference, actor, traceID, time.Now(), data)
}

// LogNotice logs a notice message
func (e *EmbeddedLogs) LogNotice(message, reference, actor, traceID string, data map[string]string) {
	e.Log(proto.LogLevel_Notice, message, reference, actor, traceID, time.Now(), data)
}

// LogWarn logs a warning message
func (e *EmbeddedLogs) LogWarn(message, reference, actor, traceID string, data map[string]string) {
	e.Log(proto.LogLevel_Warn, message, reference, actor, traceID, time.Now(), data)
}

// LogError logs an error message
func (e *EmbeddedLogs) LogError(message, reference, actor, traceID string, data map[string]string) {
	e.Log(proto.LogLevel_Error, message, reference, actor, traceID, time.Now(), data)
}

// LogCritical logs a critical message
func (e *EmbeddedLogs) LogCritical(message, reference, actor, traceID string, data map[string]string) {
	e.Log(proto.LogLevel_Critical, message, reference, actor, traceID, time.Now(), data)
}

// LogAlert logs an alert message
func (e *EmbeddedLogs) LogAlert(message, reference, actor, traceID string, data map[string]string) {
	e.Log(proto.LogLevel_Alert, message, reference, actor, traceID, time.Now(), data)
}

// LogFatal logs a fatal message
func (e *EmbeddedLogs) LogFatal(message, reference, actor, traceID string, data map[string]string) {
	e.Log(proto.LogLevel_Fatal, message, reference, actor, traceID, time.Now(), data)
}

// Log logs a message
func (e *EmbeddedLogs) Log(level proto.LogLevel, message, reference, actor, traceID string, logTime time.Time, data map[string]string) {
	if e.ksEntityLogs == nil {
		e.ksEntityLogs = make([]*proto.EntityLog, 0)
	}
	e.ksEntityLogs = append(e.ksEntityLogs, &proto.EntityLog{
		Actor:     actor,
		Level:     level,
		Message:   message,
		Reference: reference,
		TraceId:   traceID,
		Time:      timestamppb.New(logTime),
		Data:      data,
	})
}
