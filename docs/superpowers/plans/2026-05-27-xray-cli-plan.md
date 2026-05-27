# xray-cli Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a single-binary CLI tool that fetches subscription nodes, tests HTTP latency, and starts a local proxy on port 16708 using the fastest node.

**Architecture:** Embed xray-core as a Go library. Generate xray JSON config from parsed nodes, use `serial.LoadJSONConfig()` to build `core.Config`, then `core.New()` + `Start()` to run the proxy. Latency testing starts temporary xray instances on random ports.

**Tech Stack:** Go 1.24+, github.com/xtls/xray-core v1.260327.0 (Xray-core v26.3.27)

---

## File Structure

```
xray-cli/
├── main.go                    -- Entry point, CLI flags, orchestration
├── config/
│   └── store.go               -- Save/load subscription URL to ~/.xray-cli/config.json
├── subscription/
│   ├── fetcher.go             -- HTTP GET subscription content
│   └── parser.go              -- Base64 decode, parse vmess/vless/trojan/ss links
├── latency/
│   └── tester.go              -- Test HTTP latency through temporary xray instances
├── proxy/
│   └── server.go              -- Build xray JSON config, start/stop xray-core instance
├── go.mod
├── go.sum
└── docs/
    └── superpowers/
        ├── specs/
        │   └── 2026-05-27-xray-cli-design.md
        └── plans/
            └── 2026-05-27-xray-cli-plan.md
```

---

### Task 1: Project Scaffolding & Config Store

**Files:**
- Create: `config/store.go`
- Modify: `go.mod` (already exists)

- [ ] **Step 1: Write config/store.go**

```go
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	SubscriptionURL string `json:"subscription_url"`
	SelectedNode    string `json:"selected_node"`
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".xray-cli")
	return dir, os.MkdirAll(dir, 0755)
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}
	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func PromptSubscriptionURL() (string, error) {
	fmt.Print("Enter subscription URL: ")
	var url string
	_, err := fmt.Scanln(&url)
	if err != nil {
		return "", err
	}
	if url == "" {
		return "", fmt.Errorf("subscription URL cannot be empty")
	}
	return url, nil
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /root/xray-cli && go build ./config/`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add config/store.go
git commit -m "feat: add config store for subscription URL persistence"
```

---

### Task 2: Subscription Fetcher

**Files:**
- Create: `subscription/fetcher.go`

- [ ] **Step 1: Write subscription/fetcher.go**

```go
package subscription

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func Fetch(url string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subscription: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("subscription returned status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read subscription body: %w", err)
	}
	return data, nil
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /root/xray-cli && go build ./subscription/`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add subscription/fetcher.go
git commit -m "feat: add subscription fetcher"
```

---

### Task 3: Subscription Parser

**Files:**
- Create: `subscription/parser.go`

- [ ] **Step 1: Write subscription/parser.go**

```go
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
	Name     string
	Protocol string
	Address  string
	Port     int
	UUID     string
	AlterId  int
	Security string
	Network  string
	Host     string
	Path     string
	TLS      bool
	Flow     string
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
	return nil, fmt.Errorf("unsupported protocol: %s", line[:min(strings.Index(line, "://"), 10)])
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
		Ps       string `json:"ps"`
		Add      string `json:"add"`
		Port     string `json:"port"`
		ID       string `json:"id"`
		Aid      string `json:"aid"`
		Scy      string `json:"scy"`
		Net      string `json:"net"`
		Type     string `json:"type"`
		Host     string `json:"host"`
		Path     string `json:"path"`
		TLS      string `json:"tls"`
		Flow     string `json:"flow"`
		Sni      string `json:"sni"`
		Alpn     string `json:"alpn"`
		Fp       string `json:"fp"`
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
	flow := query.Get("flow")
	tls := query.Get("security") == "tls" || query.Get("security") == "reality"
	return &Node{
		Name:     u.Fragment,
		Protocol: "vless",
		Address:  u.Hostname(),
		Port:     port,
		UUID:     u.User.Username(),
		Network:  network,
		Host:     host,
		Path:     query.Get("path"),
		TLS:      tls,
		Flow:     flow,
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
	hostPort := rest
	colonIdx := strings.LastIndex(hostPort, ":")
	if colonIdx == -1 {
		return nil, fmt.Errorf("invalid ss host:port")
	}
	host := hostPort[:colonIdx]
	port, _ := strconv.Atoi(hostPort[colonIdx+1:])
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
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /root/xray-cli && go build ./subscription/`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add subscription/parser.go
git commit -m "feat: add subscription parser for vmess/vless/trojan/ss"
```

---

### Task 4: Xray Config Builder & Proxy Server

**Files:**
- Create: `proxy/server.go`

- [ ] **Step 1: Write proxy/server.go**

This file builds xray JSON config from a Node, starts/stops xray-core instances.

```go
package proxy

