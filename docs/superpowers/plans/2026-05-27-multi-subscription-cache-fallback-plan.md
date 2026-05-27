# Multi-Subscription Caching & Fallback Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend xray-go with multi-subscription management, node caching, and fallback proxy fetching.

**Architecture:** config/store.go handles data model + persistence. subscription/fetcher.go gains FetchWithProxy. main.go orchestrates the new interactive flow: subscription selection → fetch/cache/fallback → region → latency → proxy.

**Tech Stack:** Go 1.26, xray-core, sing-box, golang.org/x/net/proxy (all already in go.mod)

---

### Task 1: Add JSON tags to Node struct

**Files:**
- Modify: `subscription/parser.go` (add JSON tags to Node struct fields)

- [ ] **Step 1: Add JSON tags to Node struct**

The Node struct needs JSON tags for serialization/deserialization when cached in config.json.

Replace the Node struct in `subscription/parser.go`:

```go
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
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./...`
Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add subscription/parser.go
git commit -m "feat: add JSON tags to Node struct for serialization"
```

---

### Task 2: Rewrite config/store.go

**Files:**
- Modify: `config/store.go` (complete rewrite)

- [ ] **Step 1: Write the new config/store.go**

```go
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"xray-go/subscription"
)

type Subscription struct {
	Name        string               `json:"name"`
	URL         string               `json:"url"`
	Nodes       []*subscription.Node `json:"nodes,omitempty"`
	LastNode    string               `json:"last_node"`
	LastFetched time.Time            `json:"last_fetched"`
}

type Config struct {
	Subscriptions   []*Subscription `json:"subscriptions"`
	LastUsedSub     string          `json:"last_used_subscription"`
	SubscriptionURL string          `json:"subscription_url,omitempty"`
	SelectedNode    string          `json:"selected_node,omitempty"`
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".xray-go")
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
	if cfg.SubscriptionURL != "" && len(cfg.Subscriptions) == 0 {
		sub := &Subscription{
			Name:     "default",
			URL:      cfg.SubscriptionURL,
			LastNode: cfg.SelectedNode,
		}
		cfg.Subscriptions = append(cfg.Subscriptions, sub)
		cfg.LastUsedSub = "default"
		cfg.SubscriptionURL = ""
		cfg.SelectedNode = ""
		cfg.Save()
	}
	return cfg, nil
}

