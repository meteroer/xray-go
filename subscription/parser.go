package subscription

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type Node struct {
	Name        string `json:"name"`
	Protocol    string `json:"protocol"`
	Address     string `json:"address"`
	Port        int    `json:"port"`
	UUID        string `json:"uuid"`
	AlterId     int    `json:"alter_id"`
	Security    string `json:"security"`
	Network     string `json:"network"`
	Host        string `json:"host"`
	Path        string `json:"path"`
	TLS         bool   `json:"tls"`
	Flow        string `json:"flow"`
	Reality     bool   `json:"reality"`
	PublicKey   string `json:"public_key"`
	ShortId     string `json:"short_id"`
	Fingerprint string `json:"fingerprint"`
	Spx         string `json:"spx"`
	Insecure    bool   `json:"insecure"`
	SNI         string `json:"sni"`
}

func Parse(data []byte) ([]*Node, error) {
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(data)))
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(strings.TrimSpace(string(data)))
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64: %w", err)
		}
	}
	lines := strings.Split(strings.TrimSpace(string(decoded)), "\n")
	var nodes []*Node
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		node, err := parseLine(line)
		if err != nil {
			continue
		}
		nodes = append(nodes, node)
	}
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no valid nodes found")
	}
	return nodes, nil
}

func parseLine(line string) (*Node, error) {
	if strings.HasPrefix(line, "vmess://") {
		return parseVmess(line[8:])
	}
	if strings.HasPrefix(line, "vless://") {
		return parseVless(line[8:])
	}
	if strings.HasPrefix(line, "trojan://") {
		return parseTrojan(line[9:])
	}
	if strings.HasPrefix(line, "ss://") {
		return parseShadowsocks(line[5:])
	}
	if strings.HasPrefix(line, "anytls://") {
		return parseAnyTLS(line[9:])
	}
	return nil, fmt.Errorf("unsupported protocol")
}

func parseVmess(data string) (*Node, error) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(data)
		if err != nil {
			return nil, err
		}
	}
	var v struct {
		Ps   string `json:"ps"`
		Add  string `json:"add"`
		Port string `json:"port"`
		ID   string `json:"id"`
		Aid  string `json:"aid"`
		Scy  string `json:"scy"`
		Net  string `json:"net"`
		Host string `json:"host"`
		Path string `json:"path"`
		TLS  string `json:"tls"`
		Flow string `json:"flow"`
		Sni  string `json:"sni"`
	}
	if err := json.Unmarshal(decoded, &v); err != nil {
		return nil, err
	}
	port, _ := strconv.Atoi(v.Port)
	aid, _ := strconv.Atoi(v.Aid)
	host := v.Host
	if host == "" {
		host = v.Sni
	}
	network := v.Net
	if network == "" {
		network = "tcp"
	}
	security := v.Scy
	if security == "" {
		security = "auto"
	}
	return &Node{
		Name:     v.Ps,
		Protocol: "vmess",
		Address:  v.Add,
		Port:     port,
		UUID:     v.ID,
		AlterId:  aid,
		Security: security,
		Network:  network,
		Host:     host,
		Path:     v.Path,
		TLS:      v.TLS == "tls",
		Flow:     v.Flow,
	}, nil
}

func parseVless(data string) (*Node, error) {
	u, err := url.Parse("vless://" + data)
	if err != nil {
		return nil, err
	}
	port, _ := strconv.Atoi(u.Port())
	query := u.Query()
	network := query.Get("type")
	if network == "" {
		network = "tcp"
	}
	host := query.Get("host")
	if host == "" {
		host = query.Get("sni")
	}
	security := query.Get("security")
	tls := security == "tls"
	reality := security == "reality"
	sni := query.Get("sni")
	if sni == "" {
		sni = host
	}
	return &Node{
		Name:        u.Fragment,
		Protocol:    "vless",
		Address:     u.Hostname(),
		Port:        port,
		UUID:        u.User.Username(),
		Network:     network,
		Host:        sni,
		Path:        query.Get("path"),
		TLS:         tls,
		Flow:        query.Get("flow"),
		Reality:     reality,
		PublicKey:   query.Get("pbk"),
		ShortId:     query.Get("sid"),
		Fingerprint: query.Get("fp"),
		Spx:         query.Get("spx"),
	}, nil
}

func parseTrojan(data string) (*Node, error) {
	u, err := url.Parse("trojan://" + data)
	if err != nil {
		return nil, err
	}
	port, _ := strconv.Atoi(u.Port())
	query := u.Query()
	network := query.Get("type")
	if network == "" {
		network = "tcp"
	}
	host := query.Get("host")
	if host == "" {
		host = query.Get("sni")
	}
	return &Node{
		Name:     u.Fragment,
		Protocol: "trojan",
		Address:  u.Hostname(),
		Port:     port,
		UUID:     u.User.Username(),
		Network:  network,
		Host:     host,
		Path:     query.Get("path"),
		TLS:      true,
	}, nil
}

func parseShadowsocks(data string) (*Node, error) {
	atIdx := strings.LastIndex(data, "@")
	if atIdx == -1 {
		return nil, fmt.Errorf("invalid ss format")
	}
	encoded := data[:atIdx]
	rest := data[atIdx+1:]
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, err
		}
	}
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid ss method")
	}
	method := parts[0]
	password := parts[1]
	name := ""
	if hashIdx := strings.LastIndex(rest, "#"); hashIdx != -1 {
		name, _ = url.PathUnescape(rest[hashIdx+1:])
		rest = rest[:hashIdx]
	}
	colonIdx := strings.LastIndex(rest, ":")
	if colonIdx == -1 {
		return nil, fmt.Errorf("invalid ss host:port")
	}
	host := rest[:colonIdx]
	port, _ := strconv.Atoi(rest[colonIdx+1:])
	return &Node{
		Name:     name,
		Protocol: "shadowsocks",
		Address:  host,
		Port:     port,
		UUID:     password,
		Security: method,
		Network:  "tcp",
	}, nil
}

func parseAnyTLS(data string) (*Node, error) {
	u, err := url.Parse("anytls://" + data)
	if err != nil {
		return nil, err
	}
	port, _ := strconv.Atoi(u.Port())
	if port == 0 {
		port = 443
	}
	query := u.Query()
	insecure := query.Get("insecure") == "1"
	sni := query.Get("sni")
	return &Node{
		Name:     u.Fragment,
		Protocol: "anytls",
		Address:  u.Hostname(),
		Port:     port,
		UUID:     u.User.Username(),
		Network:  "tcp",
		TLS:      true,
		Insecure: insecure,
		SNI:      sni,
	}, nil
}