import (
	"context"
	"fmt"
	"net"
	"strings"

	_ "github.com/xtls/xray-core/main/json"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf/serial"
)

type Server struct {
	instance *core.Instance
}

func buildXrayConfig(node *Node, listenPort int) string {
	streamSettings := buildStreamSettings(node)
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
}`, listenPort, listenPort, outbound)
}

func buildStreamSettings(node *Node) string {
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

func buildOutbound(node *Node) string {
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

func flowField(flow string) string {
	if flow == "" {
		return ""
	}
	return fmt.Sprintf(`, "flow": "%s"`, flow)
}

func Start(node *Node, port int) (*Server, error) {
	configJSON := buildXrayConfig(node, port)
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
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /root/xray-cli && go build ./proxy/`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add proxy/server.go
git commit -m "feat: add xray config builder and proxy server"
```

---

### Task 5: Latency Tester

**Files:**
- Create: `latency/tester.go`

- [ ] **Step 1: Write latency/tester.go**

```go
package latency

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"xray-cli/proxy"
	"xray-cli/subscription"
)

type Result struct {
	Node   *subscription.Node
	Latency time.Duration
	Err     error
}

func TestAll(nodes []*subscription.Node, maxConcurrent int) []*Result {
	if maxConcurrent <= 0 {
		maxConcurrent = 5
	}
	results := make([]*Result, len(nodes))
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrent)

	for i, node := range nodes {
		wg.Add(1)
		go func(idx int, n *subscription.Node) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			latency, err := testNode(n)
			results[idx] = &Result{Node: n, Latency: latency, Err: err}
		}(i, node)
	}
	wg.Wait()
	return results
}

