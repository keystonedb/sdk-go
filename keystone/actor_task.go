package keystone

import (
	"context"
	"errors"
	"io"

	"github.com/keystonedb/sdk-go/proto"
	"github.com/packaged/logger/v3/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func (a *Actor) TaskPush(ctx context.Context, taskName, taskID string, data map[string]string) error {
	resp, err := a.Connection().PushTask(ctx, &proto.PushTaskRequest{
		Authorization: a.Authorization(),
		TaskName:      taskName,
		TaskId:        taskID,
		Data:          data,
	})
	if err != nil {
		return err
	}
	if !resp.GetSuccess() {
		return errors.New("unable to push task")
	}
	return nil
}

func (a *Actor) TaskStream(ctx context.Context, taskName string, handler func(response *proto.TaskResponse) error) error {
	streamCtx := a.AuthorizeContext(ctx)
	streamCtx = metadata.AppendToOutgoingContext(streamCtx, "task_name", taskName)
	stream, err := a.Connection().TaskStream(streamCtx)
	if err != nil {
		logger.I().Error("Failed to open task stream", zap.Error(err))
		return err
	}

	defer func(stream grpc.ServerStreamingClient[proto.TaskResponse]) {
		_ = stream.CloseSend()
	}(stream)

	for {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.Canceled) {
				return nil
			}
		case <-stream.Context().Done():
			if errors.Is(stream.Context().Err(), context.Canceled) {
				return nil
			}
			return stream.Context().Err()
		default:
			tsk, recErr := stream.Recv()
			if recErr == io.EOF {
				return nil
			}
			if recErr != nil {
				return recErr
			}
			handleErr := handler(tsk)
			// Ack the message
			streamErr := stream.Send(&proto.TaskAckRequest{TaskId: tsk.GetTaskId(), Acked: handleErr == nil})
			if streamErr != nil {
				if streamErr == io.EOF {
					return nil
				}
				return streamErr
			}
		}
	}
}
