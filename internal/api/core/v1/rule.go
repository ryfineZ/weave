package apiv1

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	corev1 "github.com/ryfineZ/weave/gen/go/proxy/core/v1"
	"github.com/ryfineZ/weave/internal/log"
)

// RuleHandler implements corev1connect.RuleServiceHandler.
type RuleHandler struct {
	log *log.Logger
}

func NewRuleHandler(l *log.Logger) *RuleHandler {
	return &RuleHandler{log: l}
}

func (h *RuleHandler) ListIdentityRules(_ context.Context, _ *connect.Request[emptypb.Empty]) (*connect.Response[corev1.ListIdentityRulesResponse], error) {
	return connect.NewResponse(&corev1.ListIdentityRulesResponse{}), nil
}

func (h *RuleHandler) UpsertIdentityRule(_ context.Context, _ *connect.Request[corev1.UpsertIdentityRuleRequest]) (*connect.Response[corev1.UpsertIdentityRuleResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *RuleHandler) DeleteIdentityRule(_ context.Context, _ *connect.Request[corev1.DeleteIdentityRuleRequest]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *RuleHandler) ReorderIdentityRules(_ context.Context, _ *connect.Request[corev1.ReorderRulesRequest]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *RuleHandler) ListDestinationRules(_ context.Context, _ *connect.Request[emptypb.Empty]) (*connect.Response[corev1.ListDestinationRulesResponse], error) {
	return connect.NewResponse(&corev1.ListDestinationRulesResponse{}), nil
}

func (h *RuleHandler) UpsertDestinationRule(_ context.Context, _ *connect.Request[corev1.UpsertDestinationRuleRequest]) (*connect.Response[corev1.UpsertDestinationRuleResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *RuleHandler) DeleteDestinationRule(_ context.Context, _ *connect.Request[corev1.DeleteDestinationRuleRequest]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *RuleHandler) ReorderDestinationRules(_ context.Context, _ *connect.Request[corev1.ReorderRulesRequest]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *RuleHandler) ListRuleSets(_ context.Context, _ *connect.Request[emptypb.Empty]) (*connect.Response[corev1.ListRuleSetsResponse], error) {
	return connect.NewResponse(&corev1.ListRuleSetsResponse{}), nil
}

func (h *RuleHandler) ImportRuleSet(_ context.Context, _ *connect.Request[corev1.ImportRuleSetRequest]) (*connect.Response[corev1.ImportRuleSetResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *RuleHandler) UpdateRuleSet(_ context.Context, _ *connect.Request[corev1.UpdateRuleSetRequest]) (*connect.Response[corev1.UpdateRuleSetResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *RuleHandler) DeleteRuleSet(_ context.Context, _ *connect.Request[corev1.DeleteRuleSetRequest]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *RuleHandler) SetRuleSetEnabled(_ context.Context, _ *connect.Request[corev1.SetRuleSetEnabledRequest]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}
