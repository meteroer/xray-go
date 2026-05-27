package xrayproxy

import (
	"fmt"
	"net"
	"strings"

	_ "github.com/xtls/xray-core/main/distro/all"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf/serial"
	"xray-go/config"
	"xray-go/subscription"
)

type Server struct {
	instance *core.Instance
}

func buildStreamSettings(node *subscription.Node) string {
	var parts []string
	parts = append(parts, fmt.Sprintf(`"network": "%s"`, node.Network))

	if node.Reality {
		realityConfig := `"security": "reality", "realitySettings": {`
		realityConfig += fmt.Sprintf(`"serverName": "%s", `, node.Host)
		realityConfig += fmt.Sprintf(`"publicKey": "%s", `, node.PublicKey)
		realityConfig += fmt.Sprintf(`"shortId": "%s"`, node.ShortId)
		if node.Fingerprint != "" {
			realityConfig += fmt.Sprintf(`, "fingerprint": "%s"`, node.Fingerprint)
		}
		if node.Spx != "" {
			realityConfig += fmt.Sprintf(`, "spiderX": "%s"`, node.Spx)
		}
		realityConfig += "}"
		parts = append(parts, realityConfig)
	} else if node.TLS {
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

func buildRoutingRules(routeMode config.RouteMode, whitelist, blacklist []string) string {
	if routeMode == config.RouteModeGlobal {
		return `{"type": "field", "outboundTag": "proxy", "network": "tcp,udp"}`
	}

	items := whitelist
	outboundTag := "proxy"
	defaultTag := "direct"
	if routeMode == config.RouteModeBlacklist {
		items = blacklist
		outboundTag = "direct"
		defaultTag = "proxy"
	}

	var domains, geosites, geoips []string
	for _, item := range items {
		if strings.HasPrefix(item, "geosite:") {
			geosites = append(geosites, fmt.Sprintf(`"geosite:%s"`, strings.TrimPrefix(item, "geosite:")))
		} else if strings.HasPrefix(item, "geoip:") {
			geoips = append(geoips, fmt.Sprintf(`"geoip:%s"`, strings.TrimPrefix(item, "geoip:")))
		} else {
			domains = append(domains, fmt.Sprintf(`"%s"`, item))
		}
	}

	var rules []string
	rule := fmt.Sprintf(`{"type": "field", "outboundTag": "%s"`, outboundTag)
	if len(domains) > 0 {
		rule += fmt.Sprintf(`, "domain": [%s]`, strings.Join(domains, ", "))
	}
	if len(geosites) > 0 {
		rule += fmt.Sprintf(`, "domain": [%s]`, strings.Join(geosites, ", "))
	}
	if len(geoips) > 0 {
		rule += fmt.Sprintf(`, "ip": [%s]`, strings.Join(geoips, ", "))
	}
	rule += fmt.Sprintf(`, "network": "tcp,udp"}`)
	rules = append(rules, rule)

	rules = append(rules, fmt.Sprintf(`{"type": "field", "outboundTag": "%s", "network": "tcp,udp"}`, defaultTag))

	return strings.Join(rules, ",")
}

func buildXrayConfig(node *subscription.Node, socksPort int, httpPort int, routeMode config.RouteMode, whitelist, blacklist []string) string {
	outbound := buildOutbound(node)
	routingRules := buildRoutingRules(routeMode, whitelist, blacklist)
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
  "outbounds": [%s, {"protocol": "freedom", "tag": "direct"}],
  "routing": {
    "domainStrategy": "IPOnDemand",
    "rules": [%s]
  }
}`, socksPort, httpPort, outbound, routingRules)
}

func Start(node *subscription.Node, socksPort int, httpPort int, routeMode config.RouteMode, whitelist, blacklist []string) (*Server, error) {
	configJSON := buildXrayConfig(node, socksPort, httpPort, routeMode, whitelist, blacklist)
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
