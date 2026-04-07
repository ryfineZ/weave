// Package orchestrator translates Weave's Chain/Node model into a sing-box
// Options struct and manages the sing-box engine lifecycle.
//
// Design:
//   - One sing-box Box instance runs at a time.
//   - When chains or rules change, Apply() rebuilds the Options and performs a
//     hot-reload (stop old box, start new one). Fast enough for interactive
//     editing (< 500 ms on typical hardware).
//   - The box runs with TUN + SOCKS/HTTP inbounds as configured.
package orchestrator

import (
	"context"
	"fmt"
	"net/netip"
	"sync"

	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json/badoption"
	"go.uber.org/zap"

	corev1 "github.com/ryfineZ/weave/gen/go/proxy/core/v1"
	"github.com/ryfineZ/weave/internal/log"
)

// Config holds the inputs the orchestrator needs to build sing-box options.
type Config struct {
	Nodes          []*corev1.Node
	Groups         []*corev1.NodeGroup
	Chains         []*corev1.Chain
	IdentityRules  []*corev1.IdentityRule
	DestRules      []*corev1.DestinationRule
	RuleSets       []*corev1.RuleSet
	ListenSocks    string // e.g. "127.0.0.1:7890"
	ListenHTTP     string // e.g. "127.0.0.1:7891"
	EnableTUN      bool
	EnableSysproxy bool
}

// Orchestrator manages a running sing-box instance.
type Orchestrator struct {
	mu  sync.Mutex
	box *box.Box
	log *log.Logger
}

// New creates an idle Orchestrator. Call Apply() to start sing-box.
func New(l *log.Logger) *Orchestrator {
	return &Orchestrator{log: l}
}

// Apply rebuilds sing-box options from cfg and performs a hot-reload.
// If no instance is running, it starts one. Safe to call concurrently.
func (o *Orchestrator) Apply(ctx context.Context, cfg Config) error {
	opts, err := buildOptions(cfg)
	if err != nil {
		return fmt.Errorf("orchestrator: build options: %w", err)
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	if o.box != nil {
		o.log.Info("orchestrator: stopping previous sing-box instance")
		o.box.Close()
		o.box = nil
	}

	if len(cfg.Chains) == 0 || !hasEnabledChain(cfg.Chains) {
		o.log.Info("orchestrator: no enabled chains — sing-box not started")
		return nil
	}

	b, err := box.New(box.Options{
		Context: ctx,
		Options: opts,
	})
	if err != nil {
		return fmt.Errorf("orchestrator: create box: %w", err)
	}
	if err := b.Start(); err != nil {
		return fmt.Errorf("orchestrator: start box: %w", err)
	}

	o.box = b
	o.log.Info("orchestrator: sing-box started",
		zap.Int("chains", len(cfg.Chains)),
		zap.Int("nodes", len(cfg.Nodes)),
	)
	return nil
}

// Stop shuts down the running sing-box instance (if any).
func (o *Orchestrator) Stop() {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.box != nil {
		o.box.Close()
		o.box = nil
		o.log.Info("orchestrator: sing-box stopped")
	}
}

// IsRunning reports whether sing-box is currently active.
func (o *Orchestrator) IsRunning() bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.box != nil
}

// ─── Option builder ───────────────────────────────────────────────────────────

func buildOptions(cfg Config) (option.Options, error) {
	nodeByID := make(map[string]*corev1.Node, len(cfg.Nodes))
	for _, n := range cfg.Nodes {
		nodeByID[n.Id] = n
	}

	var outbounds []option.Outbound
	outboundIndex := make(map[string]string) // chainID → entry outbound tag

	for _, chain := range cfg.Chains {
		if !chain.Enabled || len(chain.Hops) == 0 {
			continue
		}
		tags, obs, err := buildChainOutbounds(chain, nodeByID)
		if err != nil {
			return option.Options{}, fmt.Errorf("chain %q: %w", chain.Name, err)
		}
		outbounds = append(outbounds, obs...)
		outboundIndex[chain.Id] = tags[0]
	}

	outbounds = append(outbounds,
		option.Outbound{Tag: "DIRECT", Type: "direct"},
		option.Outbound{Tag: "REJECT", Type: "block"},
	)

	routes, err := buildRoutes(cfg, outboundIndex)
	if err != nil {
		return option.Options{}, err
	}

	return option.Options{
		Log: &option.LogOptions{
			Level:  "info",
			Output: "stderr",
		},
		Inbounds:  buildInbounds(cfg),
		Outbounds: outbounds,
		Route:     routes,
	}, nil
}

