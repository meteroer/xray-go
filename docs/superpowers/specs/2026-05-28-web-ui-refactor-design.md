# Web UI Refactor Design: Vue 3 + Element Plus

## Overview

Refactor the xray-go web UI from a vanilla JS single-file SPA to a modern Vue 3 + Element Plus application, referencing v2rayA's ngui architecture. The web module and CLI module remain completely independent.

## Architecture

### Frontend

- **Framework**: Vue 3 + Vite + Element Plus + Vue Router + Pinia
- **Source**: `web/frontend/`
- **Build output**: `web/static/` (embedded by Go via `go:embed`)
- **i18n**: vue-i18n with zh/en
- **Real-time**: WebSocket for proxy status and latency results

### Backend

- **HTTP server**: Go net/http with chi router (already in go.mod)
- **WebSocket**: gorilla/websocket (already indirect dep)
- **API**: RESTful JSON API, same auth (JWT + bcrypt)
- **Static files**: `go:embed web/static`

### Build Dependencies

All build toolchain installed under dedicated `/mnt` subdirectories:

| Tool | Path | Purpose |
|------|------|---------|
| Node.js | `/mnt/nodejs/` | JS runtime for Vite build |
| npm cache | `/mnt/npm-cache/` | Package cache directory |

### Modularity

- `web` subcommand and CLI mode are independent
- `web` package imports only core packages: `config`, `subscription`, `latency`, `region`, `xrayproxy`, `singbox`
- CLI code in `main.go` never imports `web`
- Both modes can work independently

## Page Structure

Top navigation bar + multi-page routing (referencing v2rayA ngui layout):

| Page | Route | Description |
|------|-------|-------------|
| Login/Register | `/login` | First-time user creation / login |
| Nodes | `/` (home) | Server list, subscription groups, connect/disconnect, latency test |
| Subscriptions | `/subscription` | Add/delete/refresh subscriptions |
| Routing | `/routing` | Route mode, whitelist/blacklist editing |
| Settings | `/settings` | Port config, language, password change, logout |

## Top Navigation Bar

- Left: Logo + proxy status indicator (green dot = running, red dot = stopped)
- Center: Nav links — Nodes / Subscriptions / Routing / Settings
- Right: Language toggle (中/EN) / User dropdown (logout)

## WebSocket Real-time Communication

### Endpoint

- Path: `/api/ws?token=<jwt>`
- Auth: JWT token passed as query parameter during handshake
- Protocol: JSON text frames

### Message Types (Server → Client)

| Type | Payload | Description |
|------|---------|-------------|
| `proxy_status` | `{running, node, http_port, socks_port, route_mode}` | Proxy started/stopped |
| `latency_progress` | `{name, latency_ms}` | Single node latency result (streaming) |
| `latency_done` | `{results: [{name, latency_ms, error?}]}` | All latency tests complete |
| `error` | `{message}` | Server error notification |

### Client → Server Messages

| Type | Payload | Description |
|------|---------|-------------|
| `ping` | — | Keepalive |

### Reconnection

- Auto-reconnect with exponential backoff (1s, 2s, 4s, max 30s)
- On reconnect, server sends current `proxy_status`

## Backend API Changes

### New Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/ws` | WebSocket upgrade |
| POST | `/api/nodes/{name}/latency` | Test single node latency |
| DELETE | `/api/nodes/{name}` | Delete a standalone node |
| PUT | `/api/settings/route-mode` | Save route mode independently |
| GET | `/api/settings/whitelist` | Get whitelist |
| PUT | `/api/settings/whitelist` | Update whitelist |
| GET | `/api/settings/blacklist` | Get blacklist |
| PUT | `/api/settings/blacklist` | Update blacklist |

### Modified Endpoints

| Method | Path | Change |
|--------|------|--------|
| POST | `/api/proxy/start` | Add optional `http_port` / `socks_port` fields |

### Existing Endpoints (unchanged)

All existing auth, config, subscription, node, and proxy endpoints remain compatible.

## Backend Code Organization

Current `web/handler.go` (584 lines) splits into focused files:

```
web/
├── server.go              # Server struct, NewServer, Start, Stop
├── router.go              # Route registration, SPA handler, middleware
├── auth.go                # AuthManager (unchanged)
├── handler_config.go      # GET /api/config
├── handler_auth.go        # /api/auth/* handlers + authMiddleware
├── handler_subscription.go # /api/subscriptions/* handlers
├── handler_node.go        # /api/nodes/* handlers
├── handler_proxy.go       # /api/proxy/* handlers
├── handler_settings.go    # /api/settings/* handlers
├── handler_ws.go          # WebSocket handler
├── ws_hub.go              # WebSocket hub (connection management, broadcast)
├── static/                # Built frontend (go:embed)
│   ├── index.html
│   ├── assets/
│   │   ├── index-*.js
│   │   └── index-*.css
│   └── ...
```

