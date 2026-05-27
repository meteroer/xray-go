package singbox

import (
	"context"
	"fmt"
	"net/netip"

	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/include"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json/badoption"
	"xray-cli/subscription"
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

func Start(node *subscription.Node, socksPort int, httpPort int) (*Server, error) {
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

	listenAddr := badoption.Addr(netip.MustParseAddr("127.0.0.1"))

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
		Route: &option.RouteOptions{
			Final: "proxy",
		},
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