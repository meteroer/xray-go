package xrayproxy

import (
	"fmt"
	"net"
	"strings"

	_ "github.com/xtls/xray-core/main/json"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf/serial"
	"xray-cli/subscription"
)

type Server struct {
	instance *core.Instance
}

func buildStreamSettings(node *subscription.Node) string {
	var parts []string
	parts = append(parts, fmt.Sprintf(`"network": "%s"`, node.Network))

	if node.TLS {
		tlsConfig := `"security": "tls", "tlsSettings": {`
		if node.Host != "" {
			tlsConfig += fmt.Sprintf(`"serverName": "%s"`, node.Host)
		}
		tlsConfig += "}"
		parts = append(parts, tlsConfig)
	}

	switch node.Network {
	case "ws":
		wsConfig := `"wsSettings": {"connectionReuse": true, `
		if node.Path != "" {
			wsConfig += fmt.Sprintf(`"path": "%s", `, node.Path)
		}
		if node.Host != "" {
			wsConfig += fmt.Sprintf(`"headers": {"Host": "%s"}`, node.Host)
		}
		wsConfig += "}"
		parts = append(parts, wsConfig)
	case "grpc":
		grpcConfig := `"grpcSettings": {`
		if node.Path != "" {
			grpcConfig += fmt.Sprintf(`"serviceName": "%s"`, node.Path)
		}
		grpcConfig += "}"
		parts = append(parts, grpcConfig)
	}

	return "{" + strings.Join(parts, ", ") + "}"
}

func flowField(flow string) string {
	if flow == "" {
		return ""
	}
	return fmt.Sprintf(`, "flow": "%s"`, flow)
}

func buildOutbound(node *subscription.Node) string {
	streamSettings := buildStreamSettings(node)
	tag := `"tag": "proxy",`

	switch node.Protocol {
	case "vmess":
		return fmt.Sprintf(`{%s "protocol": "vmess", "settings": {
      "vnext": [{
        "address": "%s", "port": %d,
        "users": [{"id": "%s", "alterId": %d, "security": "%s"%s}]
      }]
    }, "streamSettings": %s}`, tag, node.Address, node.Port, node.UUID, node.AlterId, node.Security, flowField(node.Flow), streamSettings)

	case "vless":
		return fmt.Sprintf(`{%s "protocol": "vless", "settings": {
      "vnext": [{
        "address": "%s", "port": %d,
        "users": [{"id": "%s", "encryption": "none"%s}]
      }]
    }, "streamSettings": %s}`, tag, node.Address, node.Port, node.UUID, flowField(node.Flow), streamSettings)

	case "trojan":
		return fmt.Sprintf(`{%s "protocol": "trojan", "settings": {
      "servers": [{
        "address": "%s", "port": %d, "password": "%s"
      }]
    }, "streamSettings": %s}`, tag, node.Address, node.Port, node.UUID, streamSettings)

	case "shadowsocks":
		return fmt.Sprintf(`{%s "protocol": "shadowsocks", "settings": {
      "servers": [{
        "address": "%s", "port": %d, "method": "%s", "password": "%s"
      }]
    }}`, tag, node.Address, node.Port, node.Security, node.UUID)

	default:
		return `{` + tag + `"protocol": "freedom"}`
	}
}

func buildXrayConfig(node *subscription.Node, socksPort int, httpPort int) string {
	outbound := buildOutbound(node)
	return fmt.Sprintf(`{
  "log": {"loglevel": "warning"},
  "inbounds": [{
    "port": %d,
    "listen": "127.0.0.1",
    "protocol": "socks",
    "settings": {
      "udp": true,
      "auth": "noauth"
    },
    "sniffing": {
      "enabled": true,
      "destOverride": ["http", "tls"]
    }
  }, {
    "port": %d,
    "listen": "127.0.0.1",
    "protocol": "http",
    "settings": {}
  }],
  "outbounds": [%s],
  "routing": {
    "domainStrategy": "AsIs",
    "rules": [{
      "type": "field",
      "outboundTag": "proxy",
      "network": "tcp,udp"
    }]
  }
}`, socksPort, httpPort, outbound)
}

func Start(node *subscription.Node, socksPort int, httpPort int) (*Server, error) {
	configJSON := buildXrayConfig(node, socksPort, httpPort)
	cfg, err := serial.LoadJSONConfig(strings.NewReader(configJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to load xray config: %w", err)
	}
	inst, err := core.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create xray instance: %w", err)
	}
	if err := inst.Start(); err != nil {
		return nil, fmt.Errorf("failed to start xray: %w", err)
	}
	return &Server{instance: inst}, nil
}

func (s *Server) Stop() error {
	if s.instance != nil {
		return s.instance.Close()
	}
	return nil
}

func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
