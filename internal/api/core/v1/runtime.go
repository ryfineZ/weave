package apiv1

import (
	"context"
	"runtime"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	corev1 "github.com/ryfineZ/weave/gen/go/proxy/core/v1"
	"github.com/ryfineZ/weave/internal/log"
)

const daemonVersion = "0.1.0"

// RuntimeHandler implements corev1connect.RuntimeServiceHandler.
type RuntimeHandler struct {
	log       *log.Logger
	startedAt time.Time
}

func NewRuntimeHandler(l *log.Logger) *RuntimeHandler {
	return &RuntimeHandler{log: l, startedAt: time.Now()}
}

func (h *RuntimeHandler) GetVersion(
	_ context.Context,
	_ *connect.Request[emptypb.Empty],
) (*connect.Response[corev1.VersionInfo], error) {
	return connect.NewResponse(&corev1.VersionInfo{
		Version:   daemonVersion,
		GoVersion: runtime.Version(),
		// Commit and BuildDate injected at link time via -ldflags in production.
		// sing-box version available once the engine is initialised.
	}), nil
}

func (h *RuntimeHandler) GetStatus(
	_ context.Context,
	_ *connect.Request[emptypb.Empty],
) (*connect.Response[corev1.DaemonStatus], error) {
	return connect.NewResponse(&corev1.DaemonStatus{
		EngineState: corev1.EngineState_ENGINE_STATE_STOPPED,
		StartedAt:   timestamppb.New(h.startedAt),
	}), nil
}

func (h *RuntimeHandler) Start(
	_ context.Context,
	_ *connect.Request[emptypb.Empty],
) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *RuntimeHandler) Stop(
	_ context.Context,
	_ *connect.Request[emptypb.Empty],
) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *RuntimeHandler) WatchTraffic(
	_ context.Context,
	_ *connect.Request[emptypb.Empty],
	stream *connect.ServerStream[corev1.TrafficSnapshot],
) error {
	return connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *RuntimeHandler) WatchLogs(
	ctx context.Context,
	req *connect.Request[corev1.WatchLogsRequest],
	stream *connect.ServerStream[corev1.LogEntry],
) error {
	minLevel := log.Level(req.Msg.GetMinLevel())
	if minLevel == 0 {
		minLevel = log.LevelInfo
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case entry, ok := <-h.log.Entries():
			if !ok {
				return nil
			}
			if entry.Level < minLevel {
				continue
			}
			if err := stream.Send(&corev1.LogEntry{
				Level:   corev1.LogLevel(entry.Level),
				Message: entry.Message,
				Ts:      timestamppb.Now(),
			}); err != nil {
				return err
			}
		}
	}
}
