package apiv1

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	corev1 "github.com/ryfineZ/weave/gen/go/proxy/core/v1"
	"github.com/ryfineZ/weave/internal/log"
)

// NodeHandler implements corev1connect.NodeServiceHandler.
type NodeHandler struct {
	log *log.Logger
}

func NewNodeHandler(l *log.Logger) *NodeHandler {
	return &NodeHandler{log: l}
}

func (h *NodeHandler) ListNodes(_ context.Context, _ *connect.Request[corev1.ListNodesRequest]) (*connect.Response[corev1.ListNodesResponse], error) {
	return connect.NewResponse(&corev1.ListNodesResponse{}), nil
}

func (h *NodeHandler) GetNode(_ context.Context, _ *connect.Request[corev1.GetNodeRequest]) (*connect.Response[corev1.GetNodeResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *NodeHandler) CreateNode(_ context.Context, _ *connect.Request[corev1.CreateNodeRequest]) (*connect.Response[corev1.CreateNodeResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *NodeHandler) UpdateNode(_ context.Context, _ *connect.Request[corev1.UpdateNodeRequest]) (*connect.Response[corev1.UpdateNodeResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *NodeHandler) DeleteNode(_ context.Context, _ *connect.Request[corev1.DeleteNodeRequest]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *NodeHandler) ListGroups(_ context.Context, _ *connect.Request[corev1.ListGroupsRequest]) (*connect.Response[corev1.ListGroupsResponse], error) {
	return connect.NewResponse(&corev1.ListGroupsResponse{}), nil
}

func (h *NodeHandler) CreateGroup(_ context.Context, _ *connect.Request[corev1.CreateGroupRequest]) (*connect.Response[corev1.CreateGroupResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *NodeHandler) UpdateGroup(_ context.Context, _ *connect.Request[corev1.UpdateGroupRequest]) (*connect.Response[corev1.UpdateGroupResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *NodeHandler) DeleteGroup(_ context.Context, _ *connect.Request[corev1.DeleteGroupRequest]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *NodeHandler) WatchProbeResults(_ context.Context, _ *connect.Request[corev1.WatchProbeResultsRequest], _ *connect.ServerStream[corev1.ProbeResult]) error {
	return connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *NodeHandler) TriggerProbe(_ context.Context, _ *connect.Request[corev1.TriggerProbeRequest]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}
