package apiv1

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	corev1 "github.com/ryfineZ/weave/gen/go/proxy/core/v1"
	"github.com/ryfineZ/weave/internal/log"
)

// ChainHandler implements corev1connect.ChainServiceHandler.
type ChainHandler struct {
	log *log.Logger
}

func NewChainHandler(l *log.Logger) *ChainHandler {
	return &ChainHandler{log: l}
}

func (h *ChainHandler) ListChains(_ context.Context, _ *connect.Request[corev1.ListChainsRequest]) (*connect.Response[corev1.ListChainsResponse], error) {
	return connect.NewResponse(&corev1.ListChainsResponse{}), nil
}

func (h *ChainHandler) GetChain(_ context.Context, _ *connect.Request[corev1.GetChainRequest]) (*connect.Response[corev1.GetChainResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ChainHandler) CreateChain(_ context.Context, _ *connect.Request[corev1.CreateChainRequest]) (*connect.Response[corev1.CreateChainResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ChainHandler) UpdateChain(_ context.Context, _ *connect.Request[corev1.UpdateChainRequest]) (*connect.Response[corev1.UpdateChainResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ChainHandler) DeleteChain(_ context.Context, _ *connect.Request[corev1.DeleteChainRequest]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ChainHandler) SetChainEnabled(_ context.Context, _ *connect.Request[corev1.SetChainEnabledRequest]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ChainHandler) WatchChainStates(_ context.Context, _ *connect.Request[corev1.WatchChainStatesRequest], _ *connect.ServerStream[corev1.ChainState]) error {
	return connect.NewError(connect.CodeUnimplemented, nil)
}