## Frontend Component Structure

```
web/frontend/
├── package.json
├── vite.config.ts
├── tsconfig.json
├── index.html
├── src/
│   ├── App.vue                    # Top nav layout + RouterView
│   ├── main.ts                    # Entry point
│   ├── router/index.ts            # Route definitions + auth guard
│   ├── stores/
│   │   ├── auth.ts                # Auth state (token, user)
│   │   ├── proxy.ts               # Proxy state (WS-driven)
│   │   ├── subscription.ts        # Subscription data
│   │   └── settings.ts            # Settings data
│   ├── composables/
│   │   ├── useWebSocket.ts        # WS connection, reconnect, message dispatch
│   │   └── useApi.ts              # Fetch wrapper with auth header + error handling
│   ├── views/
│   │   ├── LoginView.vue          # Login / register form
│   │   ├── NodesView.vue          # Main page: node list + proxy control
│   │   ├── SubscriptionView.vue   # Subscription CRUD
│   │   ├── RoutingView.vue        # Route mode + rules editor
│   │   └── SettingsView.vue       # Misc settings
│   ├── components/
│   │   ├── TopNavBar.vue          # Navigation bar
│   │   ├── ProxyControl.vue       # Start/stop proxy buttons + status
│   │   ├── NodeTable.vue          # Node list table with region filter
│   │   ├── NodeRow.vue            # Single node row
│   │   ├── SubscriptionCard.vue   # Subscription expandable card
│   │   ├── LatencyBadge.vue       # Latency display (color-coded)
│   │   ├── AddNodeDialog.vue      # Dialog: add node by link
│   │   ├── AddSubscriptionDialog.vue # Dialog: add subscription
│   │   └── RouteRuleEditor.vue    # Whitelist/blacklist tag editor
│   ├── i18n/
│   │   ├── index.ts               # vue-i18n setup
│   │   ├── zh.ts
│   │   └── en.ts
│   └── styles/
│       └── variables.scss         # Element Plus theme overrides
```

## Key UI Features (referencing v2rayA)

### Nodes Page (Home)

- **Proxy control bar** at top: status badge + start/stop buttons + current node info
- **Region filter**: dropdown to filter nodes by region
- **Subscription groups**: expandable sections, each showing its nodes
- **Standalone nodes**: separate expandable section
- **Node table columns**: Name | Address:Port | Protocol | Latency | Actions (Connect/Delete)
- **Latency test**: "Test All" button, results stream in via WebSocket, each row updates live
- **Connect**: click to start proxy with that node
- **Batch latency**: progress bar during test

### Subscription Page

- Table: Name | URL | Nodes Count | Last Fetched | Actions
- Add dialog: name + URL inputs
- Refresh: triggers fetch, shows loading state
- Delete: confirmation dialog

### Routing Page

- Route mode selector: Global / Whitelist / Blacklist (radio group)
- Whitelist editor: tag-style input (add/remove geosite/geoip rules)
- Blacklist editor: same tag-style input
- Save button

### Settings Page

- HTTP port / SOCKS5 port configuration
- Language toggle
- Change password
- Logout button

## Development Workflow

### Frontend Development

```bash
# Terminal 1: Start Go backend
go run . web

# Terminal 2: Start Vite dev server
cd web/frontend
npm run dev
# Vite dev server on :5173, proxies /api/* to :18700
```

### Production Build

```bash
cd web/frontend
npm run build
# Output to web/static/

# Build Go binary (embeds web/static/)
go build -o xray-go .
```

### Vite Dev Proxy Config

```ts
// vite.config.ts
server: {
  proxy: {
    '/api': 'http://localhost:18700'
  }
}
```

## Error Handling

- API errors: Element Plus `ElMessage.error()` with i18n message
- WebSocket disconnect: show warning banner, auto-reconnect with status indicator
- Network errors: retry with backoff for API calls

## Security

- JWT auth on all protected endpoints and WebSocket
- XSS: Vue template auto-escaping (no v-html with user data)
- CSP headers on static file responses
- No secrets in frontend code

## Migration Strategy

1. Create `web/frontend/` with Vue 3 + Vite scaffold
2. Implement backend handler file split and WebSocket
3. Build frontend pages one by one (Login → Nodes → Subscriptions → Routing → Settings)
4. Test each page against backend API
5. Build production frontend, verify `go:embed` works
6. Remove old `web/static/index.html`, `app.js`, `style.css`
7. Verify CLI mode unaffected