func testNode(node *subscription.Node) (time.Duration, error) {
	port, err := proxy.GetFreePort()
	if err != nil {
		return 0, fmt.Errorf("get free port: %w", err)
	}

	srv, err := proxy.Start(node, port)
	if err != nil {
		return 0, fmt.Errorf("start proxy: %w", err)
	}
	defer srv.Stop()

	time.Sleep(100 * time.Millisecond)

	proxyURL := fmt.Sprintf("socks5://127.0.0.1:%d", port)
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return (&net.Dialer{Timeout: 3 * time.Second}).DialContext(ctx, "tcp", fmt.Sprintf("127.0.0.1:%d", port))
			},
		},
	}
	_ = proxyURL

	start := time.Now()
	resp, err := client.Get("http://www.gstatic.com/generate_204")
	if err != nil {
		return 0, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()
	elapsed := time.Since(start)

	if resp.StatusCode != 204 && resp.StatusCode != 200 {
		return 0, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return elapsed, nil
}

func FindBest(nodes []*subscription.Node) (*subscription.Node, time.Duration, error) {
	results := TestAll(nodes, 5)
	var bestNode *subscription.Node
	var bestLatency time.Duration
	var lastErr error
	for _, r := range results {
		if r.Err != nil {
			lastErr = r.Err
			fmt.Printf("  ✗ %s: %v\n", r.Node.Name, r.Err)
			continue
		}
		fmt.Printf("  ✓ %s: %v\n", r.Node.Name, r.Latency)
		if bestNode == nil || r.Latency < bestLatency {
			bestNode = r.Node
			bestLatency = r.Latency
		}
	}
	if bestNode == nil {
		return nil, 0, fmt.Errorf("all nodes unreachable: %v", lastErr)
	}
	return bestNode, bestLatency, nil
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /root/xray-cli && go build ./latency/`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add latency/tester.go
git commit -m "feat: add concurrent latency tester"
```

---

### Task 6: Main Entry Point

**Files:**
- Create: `main.go` (replace existing placeholder)

- [ ] **Step 1: Write main.go**

```go
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"xray-cli/config"
	"xray-cli/latency"
	"xray-cli/proxy"
	"xray-cli/subscription"
)

func main() {
	urlFlag := flag.String("url", "", "subscription URL (overrides saved config)")
	portFlag := flag.Int("port", 16708, "local proxy port")
	updateFlag := flag.Bool("update", false, "force re-fetch subscription and re-test latency")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	subURL := *urlFlag
	if subURL == "" {
		subURL = cfg.SubscriptionURL
	}
	if subURL == "" || *updateFlag {
		url, err := config.PromptSubscriptionURL()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		subURL = url
	}

	cfg.SubscriptionURL = subURL
	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save config: %v\n", err)
	}

	fmt.Println("Fetching subscription...")
	data, err := subscription.Fetch(subURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching subscription: %v\n", err)
		os.Exit(1)
	}

	nodes, err := subscription.Parse(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing subscription: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Found %d nodes\n", len(nodes))

	fmt.Println("Testing latency...")
	bestNode, bestLatency, err := latency.FindBest(nodes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Best node: %s (%v)\n", bestNode.Name, bestLatency)

	cfg.SelectedNode = bestNode.Name
	config.Save(cfg)

	fmt.Printf("Starting proxy on 127.0.0.1:%d...\n", *portFlag)
	srv, err := proxy.Start(bestNode, *portFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting proxy: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Proxy running at 127.0.0.1:%d (SOCKS5+HTTP)\n", *portFlag)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	fmt.Println("\nShutting down...")
	srv.Stop()
	fmt.Println("Done.")
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /root/xray-cli && go build -o xray-cli .`
Expected: binary created successfully

- [ ] **Step 3: Commit**

```bash
git add main.go
git commit -m "feat: add main entry point with CLI flags and orchestration"
```

---

### Task 7: Fix Compilation & Integration

**Files:**
- Modify: various files as needed

- [ ] **Step 1: Run full build**

Run: `cd /root/xray-cli && go build -o xray-cli .`
Expected: may have compilation errors to fix

- [ ] **Step 2: Fix any compilation errors**

Common issues to expect:
- Import path issues (need to ensure module name matches)
- Missing imports for xray-core sub-packages
- The `proxy` package name conflicts with xray-core's `proxy` package — may need to rename our package

If the `proxy` package name conflicts with xray-core, rename our `proxy/` directory to `xrayproxy/` and update all imports.

- [ ] **Step 3: Verify clean build**

Run: `cd /root/xray-cli && go build -o xray-cli .`
Expected: clean build, binary produced

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "fix: resolve compilation issues and integration"
```

---

### Task 8: Final Verification

- [ ] **Step 1: Verify binary runs and shows help**

Run: `cd /root/xray-cli && ./xray-cli -h`
Expected: flag usage output showing --url, --port, --update

- [ ] **Step 2: Verify binary size is reasonable**

Run: `ls -lh /root/xray-cli/xray-cli`
Expected: ~20-30MB single binary

- [ ] **Step 3: Commit final state**

```bash
git add -A
git commit -m "chore: final verification"
```