func buildChainOutbounds(chain *corev1.Chain, nodeByID map[string]*corev1.Node) ([]string, []option.Outbound, error) {
	hops := chain.Hops
	tags := make([]string, len(hops))
	obs := make([]option.Outbound, len(hops))

	for i, hop := range hops {
		nodeID := hop.GetNodeId()
		if nodeID == "" {
			return nil, nil, fmt.Errorf("hop %d: group hops not supported in 0.1.0", i)
		}
		node, ok := nodeByID[nodeID]
		if !ok {
			return nil, nil, fmt.Errorf("hop %d: node %q not found", i, nodeID)
		}

		tag := fmt.Sprintf("%s-hop%d", chain.Id, i)
		tags[i] = tag

		ob, err := nodeToOutbound(node, tag)
		if err != nil {
			return nil, nil, fmt.Errorf("hop %d (node %q): %w", i, node.Name, err)
		}

		if i < len(hops)-1 {
			nextTag := fmt.Sprintf("%s-hop%d", chain.Id, i+1)
			setDetour(&ob, nextTag)
		}
		obs[i] = ob
	}
	return tags, obs, nil
}

func nodeToOutbound(n *corev1.Node, tag string) (option.Outbound, error) {
	switch n.Protocol {
	case corev1.Protocol_PROTOCOL_VLESS:
		return vlessOutbound(n, tag)
	case corev1.Protocol_PROTOCOL_VMESS:
		return vmessOutbound(n, tag)
	case corev1.Protocol_PROTOCOL_TROJAN:
		return trojanOutbound(n, tag)
	case corev1.Protocol_PROTOCOL_SHADOWSOCKS:
		return shadowsocksOutbound(n, tag)
	case corev1.Protocol_PROTOCOL_HYSTERIA2:
		return hysteria2Outbound(n, tag)
	case corev1.Protocol_PROTOCOL_TUIC:
		return tuicOutbound(n, tag)
	case corev1.Protocol_PROTOCOL_SOCKS5:
		return socks5Outbound(n, tag)
	case corev1.Protocol_PROTOCOL_HTTP:
		return httpOutbound(n, tag)
	default:
		return option.Outbound{}, fmt.Errorf("unsupported protocol %v", n.Protocol)
	}
}

// ─── Route builder ────────────────────────────────────────────────────────────

func buildRoutes(cfg Config, outboundIndex map[string]string) (*option.RouteOptions, error) {
	var rules []option.Rule

	// Stage 1: identity rules (process name) — always evaluated first.
	for _, ir := range cfg.IdentityRules {
		outTag, err := resolveAction(ir.Action, outboundIndex)
		if err != nil {
			return nil, fmt.Errorf("identity rule %q: %w", ir.Id, err)
		}
		rules = append(rules, option.Rule{
			DefaultOptions: option.DefaultRule{
				RawDefaultRule: option.RawDefaultRule{
					ProcessName: badoption.Listable[string]{ir.ProcessName},
				},
				RuleAction: makeRouteAction(outTag),
			},
		})
	}

	// Stage 2: destination rules — in user-defined order.
	for _, dr := range cfg.DestRules {
		outTag, err := resolveAction(dr.Action, outboundIndex)
		if err != nil {
			return nil, fmt.Errorf("destination rule %q: %w", dr.Id, err)
		}
		rule, err := destRuleToSingBox(dr, outTag)
		if err != nil {
			return nil, fmt.Errorf("destination rule %q: %w", dr.Id, err)
		}
		rules = append(rules, rule)
	}

	defaultOut := "DIRECT"
	for _, chain := range cfg.Chains {
		if chain.Enabled {
			if tag, ok := outboundIndex[chain.Id]; ok {
				defaultOut = tag
				break
			}
		}
	}

	return &option.RouteOptions{
		Rules:               rules,
		Final:               defaultOut,
		AutoDetectInterface: true,
		FindProcess:         true, // required for process-name matching
	}, nil
}