func (c *Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}
	cfg := *c
	cfg.SubscriptionURL = ""
	cfg.SelectedNode = ""
	data, err := json.MarshalIndent(&cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (c *Config) AddSubscription(name, url string) *Subscription {
	sub := &Subscription{
		Name: name,
		URL:  url,
	}
	c.Subscriptions = append(c.Subscriptions, sub)
	return sub
}

func (c *Config) RemoveSubscription(name string) bool {
	for i, s := range c.Subscriptions {
		if s.Name == name {
			c.Subscriptions = append(c.Subscriptions[:i], c.Subscriptions[i+1:]...)
			if c.LastUsedSub == name {
				c.LastUsedSub = ""
			}
			return true
		}
	}
	return false
}

func (c *Config) FindSubscription(name string) *Subscription {
	for _, s := range c.Subscriptions {
		if s.Name == name {
			return s
		}
	}
	return nil
}

func (c *Config) FindFallbackSub(excludeName string) *Subscription {
	if c.LastUsedSub != "" && c.LastUsedSub != excludeName {
		sub := c.FindSubscription(c.LastUsedSub)
		if sub != nil && sub.LastNode != "" && len(sub.Nodes) > 0 {
			return sub
		}
	}
	for _, s := range c.Subscriptions {
		if s.Name == excludeName {
			continue
		}
		if s.LastNode != "" && len(s.Nodes) > 0 {
			return s
		}
	}
	return nil
}

func (s *Subscription) FindNode(name string) *subscription.Node {
	for _, n := range s.Nodes {
		if n.Name == name {
			return n
		}
	}
	return nil
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

func Save(cfg *Config) error {
	return cfg.Save()
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./...`
Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add config/store.go
git commit -m "feat: multi-subscription config with migration and CRUD"
```

---

### Task 3: Add FetchWithProxy to subscription/fetcher.go

**Files:**
- Modify: `subscription/fetcher.go` (add FetchWithProxy function)

- [ ] **Step 1: Add FetchWithProxy**

Add after the existing `Fetch` function in `subscription/fetcher.go`:

```go
func FetchWithProxy(url, socks5ProxyAddr string) ([]byte, error) {
	dialer, err := proxy.SOCKS5("tcp", socks5ProxyAddr, nil, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("socks5 dialer: %w", err)
	}
	ctxDialer, ok := dialer.(proxy.ContextDialer)
	if !ok {
		return nil, fmt.Errorf("socks5 dialer does not support DialContext")
	}
	transport := &http.Transport{
		DialContext:    ctxDialer.DialContext,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Timeout: 30 * time.Second, Transport: transport}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subscription via proxy: %w", err)
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

Add the missing imports at top of file. The full imports block becomes:

```go
import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/net/proxy"
)
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./...`
Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add subscription/fetcher.go
git commit -m "feat: add FetchWithProxy using SOCKS5 proxy dialer"
```

---

### Task 4: Rewrite main.go

**Files:**
- Modify: `main.go` (complete rewrite)

- [ ] **Step 1: Write the new main.go**

```go
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"xray-go/config"
	"xray-go/latency"
	"xray-go/region"
	"xray-go/singbox"
	"xray-go/subscription"
	"xray-go/xrayproxy"
)

type ProxyServer interface {
	Stop() error
}

func main() {
	urlFlag := flag.String("url", "", "add a new subscription URL")
	portFlag := flag.Int("port", 16708, "local proxy port")
	updateFlag := flag.Bool("update", false, "force re-fetch subscription and re-test latency")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if *urlFlag != "" {
		name := promptSubName()
		sub := cfg.AddSubscription(name, *urlFlag)
		cfg.LastUsedSub = name
		cfg.Save()
		nodes, err := fetchSubOrFallback(sub, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching subscription: %v\n", err)
			os.Exit(1)
		}
		sub.Nodes = nodes
		sub.LastFetched = time.Now()
		cfg.Save()
	}

	for {
		sub := selectSubscription(cfg)
		if sub == nil {
			os.Exit(0)
		}
		cfg.LastUsedSub = sub.Name
		cfg.Save()

		nodes := sub.Nodes
		if len(nodes) == 0 || *updateFlag {
			fetchedNodes, err := fetchSubOrFallback(sub, cfg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				config.Save(cfg)
				continue
			}
			nodes = fetchedNodes
			sub.Nodes = nodes
			sub.LastFetched = time.Now()
			config.Save(cfg)
		} else {
			fmt.Printf("Using cached nodes (%d nodes)\n", len(nodes))
		}

		groups := region.GroupByRegion(nodes)
		selectedRegion := region.PromptRegion(groups)

		var targetNodes []*subscription.Node
		if selectedRegion == "" {
			targetNodes = nodes
		} else {
			targetNodes = groups[selectedRegion]
			fmt.Printf("\nSelected region: %s (%d nodes)\n", selectedRegion, len(targetNodes))
		}

		fmt.Println("\nTesting latency...")
		bestNode, bestLatency, err := latency.FindBest(targetNodes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			if len(cfg.Subscriptions) > 0 {
				continue
			}
			os.Exit(1)
		}
		fmt.Printf("Best node: %s (%v)\n", bestNode.Name, bestLatency)
		sub.LastNode = bestNode.Name
		config.Save(cfg)

		httpPort := *portFlag
		socksPort := httpPort + 1
		fmt.Printf("Starting proxy on 127.0.0.1:%d (HTTP) and 127.0.0.1:%d (SOCKS5)...\n", httpPort, socksPort)

		var srv ProxyServer
		if bestNode.Protocol == "anytls" {
			srv, err = singbox.Start(bestNode, socksPort, httpPort)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error starting sing-box proxy: %v\n", err)
				os.Exit(1)
			}
		} else {
			srv, err = xrayproxy.Start(bestNode, socksPort, httpPort)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error starting xray proxy: %v\n", err)
				os.Exit(1)
			}
		}
		fmt.Printf("Proxy running at 127.0.0.1:%d (HTTP) and 127.0.0.1:%d (SOCKS5)\n", httpPort, socksPort)

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		fmt.Println("\nShutting down...")
		srv.Stop()
		fmt.Println("Done.")
		return
	}
}

func selectSubscription(cfg *config.Config) *config.Subscription {
	if len(cfg.Subscriptions) == 0 {
		fmt.Println("No saved subscriptions.")
		return promptAddSub(cfg)
	}

	for {
		fmt.Println("\nSaved subscriptions:")
		for i, s := range cfg.Subscriptions {
			cached := len(s.Nodes)
			marker := " "
			if s.Name == cfg.LastUsedSub {
				marker = "*"
			}
			fmt.Printf("  %2d. %s%s (%s) [%d cached]\n", i+1, marker, s.Name, s.URL, cached)
		}
		fmt.Printf("  %2d. + Add new subscription\n", len(cfg.Subscriptions)+1)
		fmt.Printf("  %2d. - Delete a subscription\n", len(cfg.Subscriptions)+2)
		fmt.Printf("  %2d. Exit\n", len(cfg.Subscriptions)+3)

		fmt.Print("\nSelect option: ")
		var input string
		fmt.Scanln(&input)
		choice := 0
		fmt.Sscanf(input, "%d", &choice)

		if choice >= 1 && choice <= len(cfg.Subscriptions) {
			return cfg.Subscriptions[choice-1]
		}
		if choice == len(cfg.Subscriptions)+1 {
			sub := promptAddSub(cfg)
			if sub != nil {
				return sub
			}
			continue
		}
		if choice == len(cfg.Subscriptions)+2 {
			promptDeleteSub(cfg)
			continue
		}
		if choice == len(cfg.Subscriptions)+3 {
			return nil
		}
		fmt.Println("Invalid choice")
	}
}

func promptSubName() string {
	fmt.Print("Enter subscription name: ")
	var input string
	fmt.Scanln(&input)
	cleaned := strings.TrimSpace(input)
	if cleaned == "" {
		cleaned = fmt.Sprintf("sub-%d", time.Now().Unix())
	}
	return cleaned
}

func promptAddSub(cfg *config.Config) *config.Subscription {
	name := promptSubName()
	fmt.Print("Enter subscription URL: ")
	var url string
	fmt.Scanln(&url)
	if url == "" {
		fmt.Println("URL cannot be empty")
		return nil
	}
	sub := cfg.AddSubscription(name, url)
	cfg.LastUsedSub = name
	cfg.Save()
	fmt.Printf("Fetching subscription '%s'...\n", name)
	nodes, err := fetchSubOrFallback(sub, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching subscription: %v\n", err)
		cfg.Save()
		return nil
	}
	sub.Nodes = nodes
	sub.LastFetched = time.Now()
	cfg.Save()
	fmt.Printf("Got %d nodes from '%s'\n", len(nodes), name)
	return sub
}

func promptDeleteSub(cfg *config.Config) {
	if len(cfg.Subscriptions) == 0 {
		fmt.Println("No subscriptions to delete.")
		return
	}
	fmt.Println("\nSelect subscription to delete:")
	for i, s := range cfg.Subscriptions {
		fmt.Printf("  %2d. %s\n", i+1, s.Name)
	}
	fmt.Print("Select: ")
	var input string
	fmt.Scanln(&input)
	choice := 0
	fmt.Sscanf(input, "%d", &choice)
	if choice < 1 || choice > len(cfg.Subscriptions) {
		fmt.Println("Invalid choice")
		return
	}
	sub := cfg.Subscriptions[choice-1]
	fmt.Printf("Delete '%s'? (y/N): ", sub.Name)
	fmt.Scanln(&input)
	if strings.ToLower(strings.TrimSpace(input)) == "y" {
		cfg.RemoveSubscription(sub.Name)
		cfg.Save()
		fmt.Println("Deleted.")
	}
}

func fetchSubOrFallback(sub *config.Subscription, cfg *config.Config) ([]*subscription.Node, error) {
	data, err := subscription.Fetch(sub.URL)
	if err == nil {
		return subscription.Parse(data)
	}
	fmt.Printf("Direct fetch failed: %v\n", err)
	fmt.Println("Attempting fallback via previous working node...")

	fallbackSub := cfg.FindFallbackSub(sub.Name)
	if fallbackSub == nil {
		return nil, fmt.Errorf("no fallback subscription available")
	}
	fallbackNode := fallbackSub.FindNode(fallbackSub.LastNode)
	if fallbackNode == nil {
		return nil, fmt.Errorf("fallback node not found in cached data")
	}

	socksPort, err := xrayproxy.GetFreePort()
	if err != nil {
		return nil, fmt.Errorf("get free port: %w", err)
	}
	httpPort, err := xrayproxy.GetFreePort()
	if err != nil {
		return nil, fmt.Errorf("get free port: %w", err)
	}

	fmt.Printf("Starting fallback proxy with node '%s'...\n", fallbackNode.Name)
	var srv ProxyServer
	if fallbackNode.Protocol == "anytls" {
		srv, err = singbox.Start(fallbackNode, socksPort, httpPort)
	} else {
		srv, err = xrayproxy.Start(fallbackNode, socksPort, httpPort)
	}
	if err != nil {
		return nil, fmt.Errorf("start fallback proxy: %w", err)
	}
	defer srv.Stop()
	time.Sleep(200 * time.Millisecond)

	proxyAddr := fmt.Sprintf("127.0.0.1:%d", socksPort)
	data, err = subscription.FetchWithProxy(sub.URL, proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("fallback fetch failed: %w", err)
	}
	return subscription.Parse(data)
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./...`
Expected: No errors.

- [ ] **Step 3: Run go vet**

Run: `go vet ./...`
Expected: No issues.

- [ ] **Step 4: Commit**

```bash
git add main.go
git commit -m "feat: multi-subscription interactive flow with fallback proxy"
```

---

### Task 5: Full build and smoke test

**Files:**
- None (verification only)

- [ ] **Step 1: Build the binary**

Run: `go build -o xray-go .`
Expected: Build succeeds.

- [ ] **Step 2: Verify binary exists and runs**

Run: `./xray-go --help`
Expected: Shows flags `-url`, `-port`, `-update`.

- [ ] **Step 3: Commit any final changes**

```bash
git status
# Only commit if there are uncommitted changes
```
