package orchestrator

import (
	"fmt"
	"net"
	"net/netip"
	"strconv"

	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json/badoption"

	corev1 "github.com/ryfineZ/weave/gen/go/proxy/core/v1"
)

// ─── TLS helper ──────────────────────────────────────────────────────────────

func tlsOpts(tls *corev1.TlsConfig) *option.OutboundTLSOptions {
	if tls == nil {
		return nil
	}
	opts := &option.OutboundTLSOptions{
		Enabled:    true,
		ServerName: tls.ServerName,
		Insecure:   tls.Insecure,
		ALPN:       badoption.Listable[string](tls.Alpn),
	}
	if tls.Fingerprint != "" {
		opts.UTLS = &option.OutboundUTLSOptions{
			Enabled:     true,
			Fingerprint: tls.Fingerprint,
		}
	}
	if tls.RealityPublicKey != "" {
		opts.Reality = &option.OutboundRealityOptions{
			Enabled:   true,
			PublicKey: tls.RealityPublicKey,
			ShortID:   tls.RealityShortId,
		}
	}
	return opts
}

// ─── Transport helper ─────────────────────────────────────────────────────────

func transportOpts(t *corev1.TransportConfig) *option.V2RayTransportOptions {
	if t == nil || t.Type == "" {
		return nil
	}
	switch t.Type {
	case "ws":
		headers := make(badoption.HTTPHeader)
		for k, v := range t.WsHeaders {
			headers[k] = badoption.Listable[string]{v}
		}
		return &option.V2RayTransportOptions{
			Type: "ws",
			WebsocketOptions: option.V2RayWebsocketOptions{
				Path:    t.WsPath,
				Headers: headers,
			},
		}
	case "grpc":
		return &option.V2RayTransportOptions{
			Type: "grpc",
			GRPCOptions: option.V2RayGRPCOptions{
				ServiceName: t.GrpcServiceName,
			},
		}
	case "http":
		return &option.V2RayTransportOptions{
			Type: "http",
			HTTPOptions: option.V2RayHTTPOptions{
				// V2RayHTTPOptions.Path is a single string, HttpPath[0] used if available
				Path: func() string {
					if len(t.HttpPath) > 0 {
						return t.HttpPath[0]
					}
					return "/"
				}(),
			},
		}
	}
	return nil
}

// ─── Detour setter ────────────────────────────────────────────────────────────

// setDetour sets the Detour field on the DialerOptions embedded in ob.Options.
func setDetour(ob *option.Outbound, detour string) {
	switch ob.Type {
	case "vless":
		ob.Options.(*option.VLESSOutboundOptions).DialerOptions.Detour = detour
	case "vmess":
		ob.Options.(*option.VMessOutboundOptions).DialerOptions.Detour = detour
	case "trojan":
		ob.Options.(*option.TrojanOutboundOptions).DialerOptions.Detour = detour
	case "shadowsocks":
		ob.Options.(*option.ShadowsocksOutboundOptions).DialerOptions.Detour = detour
	case "hysteria2":
		ob.Options.(*option.Hysteria2OutboundOptions).DialerOptions.Detour = detour
	case "tuic":
		ob.Options.(*option.TUICOutboundOptions).DialerOptions.Detour = detour
	case "socks":
		ob.Options.(*option.SOCKSOutboundOptions).DialerOptions.Detour = detour
	case "http":
		ob.Options.(*option.HTTPOutboundOptions).DialerOptions.Detour = detour
	}
}

// ─── Protocol outbound constructors ──────────────────────────────────────────

func vlessOutbound(n *corev1.Node, tag string) (option.Outbound, error) {
	cfg := n.GetVless()
	if cfg == nil {
		return option.Outbound{}, fmt.Errorf("missing vless config")
	}
	opts := &option.VLESSOutboundOptions{
		ServerOptions: option.ServerOptions{
			Server:     n.Address,
			ServerPort: uint16(n.Port),
		},
		UUID:      cfg.Uuid,
		Flow:      cfg.Flow,
		Transport: transportOpts(cfg.Transport),
	}
	opts.TLS = tlsOpts(cfg.Tls)
	return option.Outbound{Tag: tag, Type: "vless", Options: opts}, nil
}

