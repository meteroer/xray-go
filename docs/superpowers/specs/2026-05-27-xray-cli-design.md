# xray-cli Design Spec

## Overview

A single-binary CLI tool for Linux that wraps xray-core, fetches subscription nodes, tests HTTP latency for each, and starts a local proxy on port 16708 using the fastest node.

## Requirements

- Single Go binary, no external dependencies (xray-core embedded)
- Linux x86_64
- Subscription format: Base64-encoded list of protocol links
- Supported protocols: vmess, vless, trojan, shadowsocks
- Proxy: SOCKS5 + HTTP mixed proxy on `127.0.0.1:16708`
- Auto-select lowest HTTP latency node
- Save subscription URL for reuse

## Architecture

```
xray-cli (single Go binary)
├── main.go              -- Entry point, prompt for subscription URL
├── subscription/
│   ├── fetcher.go       -- HTTP GET subscription content
│   └── parser.go        -- Base64 decode, parse vmess/vless/trojan/ss links
├── latency/
│   └── tester.go        -- Test HTTP latency through xray-core dial
├── proxy/
│   └── server.go        -- Start xray-core instance with mixed proxy
├── config/
│   └── store.go         -- Save/load config to ~/.xray-cli/config.json
└── xray/
    └── builder.go       -- Build xray-core config from Node struct
```

## Data Structures

### Node

```go
type Node struct {
    Name     string
    Protocol string // vmess, vless, trojan, shadowsocks
    Address  string
    Port     int
    UUID     string
    AlterId  int    // vmess only
    Security string // aes-128-gcm, chacha20-poly1305, auto, none
    Network  string // tcp, ws, grpc
    Host     string // SNI / Host header
    Path     string // ws path, grpc serviceName
    TLS      bool
    Flow     string // vless xtls flow (xtls-rprx-vision, etc.)
}
```

### Config

```go
type Config struct {
    SubscriptionURL string `json:"subscription_url"`
    SelectedNode    string `json:"selected_node"`
}
```

Stored at `~/.xray-cli/config.json`.

## Core Flow

1. **Start** → Read config or prompt for subscription URL
2. **Fetch subscription** → HTTP GET → Base64 decode → parse node list
3. **Latency test** → For each node: build temp xray config → start xray-core on random SOCKS5 port → HTTP request through proxy → record response time
4. **Select** → Pick node with lowest latency
5. **Proxy** → Start xray-core instance with mixed inbound on 127.0.0.1:16708

## Subscription Parsing

- HTTP GET the subscription URL
- Base64 decode the response body
- Split by newline
- For each line, detect protocol prefix and parse:
  - `vmess://` → Base64 decode JSON → extract fields
  - `vless://` → Parse as URI with query params
  - `trojan://` → Parse as URI with query params
  - `ss://` → Parse SIP002 format

## Latency Testing

- Concurrency: max 5 parallel
- Timeout: 5 seconds per node
- Test URL: `http://www.gstatic.com/generate_204`
- Method: For each node, start a temporary xray-core SOCKS5 proxy on a random port, send HTTP request through it, measure round-trip time
- If a node fails (timeout, connection error), mark as unavailable
- Select the node with the lowest successful latency

## Proxy Service

- Inbound: `socks` + `http` mixed protocol on `127.0.0.1:16708`
- Outbound: the selected node's protocol configuration
- Start via xray-core `core.New()` API
- Print confirmation: `Proxy running at 127.0.0.1:16708 (SOCKS5+HTTP)`

## Configuration Management

- Config path: `~/.xray-cli/config.json`
- On first run: prompt for subscription URL → save to config
- On subsequent runs: auto-load saved URL
- CLI flags:
  - `--url <url>`: override saved subscription URL
  - `--update`: force re-fetch subscription and re-test latency
  - `--port <port>`: override proxy port (default 16708)

## Graceful Shutdown

- Catch SIGINT/SIGTERM
- Stop xray-core instance
- Exit cleanly

## Error Handling

- Subscription fetch failure: print error, exit
- No available nodes: print "no available nodes", exit
- All latency tests fail: print "all nodes unreachable", exit
- Port already in use: print error with suggestion, exit

## Dependencies

- `github.com/xtls/xray-core` — embedded xray-core
- Go standard library only (no other third-party deps)
