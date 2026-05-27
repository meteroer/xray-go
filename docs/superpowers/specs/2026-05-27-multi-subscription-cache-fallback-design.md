# Multi-Subscription Caching & Fallback Design Spec

## Overview

Extend xray-go to support multiple saved subscriptions with cached node lists, interactive subscription management (add/delete/select), and a fallback mechanism that uses a previously working node as proxy to fetch new subscriptions that cannot be reached directly.

## Requirements

- Manage multiple subscription URLs with named aliases via interactive menu
- Cache complete node lists per subscription in `~/.xray-go/config.json`
- Flow: select subscription → choose region → latency test → start proxy
- If a new subscription URL is unreachable directly, use a previously successful node as a SOCKS5 proxy to fetch it
- Keep backward compatibility: `-url`, `-port`, `-update` flags continue to work
- Migrate existing single-URL config to the new multi-subscription format automatically

## Architecture

```
xray-go (single Go binary)
├── main.go                    -- Entry point, orchestrates new flow
├── subscription/
│   ├── fetcher.go             -- HTTP GET subscription, add FetchWithProxy
│   └── parser.go              -- (unchanged) Base64 decode, parse protocol links
├── latency/
│   └── tester.go              -- (unchanged) Test HTTP latency via SOCKS5 dial
├── xrayproxy/
│   └── server.go              -- (unchanged) Start/stop xray-core proxy
├── singbox/
│   └── server.go              -- (unchanged) Start/stop sing-box proxy
├── region/
│   └── region.go              -- (unchanged) Detect region, prompt selection
└── config/
    └── store.go               -- Load/Save multi-subscription config, CRUD helpers
```

## Data Model

### Config (revised)

```go
type Subscription struct {
    Name        string              `json:"name"`
    URL         string              `json:"url"`
    Nodes       []*subscription.Node `json:"nodes,omitempty"`
    LastNode    string              `json:"last_node"`
    LastFetched time.Time           `json:"last_fetched"`
}

type Config struct {
    Subscriptions      []*Subscription `json:"subscriptions"`
    LastUsedSub        string          `json:"last_used_subscription"`
    // Legacy fields kept for automatic migration
    SubscriptionURL    string          `json:"subscription_url,omitempty"`
    SelectedNode       string          `json:"selected_node,omitempty"`
}
```

File path: `~/.xray-go/config.json` (unchanged).

### Legacy Migration

On `config.Load()`:
1. If `subscription_url` is set and `subscriptions` is empty, create a single Subscription from the legacy fields
2. Save the migrated config immediately (legacy fields cleared on next save)

## Core Flow

```
1. Load config → auto-migrate legacy fields if present

2. Interactive subscription selection:
   ┌──────────────────────────────────────┐
   │ Saved subscriptions:                 │
   │   1. my-vps (https://...sub1) [15 ❄]  │
   │   2. work-vpn (https://...sub2) [8 ❄] │
   │   3. + Add new subscription           │
   │   4. - Delete a subscription          │
   │   Select option:                      │
   └──────────────────────────────────────┘

   - Choose existing → go to step 3
   - "Add new" → prompt for URL + name → fetch → save → back to step 2
   - "Delete" → show list with delete options → confirm → save → back to step 2

3. Check node cache for selected subscription:
   a. Has cache → use cached nodes (no TTL-based expiry; use `-update` to force refresh)
   b. No cache (new sub or -update flag) → HTTP fetch:
      - Success → parse, cache to config, proceed
      - Failure → fallback flow (step 3c)
   c. Fallback: find a previous subscription with LastNode set → start proxy
      on that node → use SOCKS5 proxy to fetch new subscription → success?
      - Yes → parse, cache, stop temp proxy, proceed
      - No → stop temp proxy, show error, back to step 2

4. Region selection (unchanged):
   - Group nodes by region
   - Prompt user to pick a region or "All regions"

5. Latency test (unchanged):
   - Concurrently test nodes in selected region
   - Pick node with lowest latency

6. Start proxy (unchanged):
   - Start xray-core or sing-box with best node
   - Save selected node as LastNode for the subscription
   - Block until SIGINT/SIGTERM → stop → exit
```