func vmessOutbound(n *corev1.Node, tag string) (option.Outbound, error) {
	cfg := n.GetVmess()
	if cfg == nil {
		return option.Outbound{}, fmt.Errorf("missing vmess config")
	}
	opts := &option.VMessOutboundOptions{
		ServerOptions: option.ServerOptions{
			Server:     n.Address,
			ServerPort: uint16(n.Port),
		},
		UUID:      cfg.Uuid,
		AlterId:   int(cfg.AlterId),
		Security:  cfg.Security,
		Transport: transportOpts(cfg.Transport),
	}
	opts.TLS = tlsOpts(cfg.Tls)
	return option.Outbound{Tag: tag, Type: "vmess", Options: opts}, nil
}

func trojanOutbound(n *corev1.Node, tag string) (option.Outbound, error) {
	cfg := n.GetTrojan()
	if cfg == nil {
		return option.Outbound{}, fmt.Errorf("missing trojan config")
	}
	opts := &option.TrojanOutboundOptions{
		ServerOptions: option.ServerOptions{
			Server:     n.Address,
			ServerPort: uint16(n.Port),
		},
		// Password is resolved from Keychain and set by the caller before Apply().
		Transport: transportOpts(cfg.Transport),
	}
	opts.TLS = tlsOpts(cfg.Tls)
	return option.Outbound{Tag: tag, Type: "trojan", Options: opts}, nil
}

func shadowsocksOutbound(n *corev1.Node, tag string) (option.Outbound, error) {
	cfg := n.GetShadowsocks()
	if cfg == nil {
		return option.Outbound{}, fmt.Errorf("missing shadowsocks config")
	}
	return option.Outbound{
		Tag:  tag,
		Type: "shadowsocks",
		Options: &option.ShadowsocksOutboundOptions{
			ServerOptions: option.ServerOptions{
				Server:     n.Address,
				ServerPort: uint16(n.Port),
			},
			Method: cfg.Method,
			// Password resolved from Keychain and set before Apply().
		},
	}, nil
}

func hysteria2Outbound(n *corev1.Node, tag string) (option.Outbound, error) {
	cfg := n.GetHysteria2()
	if cfg == nil {
		return option.Outbound{}, fmt.Errorf("missing hysteria2 config")
	}
	opts := &option.Hysteria2OutboundOptions{
		ServerOptions: option.ServerOptions{
			Server:     n.Address,
			ServerPort: uint16(n.Port),
		},
		UpMbps:   int(cfg.UpMbps),
		DownMbps: int(cfg.DownMbps),
	}
	opts.TLS = tlsOpts(cfg.Tls)
	if cfg.Obfs != nil {
		opts.Obfs = &option.Hysteria2Obfs{
			Type: cfg.Obfs.Type,
			// Password resolved from Keychain and set before Apply().
		}
	}
	return option.Outbound{Tag: tag, Type: "hysteria2", Options: opts}, nil
}

func tuicOutbound(n *corev1.Node, tag string) (option.Outbound, error) {
	cfg := n.GetTuic()
	if cfg == nil {
		return option.Outbound{}, fmt.Errorf("missing tuic config")
	}
	opts := &option.TUICOutboundOptions{
		ServerOptions: option.ServerOptions{
			Server:     n.Address,
			ServerPort: uint16(n.Port),
		},
		UUID:              cfg.Uuid,
		CongestionControl: cfg.CongestionControl,
		// Password resolved from Keychain.
	}
	opts.TLS = tlsOpts(cfg.Tls)
	return option.Outbound{Tag: tag, Type: "tuic", Options: opts}, nil
}

func socks5Outbound(n *corev1.Node, tag string) (option.Outbound, error) {
	cfg := n.GetSocks5()
	username := ""
	if cfg != nil {
		username = cfg.Username
	}
	return option.Outbound{
		Tag:  tag,
		Type: "socks",
		Options: &option.SOCKSOutboundOptions{
			ServerOptions: option.ServerOptions{
				Server:     n.Address,
				ServerPort: uint16(n.Port),
			},
			Username: username,
			Version:  "5",
		},
	}, nil
}

func httpOutbound(n *corev1.Node, tag string) (option.Outbound, error) {
	return option.Outbound{
		Tag:  tag,
		Type: "http",
		Options: &option.HTTPOutboundOptions{
			ServerOptions: option.ServerOptions{
				Server:     n.Address,
				ServerPort: uint16(n.Port),
			},
		},
	}, nil
}

// ─── Address helpers ──────────────────────────────────────────────────────────

func parseListenAddr(hostport string) (*badoption.Addr, uint16) {
	host, portStr, err := net.SplitHostPort(hostport)
	if err != nil {
		return nil, 0
	}
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return nil, 0
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, 0
	}
	ba := badoption.Addr(addr)
	return &ba, uint16(port)
}