func destRuleToSingBox(dr *corev1.DestinationRule, outTag string) (option.Rule, error) {
	raw := option.RawDefaultRule{}
	switch dr.Kind {
	case corev1.DestinationMatchKind_DESTINATION_MATCH_KIND_DOMAIN_SUFFIX:
		raw.DomainSuffix = badoption.Listable[string]{dr.Value}
	case corev1.DestinationMatchKind_DESTINATION_MATCH_KIND_DOMAIN_KEYWORD:
		raw.DomainKeyword = badoption.Listable[string]{dr.Value}
	case corev1.DestinationMatchKind_DESTINATION_MATCH_KIND_DOMAIN_REGEX:
		raw.DomainRegex = badoption.Listable[string]{dr.Value}
	case corev1.DestinationMatchKind_DESTINATION_MATCH_KIND_IP_CIDR:
		raw.IPCIDR = badoption.Listable[string]{dr.Value}
	case corev1.DestinationMatchKind_DESTINATION_MATCH_KIND_GEOIP:
		raw.IPCIDR = badoption.Listable[string]{"geoip:" + dr.Value}
	case corev1.DestinationMatchKind_DESTINATION_MATCH_KIND_RULESET:
		raw.RuleSet = badoption.Listable[string]{dr.Value}
	default:
		return option.Rule{}, fmt.Errorf("unknown match kind %v", dr.Kind)
	}
	return option.Rule{
		DefaultOptions: option.DefaultRule{
			RawDefaultRule: raw,
			RuleAction:     makeRouteAction(outTag),
		},
	}, nil
}

// makeRouteAction builds an option.RuleAction for routing to a named outbound.
// "DIRECT" and "REJECT" map to their native sing-box action types.
func makeRouteAction(outTag string) option.RuleAction {
	switch outTag {
	case "DIRECT":
		return option.RuleAction{Action: "direct"}
	case "REJECT":
		return option.RuleAction{Action: "reject"}
	default:
		return option.RuleAction{
			Action: "route",
			RouteOptions: option.RouteActionOptions{
				Outbound: outTag,
			},
		}
	}
}

// ─── Inbound builder ─────────────────────────────────────────────────────────

func buildInbounds(cfg Config) []option.Inbound {
	var inbounds []option.Inbound

	if cfg.ListenSocks != "" {
		listenAddr, listenPort := parseListenAddr(cfg.ListenSocks)
		inbounds = append(inbounds, option.Inbound{
			Tag:  "socks-in",
			Type: "socks",
			Options: &option.SocksInboundOptions{
				ListenOptions: option.ListenOptions{
					Listen:     listenAddr,
					ListenPort: listenPort,
				},
			},
		})
	}

	if cfg.ListenHTTP != "" {
		listenAddr, listenPort := parseListenAddr(cfg.ListenHTTP)
		inbounds = append(inbounds, option.Inbound{
			Tag:  "http-in",
			Type: "http",
			Options: &option.HTTPMixedInboundOptions{
				ListenOptions: option.ListenOptions{
					Listen:     listenAddr,
					ListenPort: listenPort,
				},
				SetSystemProxy: cfg.EnableSysproxy,
			},
		})
	}

	if cfg.EnableTUN {
		v4 := netip.MustParsePrefix("172.19.0.1/30")
		v6 := netip.MustParsePrefix("fdfe:dcba:9876::1/126")
		inbounds = append(inbounds, option.Inbound{
			Tag:  "tun-in",
			Type: "tun",
			Options: &option.TunInboundOptions{
				InterfaceName: "utun-weave",
				Address:       badoption.Listable[netip.Prefix]{v4, v6},
				MTU:           9000,
				AutoRoute:     true,
				StrictRoute:   true,
				Stack:         "mixed",
			},
		})
	}

	return inbounds
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func resolveAction(action *corev1.RuleAction, outboundIndex map[string]string) (string, error) {
	if action == nil {
		return "DIRECT", nil
	}
	switch action.Kind {
	case corev1.RuleActionKind_RULE_ACTION_KIND_DIRECT:
		return "DIRECT", nil
	case corev1.RuleActionKind_RULE_ACTION_KIND_REJECT:
		return "REJECT", nil
	case corev1.RuleActionKind_RULE_ACTION_KIND_PROXY:
		tag, ok := outboundIndex[action.ChainId]
		if !ok {
			return "", fmt.Errorf("chain %q not found or not enabled", action.ChainId)
		}
		return tag, nil
	default:
		return "DIRECT", nil
	}
}

func hasEnabledChain(chains []*corev1.Chain) bool {
	for _, c := range chains {
		if c.Enabled {
			return true
		}
	}
	return false
}