## Fallback Mechanism Detail

When fetching a subscription fails with a direct HTTP request:

```
fetchWithFallback(sub *Subscription, cfg *Config) ([]*Node, error):
    // findFallbackSubscription iterates all subs excluding current:
    //   1. prefer cfg.LastUsedSub if it has LastNode set
    //   2. then any subscription with LastNode set
    fallbackSub := findFallbackSubscription(cfg, sub.Name)
    if fallbackSub == nil || fallbackSub.LastNode == "":
        return nil, errNoFallback

    // Find the actual Node object from the fallback sub's cached nodes
    fallbackNode := findNodeByName(fallbackSub.Nodes, fallbackSub.LastNode)
    if fallbackNode == nil:
        return nil, errNoFallbackNode

    // Allocate free ports for temp proxy
    tempSocks := getFreePort()
    tempHTTP := getFreePort()

    // Start temp proxy with the fallback node
    var srv ProxyServer
    if fallbackNode.Protocol == "anytls":
        srv = singbox.Start(fallbackNode, tempSocks, tempHTTP)
    else:
        srv = xrayproxy.Start(fallbackNode, tempSocks, tempHTTP)
    defer srv.Stop()
    time.Sleep(200ms) // wait for proxy ready

    // Fetch through SOCKS5 proxy
    proxyAddr := fmt.Sprintf("127.0.0.1:%d", tempSocks)
    data, err := subscription.FetchWithProxy(sub.URL, proxyAddr)
    if err != nil:
        return nil, err

    return subscription.Parse(data)
```

### New fetcher method

```go
// subscription/fetcher.go
func FetchWithProxy(url, socks5ProxyAddr string) ([]byte, error)
```

Uses `golang.org/x/net/proxy` SOCKS5 dialer as the HTTP transport dialer. Reuses the same pattern already present in `latency/tester.go`.

## CLI Flags (Preserved)

| Flag | Behavior |
|------|----------|
| `-url <url>` | Add URL as a new subscription (user prompted for name), fetch nodes, then go to subscription selection menu |
| `-update` | Force re-fetch of selected subscription, ignore cached nodes |
| `-port <port>` | Override HTTP proxy port (default 16708) |

## Interactive Menus

### Subscription Selection

Displays all saved subscriptions with node count and last-used indicator. Options: pick one, add new, delete existing, or exit.

### Add Subscription

Prompt for:
1. Name (display alias, required)
2. URL (required)

After entry, immediately fetch and parse. If fetch fails, offer the fallback flow automatically. If there are no other subscriptions with a working node, show the error and return to the menu.

### First Run (No Saved Subscriptions)

When the subscription list is empty, skip the selection menu and go directly to the "Add Subscription" prompt. After adding the first subscription, proceed normally through the flow.

### Using -url Flag

Adds the URL as a new subscription (user is prompted for a name), fetches the nodes immediately, and then proceeds to the subscription selection menu showing the newly added subscription. This way users can add subs non-interactively but still choose which sub/region to use.

### Delete Subscription

List all subscriptions. User picks by number. Confirm before delete. If the deleted subscription was `LastUsedSub`, clear it.

## Testing Strategy

- Unit tests for new config functions (Load/Save with multi-sub, migration, CRUD)
- Unit test for `FetchWithProxy` (mock HTTP server behind SOCKS5 proxy)
- Integration test for the fallback flow (mock unreachable URL + working proxy)
- Manually verify backward compatibility with existing `config.json` files

## Error Handling

- Config file corruption: log warning, start with empty config
- Migration failure: log warning, continue with empty subscriptions
- All nodes in fallback unreachable: return error, user retries from subscription menu
- Port conflict on startup: detect and prompt user

## Dependencies

- `golang.org/x/net/proxy` — already in go.sum (used by latency tester)
- No new third-party dependencies
