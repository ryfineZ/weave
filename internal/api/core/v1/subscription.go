package apiv1

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	corev1 "github.com/ryfineZ/weave/gen/go/proxy/core/v1"
	"github.com/ryfineZ/weave/internal/log"
)

// SubscriptionHandler implements corev1connect.SubscriptionServiceHandler.
type SubscriptionHandler struct {
	log *log.Logger
}

func NewSubscriptionHandler(l *log.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{log: l}
}

func (h *SubscriptionHandler) ListSubscriptions(_ context.Context, _ *connect.Request[emptypb.Empty]) (*connect.Response[corev1.ListSubscriptionsResponse], error) {
	return connect.NewResponse(&corev1.ListSubscriptionsResponse{}), nil
}

func (h *SubscriptionHandler) AddSubscription(_ context.Context, _ *connect.Request[corev1.AddSubscriptionRequest]) (*connect.Response[corev1.AddSubscriptionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *SubscriptionHandler) UpdateSubscription(_ context.Context, _ *connect.Request[corev1.UpdateSubscriptionRequest]) (*connect.Response[corev1.UpdateSubscriptionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *SubscriptionHandler) DeleteSubscription(_ context.Context, _ *connect.Request[corev1.DeleteSubscriptionRequest]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *SubscriptionHandler) RefreshSubscription(_ context.Context, _ *connect.Request[corev1.RefreshSubscriptionRequest], _ *connect.ServerStream[corev1.RefreshProgress]) error {
	return connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *SubscriptionHandler) PreviewSubscription(_ context.Context, _ *connect.Request[corev1.PreviewSubscriptionRequest]) (*connect.Response[corev1.PreviewSubscriptionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}
