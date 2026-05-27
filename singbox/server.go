package singbox

import (
	"context"
	"fmt"
	"net/netip"
	"strings"

	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/include"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json/badoption"
	"xray-go/config"
	"xray-go/subscription"
)

type Server struct {
	instance *box.Box
	cancel   context.CancelFunc
}

func buildTLSOptions(node *subscription.Node) *option.OutboundTLSOptions {
	tlsOpts := &option.OutboundTLSOptions{
		Enabled:    true,
		ServerName: node.SNI,
		Insecure:   node.Insecure,
	}
	if tlsOpts.ServerName == "" {
		tlsOpts.ServerName = node.Address
	}
	if node.Fingerprint != "" {
		tlsOpts.UTLS = &option.OutboundUTLSOptions{
			Enabled:     true,
			Fingerprint: node.Fingerprint,
		}
	}
	return tlsOpts
}

func buildRouteRules(routeMode config.RouteMode, whitelist, blacklist []string) (*option.RouteOptions, []option.Outbound) {
	directOutbound := option.Outbound{
		Type: "direct",
		Tag:  "direct",
		Options: &option.DirectOutboundOptions{},
	}

	var rules []option.Rule

	makeRule := func(item string, outbound string) option.Rule {
		rule := option.Rule{
			DefaultOptions: option.DefaultRule{
				RuleAction: option.RuleAction{
					RouteOptions: option.RouteActionOptions{
						Outbound: outbound,
					},
				},
			},
		}
		if strings.HasPrefix(item, "geosite:") || strings.HasPrefix(item, "geoip:") {
			// sing-box v1.12+ removed inline geosite/geoip support.
			// Skip these entries — the user can add explicit domains instead.
			return rule
		}
		rule.DefaultOptions.RawDefaultRule.DomainSuffix = badoption.Listable[string]{item}
		return rule
	}

	switch routeMode {
	case config.RouteModeWhitelist:
		for _, item := range whitelist {
			rule := makeRule(item, "proxy")
			// Skip empty/invalid rules
			if len(rule.DefaultOptions.RawDefaultRule.DomainSuffix) > 0 {
				rules = append(rules, rule)
			}
		}
		if len(rules) == 0 {
			// Fallback to global if no valid whitelist rules
			return &option.RouteOptions{Final: "proxy"}, nil
		}
		return &option.RouteOptions{
			Rules: rules,
			Final: "direct",
		}, []option.Outbound{directOutbound}

	case config.RouteModeBlacklist:
		for _, item := range blacklist {
			rule := makeRule(item, "direct")
			if len(rule.DefaultOptions.RawDefaultRule.DomainSuffix) > 0 {
				rules = append(rules, rule)
			}
		}
		return &option.RouteOptions{
			Rules: rules,
			Final: "proxy",
		}, []option.Outbound{directOutbound}

	default:
		return &option.RouteOptions{
			Final: "proxy",
		}, nil
	}
}

func Start(node *subscription.Node, socksPort int, httpPort int, routeMode config.RouteMode, whitelist, blacklist []string) (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = include.Context(ctx)

	var outbounds []option.Outbound
	switch node.Protocol {
	case "anytls":
		tlsOpts := buildTLSOptions(node)
		outbounds = append(outbounds, option.Outbound{
			Type: "anytls",
			Tag:  "proxy",
			Options: &option.AnyTLSOutboundOptions{
				ServerOptions: option.ServerOptions{
					Server:     node.Address,
					ServerPort: uint16(node.Port),
				},
				Password: node.UUID,
				OutboundTLSOptionsContainer: option.OutboundTLSOptionsContainer{
					TLS: tlsOpts,
				},
			},
		})
	default:
		cancel()
		return nil, fmt.Errorf("unsupported protocol for sing-box: %s", node.Protocol)
	}

	routeOpts, extraOutbounds := buildRouteRules(routeMode, whitelist, blacklist)
	outbounds = append(outbounds, extraOutbounds...)

	listenAddr := badoption.Addr(netip.MustParseAddr("0.0.0.0"))

	opts := option.Options{
		Log: &option.LogOptions{
			Level: "warning",
		},
		Inbounds: []option.Inbound{
			{
				Type: "socks",
				Tag:  "socks-in",
				Options: &option.SocksInboundOptions{
					ListenOptions: option.ListenOptions{
						Listen:     &listenAddr,
						ListenPort: uint16(socksPort),
					},
				},
			},
			{
				Type: "mixed",
				Tag:  "http-in",
				Options: &option.HTTPMixedInboundOptions{
					ListenOptions: option.ListenOptions{
						Listen:     &listenAddr,
						ListenPort: uint16(httpPort),
					},
				},
			},
		},
		Outbounds: outbounds,
		Route:    routeOpts,
	}

	instance, err := box.New(box.Options{
		Options: opts,
		Context: ctx,
	})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create sing-box instance: %w", err)
	}

	if err := instance.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start sing-box: %w", err)
	}

	return &Server{instance: instance, cancel: cancel}, nil
}

func (s *Server) Stop() error {
	if s.instance != nil {
		err := s.instance.Close()
		s.cancel()
		return err
	}
	s.cancel()
	return nil
}