# Web UI Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor xray-go web UI from vanilla JS SPA to Vue 3 + Element Plus, with WebSocket real-time updates and modular backend.

**Architecture:** Vue 3 + Vite + Element Plus frontend in `web/frontend/`, building to `web/static/` embedded by Go. Backend splits monolithic `handler.go` into focused files, adds WebSocket hub. CLI and web modes remain independent.

**Tech Stack:** Vue 3, Vite, Element Plus, Vue Router, Pinia, vue-i18n, gorilla/websocket, go-chi/chi

---

## Task 1: Install Node.js build toolchain to /mnt

**Files:**
- None (system setup only)

- [ ] **Step 1: Download and install Node.js to /mnt/nodejs/**

```bash
mkdir -p /mnt/nodejs
cd /tmp
curl -fsSL https://nodejs.org/dist/v22.16.0/node-v22.16.0-linux-x64.tar.xz -o node.tar.xz
tar -xf node.tar.xz
cp -r node-v22.16.0-linux-x64/* /mnt/nodejs/
rm -rf node-v22.16.0-linux-x64 node.tar.xz
```

- [ ] **Step 2: Verify Node.js works**

Run: `/mnt/nodejs/bin/node --version`
Expected: `v22.16.0`

- [ ] **Step 3: Configure npm cache to /mnt/npm-cache**

```bash
mkdir -p /mnt/npm-cache
/mnt/nodejs/bin/npm config set cache /mnt/npm-cache --global
```

---

## Task 2: Scaffold Vue 3 + Vite frontend project

**Files:**
- Create: `web/frontend/package.json`
- Create: `web/frontend/vite.config.ts`
- Create: `web/frontend/tsconfig.json`
- Create: `web/frontend/index.html`
- Create: `web/frontend/src/main.ts`
- Create: `web/frontend/src/App.vue`
- Create: `web/frontend/src/env.d.ts`

- [ ] **Step 1: Create frontend directory and initialize npm project**

```bash
mkdir -p /mnt/software/xray-go/web/frontend/src
```

- [ ] **Step 2: Write package.json**

File: `web/frontend/package.json`
```json
{
  "name": "xray-go-web",
  "private": true,
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vue-tsc --noEmit && vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "vue": "^3.5.13",
    "vue-router": "^4.5.0",
    "pinia": "^3.0.2",
    "element-plus": "^2.9.7",
    "vue-i18n": "^11.1.3",
    "@element-plus/icons-vue": "^2.3.1"
  },
  "devDependencies": {
    "@vitejs/plugin-vue": "^5.2.3",
    "typescript": "^5.8.3",
    "vite": "^6.3.5",
    "vue-tsc": "^2.2.8",
    "sass": "^1.87.0"
  }
}
```

- [ ] **Step 3: Install dependencies**

```bash
cd /mnt/software/xray-go/web/frontend
/mnt/nodejs/bin/npm install
```

- [ ] **Step 4: Write vite.config.ts**

File: `web/frontend/vite.config.ts`
```ts
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:18700',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: '../../web/static',
    emptyOutDir: true,
  },
})
```

- [ ] **Step 5: Write tsconfig.json**

File: `web/frontend/tsconfig.json`
```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "module": "ESNext",
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "isolatedModules": true,
    "moduleDetection": "force",
    "noEmit": true,
    "jsx": "preserve",
    "strict": true,
    "noUnusedLocals": false,
    "noUnusedParameters": false,
    "noFallthroughCasesInSwitch": true,
    "paths": {
      "@/*": ["./src/*"]
    },
    "baseUrl": "."
  },
  "include": ["src/**/*.ts", "src/**/*.tsx", "src/**/*.vue", "src/env.d.ts"]
}
```

- [ ] **Step 6: Write index.html**

File: `web/frontend/index.html`
```html
<!DOCTYPE html>
<html lang="zh">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>xray-go</title>
</head>
<body>
  <div id="app"></div>
  <script type="module" src="/src/main.ts"></script>
</body>
</html>
```

- [ ] **Step 7: Write src/env.d.ts**

File: `web/frontend/src/env.d.ts`
```ts
/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}
```

- [ ] **Step 8: Write src/main.ts**

File: `web/frontend/src/main.ts`
```ts
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import App from './App.vue'
import { router } from './router'
import { i18n } from './i18n'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(i18n)
app.use(ElementPlus)
app.mount('#app')
```

- [ ] **Step 9: Write src/App.vue**

File: `web/frontend/src/App.vue`
```vue
<template>
  <router-view />
</template>
```

- [ ] **Step 10: Create minimal router stub so build succeeds**

```bash
mkdir -p /mnt/software/xray-go/web/frontend/src/router
mkdir -p /mnt/software/xray-go/web/frontend/src/i18n
```

File: `web/frontend/src/router/index.ts`
```ts
import { createRouter, createWebHistory } from 'vue-router'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
    },
    {
      path: '/',
      name: 'nodes',
      component: () => import('@/views/NodesView.vue'),
    },
  ],
})

router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('token')
  if (to.name !== 'login' && !token) {
    next({ name: 'login' })
  } else {
    next()
  }
})
```

- [ ] **Step 11: Create minimal i18n stub**

File: `web/frontend/src/i18n/index.ts`
```ts
import { createI18n } from 'vue-i18n'
import zh from './zh'
import en from './en'

const savedLang = localStorage.getItem('lang') || 'zh'

export const i18n = createI18n({
  legacy: false,
  locale: savedLang,
  fallbackLocale: 'en',
  messages: { zh, en },
})
```

File: `web/frontend/src/i18n/zh.ts`
```ts
export default {
  app: {
    title: 'xray-go',
  },
  nav: {
    nodes: '节点',
    subscription: '订阅',
    routing: '路由',
    settings: '设置',
  },
  auth: {
    login: '登录',
    register: '创建用户',
    username: '用户名',
    password: '密码',
    confirmPassword: '确认密码',
    submit: '提交',
    passwordMismatch: '两次密码不一致',
  },
  proxy: {
    status: '代理状态',
    running: '运行中',
    stopped: '已停止',
    start: '启动代理',
    stop: '停止代理',
    currentNode: '当前节点',
    httpPort: 'HTTP 端口',
    socksPort: 'SOCKS5 端口',
    routeMode: '路由模式',
  },
  node: {
    title: '节点',
    name: '名称',
    address: '地址',
    protocol: '协议',
    latency: '延迟',
    connect: '连接',
    delete: '删除',
    add: '添加节点',
    link: '节点链接',
    testLatency: '测速',
    region: '地区',
    allRegions: '全部地区',
    noData: '暂无数据',
  },
  sub: {
    title: '订阅管理',
    name: '名称',
    url: '地址',
    nodes: '节点数',
    lastFetched: '最后更新',
    add: '添加订阅',
    refresh: '刷新',
    delete: '删除',
    refreshSuccess: '刷新成功',
  },
  routing: {
    title: '路由设置',
    mode: '路由模式',
    global: '全局',
    whitelist: '白名单',
    blacklist: '黑名单',
    whitelistLabel: '白名单规则（仅这些走代理）',
    blacklistLabel: '黑名单规则（这些直连，其他走代理）',
    addRule: '添加规则',
    save: '保存',
  },
  settings: {
    title: '设置',
    language: '语言',
    httpPort: 'HTTP 端口',
    socksPort: 'SOCKS5 端口',
    changePassword: '修改密码',
    logout: '退出登录',
    save: '保存',
  },
  common: {
    confirm: '确认',
    cancel: '取消',
    delete: '删除',
    error: '错误',
    success: '成功',
  },
}
```

File: `web/frontend/src/i18n/en.ts`
```ts
export default {
  app: {
    title: 'xray-go',
  },
  nav: {
    nodes: 'Nodes',
    subscription: 'Subscriptions',
    routing: 'Routing',
    settings: 'Settings',
  },
  auth: {
    login: 'Login',
    register: 'Register',
    username: 'Username',
    password: 'Password',
    confirmPassword: 'Confirm Password',
    submit: 'Submit',
    passwordMismatch: 'Passwords do not match',
  },
  proxy: {
    status: 'Proxy Status',
    running: 'Running',
    stopped: 'Stopped',
    start: 'Start Proxy',
    stop: 'Stop Proxy',
    currentNode: 'Current Node',
    httpPort: 'HTTP Port',
    socksPort: 'SOCKS5 Port',
    routeMode: 'Route Mode',
  },
  node: {
    title: 'Nodes',
    name: 'Name',
    address: 'Address',
    protocol: 'Protocol',
    latency: 'Latency',
    connect: 'Connect',
    delete: 'Delete',
    add: 'Add Node',
    link: 'Node Link',
    testLatency: 'Test Latency',
    region: 'Region',
    allRegions: 'All Regions',
    noData: 'No data',
  },
  sub: {
    title: 'Subscriptions',
    name: 'Name',
    url: 'URL',
    nodes: 'Nodes',
    lastFetched: 'Last Fetched',
    add: 'Add Subscription',
    refresh: 'Refresh',
    delete: 'Delete',
    refreshSuccess: 'Refresh succeeded',
  },
  routing: {
    title: 'Routing',
    mode: 'Route Mode',
    global: 'Global',
    whitelist: 'Whitelist',
    blacklist: 'Blacklist',
    whitelistLabel: 'Whitelist rules (proxy only these)',
    blacklistLabel: 'Blacklist rules (direct these, proxy others)',
    addRule: 'Add Rule',
    save: 'Save',
  },
  settings: {
    title: 'Settings',
    language: 'Language',
    httpPort: 'HTTP Port',
    socksPort: 'SOCKS5 Port',
    changePassword: 'Change Password',
    logout: 'Logout',
    save: 'Save',
  },
  common: {
    confirm: 'Confirm',
    cancel: 'Cancel',
    delete: 'Delete',
    error: 'Error',
    success: 'Success',
  },
}
```

- [ ] **Step 12: Create minimal view stubs so build succeeds**

```bash
mkdir -p /mnt/software/xray-go/web/frontend/src/views
```

File: `web/frontend/src/views/LoginView.vue`
```vue
<template>
  <div class="login-page">
    <el-card class="login-card">
      <template #header>
        <h2>{{ isRegister ? t('auth.register') : t('auth.login') }}</h2>
      </template>
      <el-form @submit.prevent="handleSubmit">
        <el-form-item :label="t('auth.username')">
          <el-input v-model="form.username" />
        </el-form-item>
        <el-form-item :label="t('auth.password')">
          <el-input v-model="form.password" type="password" show-password />
        </el-form-item>
        <el-form-item v-if="isRegister" :label="t('auth.confirmPassword')">
          <el-input v-model="form.confirmPassword" type="password" show-password />
        </el-form-item>
        <el-button type="primary" native-type="submit" :loading="loading" style="width:100%">
          {{ t('auth.submit') }}
        </el-button>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { useApi } from '@/composables/useApi'

const { t } = useI18n()
const router = useRouter()
const api = useApi()

const isRegister = ref(false)
const loading = ref(false)
const form = ref({ username: '', password: '', confirmPassword: '' })

onMounted(async () => {
  try {
    const res = await api.get('/api/auth/status')
    isRegister.value = !res.initialized
  } catch {
    isRegister.value = true
  }
})

async function handleSubmit() {
  if (isRegister.value && form.value.password !== form.value.confirmPassword) {
    ElMessage.error(t('auth.passwordMismatch'))
    return
  }
  loading.value = true
  try {
    const endpoint = isRegister.value ? '/api/auth/init' : '/api/auth/login'
    const data = await api.post(endpoint, {
      username: form.value.username,
      password: form.value.password,
    })
    localStorage.setItem('token', data.token)
    router.push('/')
  } catch (e: any) {
    ElMessage.error(e.message || t('common.error'))
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background: #2c3e50;
}
.login-card {
  width: 400px;
}
.login-card h2 {
  text-align: center;
  margin: 0;
}
</style>
```

File: `web/frontend/src/views/NodesView.vue`
```vue
<template>
  <div>
    <p>Nodes page placeholder</p>
  </div>
</template>
```

- [ ] **Step 13: Create useApi composable stub**

```bash
mkdir -p /mnt/software/xray-go/web/frontend/src/composables
```

File: `web/frontend/src/composables/useApi.ts`
```ts
export function useApi() {
  function getToken(): string {
    return localStorage.getItem('token') || ''
  }

  function headers(): Record<string, string> {
    const h: Record<string, string> = { 'Content-Type': 'application/json' }
    const token = getToken()
    if (token) h['Authorization'] = `Bearer ${token}`
    return h
  }

  async function request(method: string, url: string, body?: any): Promise<any> {
    const opts: RequestInit = { method, headers: headers() }
    if (body !== undefined) opts.body = JSON.stringify(body)
    const res = await fetch(url, opts)
    const data = await res.json()
    if (!res.ok) {
      if (res.status === 401) {
        localStorage.removeItem('token')
        window.location.href = '/login'
      }
      throw new Error(data.error || 'Request failed')
    }
    return data
  }

  return {
    get: (url: string) => request('GET', url),
    post: (url: string, body?: any) => request('POST', url, body),
    put: (url: string, body?: any) => request('PUT', url, body),
    del: (url: string) => request('DELETE', url),
  }
}
```

- [ ] **Step 14: Verify frontend builds**

```bash
cd /mnt/software/xray-go/web/frontend
/mnt/nodejs/bin/npx vue-tsc --noEmit 2>&1 || true
/mnt/nodejs/bin/npx vite build
```

Expected: Build succeeds, output in `web/static/`

- [ ] **Step 15: Commit scaffold**

```bash
cd /mnt/software/xray-go
git add web/frontend/
git commit -m "feat(web): scaffold Vue 3 + Vite + Element Plus frontend"
```

---

## Task 3: Split backend handler.go into focused files

**Files:**
- Create: `web/router.go` (extracted from handler.go)
- Create: `web/handler_auth.go` (extracted from handler.go)
- Create: `web/handler_config.go` (extracted from handler.go)
- Create: `web/handler_subscription.go` (extracted from handler.go)
- Create: `web/handler_node.go` (extracted from handler.go)
- Create: `web/handler_proxy.go` (extracted from handler.go)
- Modify: `web/handler.go` (delete all moved code, keep shared helpers only)

- [ ] **Step 1: Create web/router.go with route registration and SPA handler**

File: `web/router.go`
```go
package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.Handle("/", s.spaHandler())

	mux.HandleFunc("/api/auth/init", s.handleAuthInit)
	mux.HandleFunc("/api/auth/login", s.handleAuthLogin)
	mux.HandleFunc("/api/auth/status", s.handleAuthStatus)
	mux.HandleFunc("/api/auth/logout", s.handleAuthLogout)

	mux.HandleFunc("/api/config", s.authMiddleware(s.handleConfig))

	mux.HandleFunc("/api/subscriptions", s.authMiddleware(s.handleSubscriptions))
	mux.HandleFunc("/api/subscriptions/", s.authMiddleware(s.handleSubscriptionDetail))

	mux.HandleFunc("/api/nodes", s.authMiddleware(s.handleNodes))
	mux.HandleFunc("/api/nodes/regions", s.authMiddleware(s.handleNodeRegions))

	mux.HandleFunc("/api/proxy/start", s.authMiddleware(s.handleProxyStart))
	mux.HandleFunc("/api/proxy/stop", s.authMiddleware(s.handleProxyStop))
	mux.HandleFunc("/api/proxy/status", s.authMiddleware(s.handleProxyStatus))
	mux.HandleFunc("/api/proxy/test", s.authMiddleware(s.handleProxyTest))
}

//go:embed static/*
var staticFS embed.FS

var staticSubFS = mustSub(staticFS, "static")

func mustSub(fsys embed.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(err)
	}
	return sub
}

func (s *Server) spaHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		filePath := strings.TrimPrefix(r.URL.Path, "/")
		filePath = strings.TrimPrefix(filePath, "static/")
		if filePath == "" {
			filePath = "index.html"
		}

		data, err := fs.ReadFile(staticSubFS, filePath)
		if err != nil {
			data, err = fs.ReadFile(staticSubFS, "index.html")
			if err != nil {
				http.NotFound(w, r)
				return
			}
			filePath = "index.html"
		}

		contentType := "text/html; charset=utf-8"
		if strings.HasSuffix(filePath, ".css") {
			contentType = "text/css; charset=utf-8"
		} else if strings.HasSuffix(filePath, ".js") {
			contentType = "application/javascript; charset=utf-8"
		} else if strings.HasSuffix(filePath, ".svg") {
			contentType = "image/svg+xml"
		} else if strings.HasSuffix(filePath, ".png") {
			contentType = "image/png"
		} else if strings.HasSuffix(filePath, ".ico") {
			contentType = "image/x-icon"
		} else if strings.HasSuffix(filePath, ".woff2") {
			contentType = "font/woff2"
		} else if strings.HasSuffix(filePath, ".woff") {
			contentType = "font/woff"
		}
		w.Header().Set("Content-Type", contentType)
		w.Write(data)
	})
}

func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := s.auth.extractToken(r)
		if token == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		username, err := s.auth.ValidateToken(token)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			return
		}
		r.Header.Set("X-Username", username)
		next(w, r)
	}
}
```

- [ ] **Step 2: Create web/handler_auth.go**

File: `web/handler_auth.go`
```go
package web

import (
	"net/http"
)

func (s *Server) handleAuthInit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	if s.auth.HasUser() {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "user already exists"})
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if err := s.auth.CreateUser(req.Username, req.Password); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	token, err := s.auth.ValidateUser(req.Username, req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token":    token,
		"username": req.Username,
	})
}

func (s *Server) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	token, err := s.auth.ValidateUser(req.Username, req.Password)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token":    token,
		"username": req.Username,
	})
}

func (s *Server) handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{
		"initialized": s.auth.HasUser(),
	})
}

func (s *Server) handleAuthLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "logged out",
	})
}
```

- [ ] **Step 3: Create web/handler_config.go**

File: `web/handler_config.go`
```go
package web

import (
	"net/http"
)

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, s.cfg)
}
```

- [ ] **Step 4: Create web/handler_subscription.go**

File: `web/handler_subscription.go`
```go
package web

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"xray-go/config"
	"xray-go/subscription"
)

func (s *Server) handleSubscriptions(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, s.cfg.Subscriptions)
		return
	}
	if r.Method == http.MethodPost {
		var req struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		}
		if err := readJSON(r, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.URL) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and url required"})
			return
		}
		sub := s.cfg.AddSubscription(req.Name, req.URL)
		if err := s.cfg.Save(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, sub)
		return
	}
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func (s *Server) handleSubscriptionDetail(w http.ResponseWriter, r *http.Request) {
	rawPath := r.URL.RawPath
	if rawPath == "" {
		rawPath = r.URL.Path
	}
	rawPath = strings.TrimPrefix(rawPath, "/api/subscriptions/")
	parts := strings.Split(rawPath, "/")
	rawName := parts[0]
	if r.URL.RawQuery != "" {
		rawName = rawName + "?" + r.URL.RawQuery
	}
	name, err := url.PathUnescape(rawName)
	if err != nil {
		name = rawName
	}

	if r.Method == http.MethodDelete {
		if !s.cfg.RemoveSubscription(name) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "subscription not found"})
			return
		}
		if err := s.cfg.Save(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
		return
	}

	if r.Method == http.MethodPost && len(parts) > 1 && parts[1] == "refresh" {
		sub := s.cfg.FindSubscription(name)
		if sub == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "subscription not found"})
			return
		}
		data, err := subscription.Fetch(sub.URL)
		if err != nil {
			log.Printf("Direct fetch failed for '%s': %v, trying fallback...", sub.Name, err)
			data, err = s.fetchWithFallback(sub)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
		}
		nodes, err := subscription.Parse(data)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		sub.Nodes = nodes
		sub.LastFetched = time.Now()
		if err := s.cfg.Save(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, sub)
		return
	}

	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func (s *Server) fetchWithFallback(sub *config.Subscription) ([]byte, error) {
	var fallbackNode *subscription.Node
	var fallbackSub *config.Subscription

	fallbackSub = s.cfg.FindFallbackSub(sub.Name)
	if fallbackSub != nil {
		fallbackNode = fallbackSub.FindNode(fallbackSub.LastNode)
	}
	if fallbackNode == nil {
		for _, candidate := range s.cfg.Subscriptions {
			if candidate.Name == sub.Name {
				continue
			}
			if len(candidate.Nodes) > 0 {
				fallbackNode = candidate.Nodes[0]
				fallbackSub = candidate
				break
			}
		}
	}
	if fallbackNode == nil {
		return nil, fmt.Errorf("no fallback node available")
	}

	socksPort, err := xrayproxy.GetFreePort()
	if err != nil {
		return nil, fmt.Errorf("get free port: %w", err)
	}
	httpPort, err := xrayproxy.GetFreePort()
	if err != nil {
		return nil, fmt.Errorf("get free port: %w", err)
	}

	log.Printf("Starting fallback proxy with node '%s' on ports socks=%d http=%d", fallbackNode.Name, socksPort, httpPort)
	var proxyServer ProxyServer
	if fallbackNode.Protocol == "anytls" {
		proxyServer, err = singbox.Start(fallbackNode, socksPort, httpPort, config.RouteModeGlobal, nil, nil)
	} else {
		proxyServer, err = xrayproxy.Start(fallbackNode, socksPort, httpPort, config.RouteModeGlobal, nil, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("start fallback proxy: %w", err)
	}
	defer proxyServer.Stop()
	time.Sleep(300 * time.Millisecond)

	proxyAddr := fmt.Sprintf("0.0.0.0:%d", socksPort)
	data, err := subscription.FetchWithProxy(sub.URL, proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("fallback fetch failed: %w", err)
	}
	log.Printf("Fallback fetch succeeded for '%s'", sub.Name)
	return data, nil
}
```

- [ ] **Step 5: Create web/handler_node.go**

File: `web/handler_node.go`
```go
package web

import (
	"net/http"

	"xray-go/region"
	"xray-go/subscription"
)

func (s *Server) getAllNodes() []*subscription.Node {
	var nodes []*subscription.Node
	for _, sub := range s.cfg.Subscriptions {
		nodes = append(nodes, sub.Nodes...)
	}
	nodes = append(nodes, s.cfg.StandaloneNodes...)
	return nodes
}

func (s *Server) handleNodes(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, s.getAllNodes())
		return
	}
	if r.Method == http.MethodPost {
		var req struct {
			Link string `json:"link"`
		}
		if err := readJSON(r, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		node, err := subscription.ParseNode(req.Link)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		s.cfg.AddStandaloneNode(node)
		if err := s.cfg.Save(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, node)
		return
	}
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func (s *Server) handleNodeRegions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, region.GroupByRegion(s.getAllNodes()))
}
```

- [ ] **Step 6: Create web/handler_proxy.go**

File: `web/handler_proxy.go`
```go
package web

import (
	"errors"
	"io"
	"log"
	"net/http"

	"xray-go/config"
	"xray-go/latency"
	"xray-go/region"
	"xray-go/singbox"
	"xray-go/subscription"
	"xray-go/xrayproxy"
)

func (s *Server) handleProxyStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "proxy already running"})
		return
	}

	var req struct {
		NodeName  string           `json:"node_name,omitempty"`
		Region    string           `json:"region,omitempty"`
		RouteMode config.RouteMode `json:"route_mode,omitempty"`
	}
	if err := readJSON(r, &req); err != nil {
		if !errors.Is(err, io.EOF) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
	}

	if req.RouteMode != "" {
		switch req.RouteMode {
		case config.RouteModeGlobal, config.RouteModeWhitelist, config.RouteModeBlacklist:
			s.cfg.RouteMode = req.RouteMode
		default:
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid route_mode"})
			return
		}
	}

	var node *subscription.Node
	allNodes := s.getAllNodes()

	if req.NodeName != "" {
		for _, n := range allNodes {
			if n.Name == req.NodeName {
				node = n
				break
			}
		}
		if node == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "node not found"})
			return
		}
	} else {
		var targetNodes []*subscription.Node
		if req.Region != "" {
			groups := region.GroupByRegion(allNodes)
			targetNodes = groups[req.Region]
		} else {
			targetNodes = allNodes
		}
		if len(targetNodes) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no nodes available"})
			return
		}
		var err error
		node, _, err = latency.FindBest(targetNodes)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}

	var proxy ProxyServer
	var err error
	httpPort := 16708
	socksPort := 16709

	if node.Protocol == "anytls" {
		proxy, err = singbox.Start(node, socksPort, httpPort, s.cfg.RouteMode, s.cfg.Whitelist, s.cfg.Blacklist)
	} else {
		proxy, err = xrayproxy.Start(node, socksPort, httpPort, s.cfg.RouteMode, s.cfg.Whitelist, s.cfg.Blacklist)
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	s.proxy = proxy
	s.currentNode = node
	s.isRunning = true
	s.httpPort = httpPort
	s.socksPort = socksPort

	if err := s.cfg.Save(); err != nil {
		log.Printf("failed to save config: %v", err)
	}

	s.hub.Broadcast(map[string]interface{}{
		"type": "proxy_status",
		"running": true,
		"node": node,
		"http_port": httpPort,
		"socks_port": socksPort,
		"route_mode": s.cfg.RouteMode,
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "proxy started",
		"node":       node,
		"http_port":  httpPort,
		"socks_port": socksPort,
	})
}

func (s *Server) handleProxyStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.proxy != nil {
		if err := s.proxy.Stop(); err != nil {
			log.Printf("proxy stop error: %v", err)
		}
	}
	s.proxy = nil
	s.currentNode = nil
	s.isRunning = false
	s.httpPort = 0
	s.socksPort = 0

	s.hub.Broadcast(map[string]interface{}{
		"type": "proxy_status",
		"running": false,
		"node": nil,
		"http_port": 0,
		"socks_port": 0,
		"route_mode": s.cfg.RouteMode,
	})

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "proxy stopped",
	})
}

func (s *Server) handleProxyStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"running":    s.isRunning,
		"http_port":  s.httpPort,
		"socks_port": s.socksPort,
		"route_mode": s.cfg.RouteMode,
		"node":       s.currentNode,
	})
}

func (s *Server) handleProxyTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req struct {
		Region string `json:"region,omitempty"`
	}
	if err := readJSON(r, &req); err != nil {
		if !errors.Is(err, io.EOF) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
	}

	var targetNodes []*subscription.Node
	if req.Region != "" {
		groups := region.GroupByRegion(s.getAllNodes())
		targetNodes = groups[req.Region]
	} else {
		targetNodes = s.getAllNodes()
	}
	if len(targetNodes) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no nodes available"})
		return
	}

	results := latency.TestAll(targetNodes, 5)
	var resp []map[string]interface{}
	for _, res := range results {
		item := map[string]interface{}{
			"name":    res.Node.Name,
			"latency": res.Latency.Milliseconds(),
		}
		if res.Err != nil {
			item["error"] = res.Err.Error()
		}
		resp = append(resp, item)
	}

	for _, res := range results {
		msg := map[string]interface{}{
			"type": "latency_progress",
			"name": res.Node.Name,
		}
		if res.Err != nil {
			msg["error"] = res.Err.Error()
		} else {
			msg["latency_ms"] = res.Latency.Milliseconds()
		}
		s.hub.Broadcast(msg)
	}

	writeJSON(w, http.StatusOK, resp)
}
```

- [ ] **Step 7: Rewrite web/handler.go to keep only shared helpers**

File: `web/handler.go`
```go
package web

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, `{"error":"json encode error"}`, http.StatusInternalServerError)
	}
}

func readJSON(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return json.Unmarshal([]byte{}, v)
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
```

- [ ] **Step 8: Verify Go code compiles**

Note: At this point `s.hub` doesn't exist yet. We'll add the WebSocket hub in Task 4. For now, temporarily comment out or remove the `s.hub.Broadcast(...)` calls from handler_proxy.go to verify compilation, then add them back after Task 4.

Actually — let's do Task 4 first before this verification. We'll verify both together after Task 4 is done.

- [ ] **Step 9: Commit (after Task 4 verification)**

---

## Task 4: Add WebSocket hub

**Files:**
- Create: `web/ws_hub.go`
- Create: `web/handler_ws.go`
- Modify: `web/server.go` (add hub field)

- [ ] **Step 1: Create web/ws_hub.go**

File: `web/ws_hub.go`
```go
package web

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type wsClient struct {
	conn *websocket.Conn
	send chan []byte
}

type wsHub struct {
	clients map[*wsClient]struct{}
	mu      sync.RWMutex
}

func newWsHub() *wsHub {
	return &wsHub{
		clients: make(map[*wsClient]struct{}),
	}
}

func (h *wsHub) Register(client *wsClient) {
	h.mu.Lock()
	h.clients[client] = struct{}{}
	h.mu.Unlock()
	log.Printf("WebSocket client connected, total: %d", len(h.clients))
}

func (h *wsHub) Unregister(client *wsClient) {
	h.mu.Lock()
	delete(h.clients, client)
	h.mu.Unlock()
	close(client.send)
	log.Printf("WebSocket client disconnected, total: %d", len(h.clients))
}

func (h *wsHub) Broadcast(msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("ws broadcast marshal error: %v", err)
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		select {
		case client.send <- data:
		default:
			go h.Unregister(client)
		}
	}
}

func (c *wsClient) writePump() {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func (c *wsClient) readPump() {
	defer func() {
		c.conn.Close()
	}()
	c.conn.SetReadLimit(512)
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
	}
}
```

- [ ] **Step 2: Create web/handler_ws.go**

File: `web/handler_ws.go`
```go
package web

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}
	if _, err := s.auth.ValidateToken(token); err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &wsClient{
		conn: conn,
		send: make(chan []byte, 64),
	}
	s.hub.Register(client)

	s.mu.RLock()
	status := map[string]interface{}{
		"type":       "proxy_status",
		"running":    s.isRunning,
		"node":       s.currentNode,
		"http_port":  s.httpPort,
		"socks_port": s.socksPort,
		"route_mode": s.cfg.RouteMode,
	}
	s.mu.RUnlock()
	s.hub.Broadcast(status)

	go client.writePump()
	go client.readPump()
}
```

- [ ] **Step 3: Update web/server.go to add hub field and WebSocket route**

File: `web/server.go` — replace entire file with:
```go
package web

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"xray-go/config"
	"xray-go/subscription"
)

type ProxyServer interface {
	Stop() error
}

type Server struct {
	httpServer  *http.Server
	auth        *AuthManager
	cfg         *config.Config
	hub         *wsHub
	proxy       ProxyServer
	currentNode *subscription.Node
	isRunning   bool
	httpPort    int
	socksPort   int
	mu          sync.RWMutex
}

func NewServer(addr string, cfg *config.Config) (*Server, error) {
	auth, err := NewAuthManager()
	if err != nil {
		return nil, fmt.Errorf("auth init: %w", err)
	}

	s := &Server{
		auth: auth,
		cfg:  cfg,
		hub:  newWsHub(),
	}

	mux := http.NewServeMux()
	s.registerRoutes(mux)

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s, nil
}

func (s *Server) Start() error {
	fmt.Printf("Web UI running at http://%s\n", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop() error {
	if s.proxy != nil {
		if err := s.proxy.Stop(); err != nil {
			return fmt.Errorf("proxy stop: %w", err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}
```

- [ ] **Step 4: Add WebSocket route to router.go**

In `web/router.go`, add this line inside `registerRoutes()` after the existing proxy routes:

```go
	mux.HandleFunc("/api/ws", s.handleWebSocket)
```

- [ ] **Step 5: Add missing imports to handler_subscription.go**

The `fetchWithFallback` function needs `fmt`, `log`, `time`, `config`, `singbox`, `xrayproxy`. Add these imports to `web/handler_subscription.go`:

```go
import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"xray-go/config"
	"xray-go/singbox"
	"xray-go/subscription"
	"xray-go/xrayproxy"
)
```

- [ ] **Step 6: Delete old web/static/app.js, style.css, index.html**

These will be replaced by the Vue build output. But we should keep the static directory for go:embed to work. Delete the old files:

```bash
rm /mnt/software/xray-go/web/static/app.js
rm /mnt/software/xray-go/web/static/style.css
rm /mnt/software/xray-go/web/static/index.html
```

- [ ] **Step 7: Add gorilla/websocket as direct dependency**

```bash
cd /mnt/software/xray-go
GOROOT=/mnt/go PATH=/mnt/go/bin:$PATH go get github.com/gorilla/websocket@v1.5.3
```

- [ ] **Step 8: Verify Go code compiles**

```bash
cd /mnt/software/xray-go
GOROOT=/mnt/go PATH=/mnt/go/bin:$PATH go build ./...
```

Expected: No errors

- [ ] **Step 9: Commit backend refactor**

```bash
cd /mnt/software/xray-go
git add web/
git commit -m "refactor(web): split handler.go into focused files, add WebSocket hub"
```

---

## Task 5: Add new backend API endpoints (settings, node CRUD)

**Files:**
- Create: `web/handler_settings.go`
- Modify: `web/handler_node.go` (add DELETE /api/nodes/{name})
- Modify: `web/router.go` (add new routes)

- [ ] **Step 1: Create web/handler_settings.go**

File: `web/handler_settings.go`
```go
package web

import (
	"net/http"
	"strings"

	"xray-go/config"
)

func (s *Server) handleRouteMode(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, map[string]string{"route_mode": string(s.cfg.RouteMode)})
		return
	}
	if r.Method == http.MethodPut {
		var req struct {
			RouteMode config.RouteMode `json:"route_mode"`
		}
		if err := readJSON(r, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		switch req.RouteMode {
		case config.RouteModeGlobal, config.RouteModeWhitelist, config.RouteModeBlacklist:
			s.cfg.RouteMode = req.RouteMode
		default:
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid route_mode"})
			return
		}
		if err := s.cfg.Save(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "updated"})
		return
	}
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func (s *Server) handleWhitelist(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, map[string][]string{"whitelist": s.cfg.Whitelist})
		return
	}
	if r.Method == http.MethodPut {
		var req struct {
			Whitelist []string `json:"whitelist"`
		}
		if err := readJSON(r, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		s.cfg.Whitelist = req.Whitelist
		if err := s.cfg.Save(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "updated"})
		return
	}
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func (s *Server) handleBlacklist(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, map[string][]string{"blacklist": s.cfg.Blacklist})
		return
	}
	if r.Method == http.MethodPut {
		var req struct {
			Blacklist []string `json:"blacklist"`
		}
		if err := readJSON(r, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		s.cfg.Blacklist = req.Blacklist
		if err := s.cfg.Save(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "updated"})
		return
	}
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func (s *Server) handleDeleteStandaloneNode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	rawPath := r.URL.RawPath
	if rawPath == "" {
		rawPath = r.URL.Path
	}
	rawPath = strings.TrimPrefix(rawPath, "/api/nodes/")
	name := rawPath

	for i, n := range s.cfg.StandaloneNodes {
		if n.Name == name {
			s.cfg.RemoveStandaloneNode(i)
			if err := s.cfg.Save(); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
			return
		}
	}
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "node not found"})
}
```

- [ ] **Step 2: Add new routes to router.go**

Add these routes inside `registerRoutes()` in `web/router.go`, after the existing node routes:

```go
	mux.HandleFunc("/api/nodes/", s.authMiddleware(s.handleNodesOrDelete))
	mux.HandleFunc("/api/settings/route-mode", s.authMiddleware(s.handleRouteMode))
	mux.HandleFunc("/api/settings/whitelist", s.authMiddleware(s.handleWhitelist))
	mux.HandleFunc("/api/settings/blacklist", s.authMiddleware(s.handleBlacklist))
```

Update the existing `/api/nodes/` line — replace:
```go
	mux.HandleFunc("/api/nodes/regions", s.authMiddleware(s.handleNodeRegions))
```
with the new routing function that dispatches based on path:

In `web/handler_node.go`, add:
```go
func (s *Server) handleNodesOrDelete(w http.ResponseWriter, r *http.Request) {
	rawPath := r.URL.RawPath
	if rawPath == "" {
		rawPath = r.URL.Path
	}
	if rawPath == "/api/nodes/regions" {
		s.handleNodeRegions(w, r)
		return
	}
	if strings.HasPrefix(rawPath, "/api/nodes/") && r.Method == http.MethodDelete {
		s.handleDeleteStandaloneNode(w, r)
		return
	}
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
}
```

Also add `"strings"` import to handler_node.go.

And update router.go to replace the old `/api/nodes/regions` route with:
```go
	mux.HandleFunc("/api/nodes/", s.authMiddleware(s.handleNodesOrDelete))
```

- [ ] **Step 3: Verify Go code compiles**

```bash
cd /mnt/software/xray-go
GOROOT=/mnt/go PATH=/mnt/go/bin:$PATH go build ./...
```

- [ ] **Step 4: Commit**

```bash
cd /mnt/software/xray-go
git add web/
git commit -m "feat(web): add settings and node delete API endpoints"
```

---

## Task 6: Build Vue frontend — Login page and layout

**Files:**
- Modify: `web/frontend/src/App.vue` (full layout with top nav)
- Modify: `web/frontend/src/views/LoginView.vue` (already created, finalize)
- Create: `web/frontend/src/components/TopNavBar.vue`
- Create: `web/frontend/src/stores/auth.ts`
- Create: `web/frontend/src/views/LayoutView.vue`

- [ ] **Step 1: Create auth store**

```bash
mkdir -p /mnt/software/xray-go/web/frontend/src/stores
```

File: `web/frontend/src/stores/auth.ts`
```ts
import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const username = ref('')

  function setToken(t: string) {
    token.value = t
    localStorage.setItem('token', t)
  }

  function clearToken() {
    token.value = ''
    username.value = ''
    localStorage.removeItem('token')
  }

  function isLoggedIn() {
    return !!token.value
  }

  return { token, username, setToken, clearToken, isLoggedIn }
})
```

- [ ] **Step 2: Create proxy store**

File: `web/frontend/src/stores/proxy.ts`
```ts
import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface ProxyStatus {
  running: boolean
  node: any | null
  http_port: number
  socks_port: number
  route_mode: string
}

export const useProxyStore = defineStore('proxy', () => {
  const status = ref<ProxyStatus>({
    running: false,
    node: null,
    http_port: 0,
    socks_port: 0,
    route_mode: '',
  })

  function updateStatus(s: Partial<ProxyStatus>) {
    Object.assign(status.value, s)
  }

  return { status, updateStatus }
})
```

- [ ] **Step 3: Create WebSocket composable**

File: `web/frontend/src/composables/useWebSocket.ts`
```ts
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useProxyStore } from '@/stores/proxy'

export function useWebSocket() {
  let ws: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let reconnectDelay = 1000
  const connected = ref(false)

  function connect() {
    const auth = useAuthStore()
    const proxy = useProxyStore()
    if (!auth.token) return

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const url = `${protocol}//${window.location.host}/api/ws?token=${auth.token}`

    ws = new WebSocket(url)

    ws.onopen = () => {
      connected.value = true
      reconnectDelay = 1000
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === 'proxy_status') {
          proxy.updateStatus({
            running: msg.running,
            node: msg.node,
            http_port: msg.http_port,
            socks_port: msg.socks_port,
            route_mode: msg.route_mode,
          })
        }
      } catch {
        // ignore parse errors
      }
    }

    ws.onclose = () => {
      connected.value = false
      scheduleReconnect()
    }

    ws.onerror = () => {
      ws?.close()
    }
  }

  function scheduleReconnect() {
    if (reconnectTimer) clearTimeout(reconnectTimer)
    reconnectTimer = setTimeout(() => {
      reconnectDelay = Math.min(reconnectDelay * 2, 30000)
      connect()
    }, reconnectDelay)
  }

  function disconnect() {
    if (reconnectTimer) clearTimeout(reconnectTimer)
    ws?.close()
    ws = null
    connected.value = false
  }

  return { connected, connect, disconnect }
}
```

- [ ] **Step 4: Create TopNavBar component**

```bash
mkdir -p /mnt/software/xray-go/web/frontend/src/components
```

File: `web/frontend/src/components/TopNavBar.vue`
```vue
<template>
  <el-menu mode="horizontal" :ellipsis="false" class="top-nav">
    <el-menu-item class="nav-logo">
      <span class="logo-text">xray-go</span>
      <el-badge is-dot :type="proxy.status.running ? 'success' : 'danger'" class="status-dot" />
    </el-menu-item>

    <el-menu-item
      v-for="item in navItems"
      :key="item.key"
      :index="item.key"
      :class="{ 'is-active': route.name === item.key }"
      @click="router.push({ name: item.key })"
    >
      {{ t(`nav.${item.key}`) }}
    </el-menu-item>

    <div class="nav-right">
      <el-button text @click="toggleLang">{{ lang === 'zh' ? 'EN' : '中' }}</el-button>
      <el-dropdown>
        <el-button text>
          <el-icon><User /></el-icon>
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item @click="handleLogout">{{ t('settings.logout') }}</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </el-menu>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { User } from '@element-plus/icons-vue'
import { useProxyStore } from '@/stores/proxy'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const { t, locale } = useI18n()
const proxy = useProxyStore()
const auth = useAuthStore()

const lang = computed(() => locale.value)

const navItems = [
  { key: 'nodes', route: '/' },
  { key: 'subscription', route: '/subscription' },
  { key: 'routing', route: '/routing' },
  { key: 'settings', route: '/settings' },
]

function toggleLang() {
  const newLang = locale.value === 'zh' ? 'en' : 'zh'
  locale.value = newLang
  localStorage.setItem('lang', newLang)
}

function handleLogout() {
  auth.clearToken()
  router.push('/login')
}
</script>

<style scoped>
.top-nav {
  display: flex;
  align-items: center;
  padding: 0 20px;
}
.nav-logo {
  display: flex;
  align-items: center;
  gap: 8px;
  pointer-events: none;
}
.logo-text {
  font-weight: 700;
  font-size: 18px;
}
.status-dot {
  margin-top: -2px;
}
.nav-right {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 4px;
}
</style>
```

- [ ] **Step 5: Create LayoutView (wrapper with top nav + content)**

File: `web/frontend/src/views/LayoutView.vue`
```vue
<template>
  <div class="app-layout">
    <TopNavBar />
    <div class="main-content">
      <router-view />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import TopNavBar from '@/components/TopNavBar.vue'
import { useWebSocket } from '@/composables/useWebSocket'

const ws = useWebSocket()

onMounted(() => {
  ws.connect()
})

onUnmounted(() => {
  ws.disconnect()
})
</script>

<style scoped>
.app-layout {
  height: 100vh;
  display: flex;
  flex-direction: column;
}
.main-content {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
  background: #f5f7fa;
}
</style>
```

- [ ] **Step 6: Update router to use layout**

File: `web/frontend/src/router/index.ts` — replace entire file:
```ts
import { createRouter, createWebHistory } from 'vue-router'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
    },
    {
      path: '/',
      component: () => import('@/views/LayoutView.vue'),
      children: [
        {
          path: '',
          name: 'nodes',
          component: () => import('@/views/NodesView.vue'),
        },
        {
          path: 'subscription',
          name: 'subscription',
          component: () => import('@/views/SubscriptionView.vue'),
        },
        {
          path: 'routing',
          name: 'routing',
          component: () => import('@/views/RoutingView.vue'),
        },
        {
          path: 'settings',
          name: 'settings',
          component: () => import('@/views/SettingsView.vue'),
        },
      ],
    },
  ],
})

router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('token')
  if (to.name !== 'login' && !token) {
    next({ name: 'login' })
  } else if (to.name === 'login' && token) {
    next({ name: 'nodes' })
  } else {
    next()
  }
})
```

- [ ] **Step 7: Update App.vue**

File: `web/frontend/src/App.vue`
```vue
<template>
  <router-view />
</template>
```

(unchanged, just confirming)

- [ ] **Step 8: Create placeholder view files for new pages**

File: `web/frontend/src/views/SubscriptionView.vue`
```vue
<template>
  <div><p>Subscription page placeholder</p></div>
</template>
```

File: `web/frontend/src/views/RoutingView.vue`
```vue
<template>
  <div><p>Routing page placeholder</p></div>
</template>
```

File: `web/frontend/src/views/SettingsView.vue`
```vue
<template>
  <div><p>Settings page placeholder</p></div>
</template>
```

- [ ] **Step 9: Verify frontend builds**

```bash
cd /mnt/software/xray-go/web/frontend
/mnt/nodejs/bin/npx vite build
```

Expected: Build succeeds, output in `web/static/`

- [ ] **Step 10: Commit**

```bash
cd /mnt/software/xray-go
git add web/frontend/
git commit -m "feat(web): add layout, top nav, login page, auth/proxy stores, WebSocket composable"
```

---

## Task 7: Build Vue frontend — Nodes page (home)

**Files:**
- Modify: `web/frontend/src/views/NodesView.vue`
- Create: `web/frontend/src/components/ProxyControl.vue`
- Create: `web/frontend/src/components/NodeTable.vue`
- Create: `web/frontend/src/components/AddNodeDialog.vue`
- Create: `web/frontend/src/stores/subscription.ts`

- [ ] **Step 1: Create subscription store**

File: `web/frontend/src/stores/subscription.ts`
```ts
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useApi } from '@/composables/useApi'

export const useSubscriptionStore = defineStore('subscription', () => {
  const subscriptions = ref<any[]>([])
  const standaloneNodes = ref<any[]>([])
  const api = useApi()

  async function loadConfig() {
    const data = await api.get('/api/config')
    subscriptions.value = data.subscriptions || []
    standaloneNodes.value = data.standalone_nodes || []
  }

  function allNodes(): any[] {
    const nodes: any[] = []
    for (const sub of subscriptions.value) {
      for (const n of sub.nodes || []) {
        nodes.push({ ...n, _source: sub.name })
      }
    }
    for (const n of standaloneNodes.value) {
      nodes.push({ ...n, _source: 'standalone' })
    }
    return nodes
  }

  return { subscriptions, standaloneNodes, loadConfig, allNodes }
})
```

- [ ] **Step 2: Create ProxyControl component**

File: `web/frontend/src/components/ProxyControl.vue`
```vue
<template>
  <el-card class="proxy-control">
    <div class="proxy-control-inner">
      <div class="proxy-info">
        <el-tag :type="proxy.status.running ? 'success' : 'danger'" size="large" effect="dark">
          {{ proxy.status.running ? t('proxy.running') : t('proxy.stopped') }}
        </el-tag>
        <template v-if="proxy.status.running">
          <span class="info-item">
            <span class="info-label">{{ t('proxy.currentNode') }}:</span>
            {{ proxy.status.node?.name || '-' }}
          </span>
          <span class="info-item">
            <span class="info-label">{{ t('proxy.httpPort') }}:</span>
            {{ proxy.status.http_port }}
          </span>
          <span class="info-item">
            <span class="info-label">{{ t('proxy.socksPort') }}:</span>
            {{ proxy.status.socks_port }}
          </span>
        </template>
      </div>
      <div class="proxy-actions">
        <el-button type="success" :disabled="proxy.status.running" :loading="starting" @click="handleStart">
          {{ t('proxy.start') }}
        </el-button>
        <el-button type="danger" :disabled="!proxy.status.running" :loading="stopping" @click="handleStop">
          {{ t('proxy.stop') }}
        </el-button>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { useProxyStore } from '@/stores/proxy'
import { useApi } from '@/composables/useApi'

const { t } = useI18n()
const proxy = useProxyStore()
const api = useApi()

const starting = ref(false)
const stopping = ref(false)

async function handleStart() {
  starting.value = true
  try {
    const data = await api.post('/api/proxy/start')
    proxy.updateStatus({
      running: true,
      node: data.node,
      http_port: data.http_port,
      socks_port: data.socks_port,
    })
    ElMessage.success(t('common.success'))
  } catch (e: any) {
    ElMessage.error(e.message || t('common.error'))
  } finally {
    starting.value = false
  }
}

async function handleStop() {
  stopping.value = true
  try {
    await api.post('/api/proxy/stop')
    proxy.updateStatus({ running: false, node: null, http_port: 0, socks_port: 0 })
  } catch (e: any) {
    ElMessage.error(e.message || t('common.error'))
  } finally {
    stopping.value = false
  }
}
</script>

<style scoped>
.proxy-control-inner {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.proxy-info {
  display: flex;
  align-items: center;
  gap: 16px;
}
.info-item {
  font-size: 14px;
}
.info-label {
  color: #909399;
  font-size: 12px;
  margin-right: 4px;
}
.proxy-actions {
  display: flex;
  gap: 8px;
}
</style>
```

- [ ] **Step 3: Create NodeTable component**

File: `web/frontend/src/components/NodeTable.vue`
```vue
<template>
  <div>
    <div class="node-toolbar">
      <el-select v-model="selectedRegion" :placeholder="t('node.region')" clearable style="width: 180px; margin-right: 12px;">
        <el-option :label="t('node.allRegions')" value="" />
        <el-option v-for="r in regions" :key="r" :label="`${r} (${regionCounts[r]})`" :value="r" />
      </el-select>
      <el-button type="primary" :loading="testing" @click="handleTestAll">
        {{ t('node.testLatency') }}
      </el-button>
      <el-button type="success" @click="addNodeVisible = true">
        + {{ t('node.add') }}
      </el-button>
    </div>

    <el-collapse v-model="expandedGroups">
      <el-collapse-item v-for="sub in subStore.subscriptions" :key="sub.name" :title="`${sub.name} (${(sub.nodes || []).length})`" :name="sub.name">
        <el-table :data="filterNodes(sub.nodes || [])" stripe size="small">
          <el-table-column prop="name" :label="t('node.name')" min-width="160" />
          <el-table-column :label="t('node.address')" min-width="140">
            <template #default="{ row }">{{ row.address }}:{{ row.port }}</template>
          </el-table-column>
          <el-table-column prop="protocol" :label="t('node.protocol')" width="100" />
          <el-table-column :label="t('node.latency')" width="100">
            <template #default="{ row }">
              <el-tag v-if="latencies[row.name] !== undefined" :type="latencyType(latencies[row.name])" size="small">
                {{ latencies[row.name] }}ms
              </el-tag>
              <span v-else-if="latencyErrors[row.name]" style="color:#f56c6c">✕</span>
            </template>
          </el-table-column>
          <el-table-column :label="t('common.confirm')" width="100">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleConnect(row.name)">{{ t('node.connect') }}</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-collapse-item>

      <el-collapse-item v-if="subStore.standaloneNodes.length > 0" :title="`Standalone (${subStore.standaloneNodes.length})`" name="standalone">
        <el-table :data="filterNodes(subStore.standaloneNodes)" stripe size="small">
          <el-table-column prop="name" :label="t('node.name')" min-width="160" />
          <el-table-column :label="t('node.address')" min-width="140">
            <template #default="{ row }">{{ row.address }}:{{ row.port }}</template>
          </el-table-column>
          <el-table-column prop="protocol" :label="t('node.protocol')" width="100" />
          <el-table-column :label="t('node.latency')" width="100">
            <template #default="{ row }">
              <el-tag v-if="latencies[row.name] !== undefined" :type="latencyType(latencies[row.name])" size="small">
                {{ latencies[row.name] }}ms
              </el-tag>
              <span v-else-if="latencyErrors[row.name]" style="color:#f56c6c">✕</span>
            </template>
          </el-table-column>
          <el-table-column :label="t('common.confirm')" width="160">
            <template #default="{ row }">
              <el-button type="primary" size="small" @click="handleConnect(row.name)">{{ t('node.connect') }}</el-button>
              <el-button type="danger" size="small" @click="handleDeleteNode(row.name)">{{ t('node.delete') }}</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-collapse-item>
    </el-collapse>

    <AddNodeDialog v-model:visible="addNodeVisible" @added="handleNodeAdded" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useSubscriptionStore } from '@/stores/subscription'
import { useProxyStore } from '@/stores/proxy'
import { useApi } from '@/composables/useApi'
import AddNodeDialog from './AddNodeDialog.vue'

const { t } = useI18n()
const subStore = useSubscriptionStore()
const proxy = useProxyStore()
const api = useApi()

const selectedRegion = ref('')
const testing = ref(false)
const latencies = ref<Record<string, number>>({})
const latencyErrors = ref<Record<string, boolean>>({})
const addNodeVisible = ref(false)
const expandedGroups = ref<string[]>([])

const regions = computed(() => {
  const set = new Set<string>()
  for (const n of subStore.allNodes()) {
    if (n._region) set.add(n._region)
  }
  return Array.from(set)
})

const regionCounts = computed(() => {
  const counts: Record<string, number> = {}
  for (const n of subStore.allNodes()) {
    if (n._region) counts[n._region] = (counts[n._region] || 0) + 1
  }
  return counts
})

function filterNodes(nodes: any[]) {
  if (!selectedRegion.value) return nodes
  return nodes.filter(n => n._region === selectedRegion.value)
}

function latencyType(ms: number): string {
  if (ms < 200) return 'success'
  if (ms < 500) return 'warning'
  return 'danger'
}

async function handleConnect(nodeName: string) {
  try {
    const data = await api.post('/api/proxy/start', { node_name: nodeName })
    proxy.updateStatus({
      running: true,
      node: data.node,
      http_port: data.http_port,
      socks_port: data.socks_port,
    })
    ElMessage.success(t('common.success'))
  } catch (e: any) {
    ElMessage.error(e.message || t('common.error'))
  }
}

async function handleTestAll() {
  testing.value = true
  latencies.value = {}
  latencyErrors.value = {}
  try {
    const data = await api.post('/api/proxy/test', {})
    if (Array.isArray(data)) {
      for (const item of data) {
        if (item.error) {
          latencyErrors.value[item.name] = true
        } else {
          latencies.value[item.name] = item.latency
        }
      }
    }
  } catch (e: any) {
    ElMessage.error(e.message || t('common.error'))
  } finally {
    testing.value = false
  }
}

async function handleDeleteNode(name: string) {
  try {
    await ElMessageBox.confirm(`Delete node "${name}"?`, t('common.confirm'), { type: 'warning' })
    await api.del(`/api/nodes/${encodeURIComponent(name)}`)
    await subStore.loadConfig()
  } catch {
    // cancelled or error
  }
}

function handleNodeAdded() {
  subStore.loadConfig()
}

onMounted(async () => {
  await subStore.loadConfig()
  const groups = subStore.subscriptions.map(s => s.name)
  if (subStore.standaloneNodes.length > 0) groups.push('standalone')
  expandedGroups.value = groups
})
</script>

<style scoped>
.node-toolbar {
  display: flex;
  align-items: center;
  margin-bottom: 16px;
}
</style>
```

- [ ] **Step 4: Create AddNodeDialog component**

File: `web/frontend/src/components/AddNodeDialog.vue`
```vue
<template>
  <el-dialog :model-value="visible" :title="t('node.add')" @update:model-value="$emit('update:visible', $event)" width="500px">
    <el-form @submit.prevent="handleSubmit">
      <el-form-item :label="t('node.link')">
        <el-input v-model="link" type="textarea" :rows="3" placeholder="vmess:// / vless:// / trojan:// / ss:// / anytls://" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="$emit('update:visible', false)">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" :loading="loading" @click="handleSubmit">{{ t('common.confirm') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { useApi } from '@/composables/useApi'

defineProps<{ visible: boolean }>()
const emit = defineEmits<{
  'update:visible': [value: boolean]
  added: []
}>()

const { t } = useI18n()
const api = useApi()
const link = ref('')
const loading = ref(false)

async function handleSubmit() {
  if (!link.value.trim()) return
  loading.value = true
  try {
    await api.post('/api/nodes', { link: link.value.trim() })
    link.value = ''
    emit('update:visible', false)
    emit('added')
    ElMessage.success(t('common.success'))
  } catch (e: any) {
    ElMessage.error(e.message || t('common.error'))
  } finally {
    loading.value = false
  }
}
</script>
```

- [ ] **Step 5: Update NodesView to use components**

File: `web/frontend/src/views/NodesView.vue`
```vue
<template>
  <div class="nodes-page">
    <ProxyControl style="margin-bottom: 20px;" />
    <NodeTable />
  </div>
</template>

<script setup lang="ts">
import ProxyControl from '@/components/ProxyControl.vue'
import NodeTable from '@/components/NodeTable.vue'
</script>

<style scoped>
.nodes-page {
  max-width: 1200px;
  margin: 0 auto;
}
</style>
```

- [ ] **Step 6: Verify frontend builds**

```bash
cd /mnt/software/xray-go/web/frontend
/mnt/nodejs/bin/npx vite build
```

- [ ] **Step 7: Commit**

```bash
cd /mnt/software/xray-go
git add web/frontend/
git commit -m "feat(web): implement Nodes page with proxy control, node table, add node dialog"
```

---

## Task 8: Build Vue frontend — Subscription page

**Files:**
- Modify: `web/frontend/src/views/SubscriptionView.vue`
- Create: `web/frontend/src/components/AddSubscriptionDialog.vue`

- [ ] **Step 1: Create AddSubscriptionDialog component**

File: `web/frontend/src/components/AddSubscriptionDialog.vue`
```vue
<template>
  <el-dialog :model-value="visible" :title="t('sub.add')" @update:model-value="$emit('update:visible', $event)" width="500px">
    <el-form @submit.prevent="handleSubmit">
      <el-form-item :label="t('sub.name')">
        <el-input v-model="form.name" />
      </el-form-item>
      <el-form-item :label="t('sub.url')">
        <el-input v-model="form.url" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="$emit('update:visible', false)">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" :loading="loading" @click="handleSubmit">{{ t('common.confirm') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { useApi } from '@/composables/useApi'

defineProps<{ visible: boolean }>()
const emit = defineEmits<{
  'update:visible': [value: boolean]
  added: []
}>()

const { t } = useI18n()
const api = useApi()
const loading = ref(false)
const form = reactive({ name: '', url: '' })

async function handleSubmit() {
  if (!form.name.trim() || !form.url.trim()) return
  loading.value = true
  try {
    await api.post('/api/subscriptions', { name: form.name, url: form.url })
    form.name = ''
    form.url = ''
    emit('update:visible', false)
    emit('added')
    ElMessage.success(t('common.success'))
  } catch (e: any) {
    ElMessage.error(e.message || t('common.error'))
  } finally {
    loading.value = false
  }
}
</script>
```

- [ ] **Step 2: Implement SubscriptionView**

File: `web/frontend/src/views/SubscriptionView.vue`
```vue
<template>
  <div class="sub-page">
    <div class="page-header">
      <h2>{{ t('sub.title') }}</h2>
      <el-button type="primary" @click="addVisible = true">+ {{ t('sub.add') }}</el-button>
    </div>

    <el-table :data="subStore.subscriptions" stripe>
      <el-table-column prop="name" :label="t('sub.name')" min-width="120" />
      <el-table-column prop="url" :label="t('sub.url')" min-width="240" show-overflow-tooltip />
      <el-table-column :label="t('sub.nodes')" width="80">
        <template #default="{ row }">{{ (row.nodes || []).length }}</template>
      </el-table-column>
      <el-table-column :label="t('sub.lastFetched')" width="180">
        <template #default="{ row }">{{ row.last_fetched ? new Date(row.last_fetched).toLocaleString() : '-' }}</template>
      </el-table-column>
      <el-table-column :label="t('sub.delete')" width="180">
        <template #default="{ row }">
          <el-button size="small" :loading="refreshingMap[row.name]" @click="handleRefresh(row.name)">{{ t('sub.refresh') }}</el-button>
          <el-button type="danger" size="small" @click="handleDelete(row.name)">{{ t('sub.delete') }}</el-button>
        </template>
      </el-table-column>
    </el-table>

    <AddSubscriptionDialog v-model:visible="addVisible" @added="subStore.loadConfig()" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useSubscriptionStore } from '@/stores/subscription'
import { useApi } from '@/composables/useApi'
import AddSubscriptionDialog from '@/components/AddSubscriptionDialog.vue'

const { t } = useI18n()
const subStore = useSubscriptionStore()
const api = useApi()

const addVisible = ref(false)
const refreshingMap = reactive<Record<string, boolean>>({})

async function handleRefresh(name: string) {
  refreshingMap[name] = true
  try {
    await api.post(`/api/subscriptions/${encodeURIComponent(name)}/refresh`)
    await subStore.loadConfig()
    ElMessage.success(t('sub.refreshSuccess'))
  } catch (e: any) {
    ElMessage.error(e.message || t('common.error'))
  } finally {
    refreshingMap[name] = false
  }
}

async function handleDelete(name: string) {
  try {
    await ElMessageBox.confirm(`Delete subscription "${name}"?`, t('common.confirm'), { type: 'warning' })
    await api.del(`/api/subscriptions/${encodeURIComponent(name)}`)
    await subStore.loadConfig()
  } catch {
    // cancelled
  }
}

onMounted(() => {
  subStore.loadConfig()
})
</script>

<style scoped>
.sub-page {
  max-width: 1200px;
  margin: 0 auto;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
</style>
```

- [ ] **Step 3: Verify frontend builds**

```bash
cd /mnt/software/xray-go/web/frontend
/mnt/nodejs/bin/npx vite build
```

- [ ] **Step 4: Commit**

```bash
cd /mnt/software/xray-go
git add web/frontend/
git commit -m "feat(web): implement Subscription page with add/refresh/delete"
```

---

## Task 9: Build Vue frontend — Routing page

**Files:**
- Modify: `web/frontend/src/views/RoutingView.vue`
- Create: `web/frontend/src/components/RouteRuleEditor.vue`

- [ ] **Step 1: Create RouteRuleEditor component**

File: `web/frontend/src/components/RouteRuleEditor.vue`
```vue
<template>
  <div>
    <el-tag v-for="(rule, idx) in rules" :key="idx" closable style="margin: 0 4px 4px 0;" @close="removeRule(idx)">
      {{ rule }}
    </el-tag>
    <el-input
      v-if="inputVisible"
      ref="inputRef"
      v-model="inputValue"
      size="small"
      style="width: 180px;"
      @keyup.enter="addRule"
      @blur="addRule"
    />
    <el-button v-else size="small" @click="showInput">+ {{ t('routing.addRule') }}</el-button>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{ rules: string[] }>()
const emit = defineEmits<{ 'update:rules': [value: string[]] }>()

const { t } = useI18n()
const inputVisible = ref(false)
const inputValue = ref('')
const inputRef = ref()

function removeRule(idx: number) {
  const newRules = [...props.rules]
  newRules.splice(idx, 1)
  emit('update:rules', newRules)
}

async function showInput() {
  inputVisible.value = true
  await nextTick()
  inputRef.value?.focus()
}

function addRule() {
  const val = inputValue.value.trim()
  if (val && !props.rules.includes(val)) {
    emit('update:rules', [...props.rules, val])
  }
  inputVisible.value = false
  inputValue.value = ''
}
</script>
```

- [ ] **Step 2: Implement RoutingView**

File: `web/frontend/src/views/RoutingView.vue`
```vue
<template>
  <div class="routing-page">
    <h2 style="margin-bottom: 20px;">{{ t('routing.title') }}</h2>

    <el-card style="margin-bottom: 20px;">
      <h3>{{ t('routing.mode') }}</h3>
      <el-radio-group v-model="routeMode" style="margin-top: 12px;">
        <el-radio value="global">{{ t('routing.global') }}</el-radio>
        <el-radio value="whitelist">{{ t('routing.whitelist') }}</el-radio>
        <el-radio value="blacklist">{{ t('routing.blacklist') }}</el-radio>
      </el-radio-group>
    </el-card>

    <el-card v-if="routeMode === 'whitelist'" style="margin-bottom: 20px;">
      <h3>{{ t('routing.whitelistLabel') }}</h3>
      <div style="margin-top: 12px;">
        <RouteRuleEditor :rules="whitelist" @update:rules="whitelist = $event" />
      </div>
    </el-card>

    <el-card v-if="routeMode === 'blacklist'" style="margin-bottom: 20px;">
      <h3>{{ t('routing.blacklistLabel') }}</h3>
      <div style="margin-top: 12px;">
        <RouteRuleEditor :rules="blacklist" @update:rules="blacklist = $event" />
      </div>
    </el-card>

    <el-button type="primary" :loading="saving" @click="handleSave">{{ t('routing.save') }}</el-button>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { useApi } from '@/composables/useApi'
import RouteRuleEditor from '@/components/RouteRuleEditor.vue'

const { t } = useI18n()
const api = useApi()

const routeMode = ref('global')
const whitelist = ref<string[]>([])
const blacklist = ref<string[]>([])
const saving = ref(false)

async function loadSettings() {
  try {
    const [rm, wl, bl] = await Promise.all([
      api.get('/api/settings/route-mode'),
      api.get('/api/settings/whitelist'),
      api.get('/api/settings/blacklist'),
    ])
    routeMode.value = rm.route_mode || 'global'
    whitelist.value = wl.whitelist || []
    blacklist.value = bl.blacklist || []
  } catch {
    // use defaults
  }
}

async function handleSave() {
  saving.value = true
  try {
    await api.put('/api/settings/route-mode', { route_mode: routeMode.value })
    if (routeMode.value === 'whitelist') {
      await api.put('/api/settings/whitelist', { whitelist: whitelist.value })
    }
    if (routeMode.value === 'blacklist') {
      await api.put('/api/settings/blacklist', { blacklist: blacklist.value })
    }
    ElMessage.success(t('common.success'))
  } catch (e: any) {
    ElMessage.error(e.message || t('common.error'))
  } finally {
    saving.value = false
  }
}

onMounted(loadSettings)
</script>

<style scoped>
.routing-page {
  max-width: 800px;
  margin: 0 auto;
}
</style>
```

- [ ] **Step 3: Verify frontend builds**

```bash
cd /mnt/software/xray-go/web/frontend
/mnt/nodejs/bin/npx vite build
```

- [ ] **Step 4: Commit**

```bash
cd /mnt/software/xray-go
git add web/frontend/
git commit -m "feat(web): implement Routing page with mode selector and rule editor"
```

---

## Task 10: Build Vue frontend — Settings page

**Files:**
- Modify: `web/frontend/src/views/SettingsView.vue`

- [ ] **Step 1: Implement SettingsView**

File: `web/frontend/src/views/SettingsView.vue`
```vue
<template>
  <div class="settings-page">
    <h2 style="margin-bottom: 20px;">{{ t('settings.title') }}</h2>

    <el-card style="margin-bottom: 20px;">
      <el-form label-width="140px">
        <el-form-item :label="t('settings.language')">
          <el-switch
            :model-value="locale === 'zh'"
            active-text="中文"
            inactive-text="EN"
            @change="toggleLang"
          />
        </el-form-item>
      </el-form>
    </el-card>

    <el-card style="margin-bottom: 20px;">
      <el-form label-width="140px">
        <el-form-item :label="t('proxy.routeMode')">
          <el-select v-model="routeMode" style="width: 200px;">
            <el-option value="global" :label="t('routing.global')" />
            <el-option value="whitelist" :label="t('routing.whitelist')" />
            <el-option value="blacklist" :label="t('routing.blacklist')" />
          </el-select>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card>
      <el-button type="danger" @click="handleLogout">{{ t('settings.logout') }}</el-button>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { useApi } from '@/composables/useApi'

const { t, locale } = useI18n()
const router = useRouter()
const auth = useAuthStore()
const api = useApi()

const routeMode = ref('global')

function toggleLang(val: boolean) {
  const newLang = val ? 'zh' : 'en'
  locale.value = newLang
  localStorage.setItem('lang', newLang)
}

async function loadRouteMode() {
  try {
    const data = await api.get('/api/settings/route-mode')
    routeMode.value = data.route_mode || 'global'
  } catch {
    // use default
  }
}

function handleLogout() {
  auth.clearToken()
  router.push('/login')
}

onMounted(loadRouteMode)
</script>

<style scoped>
.settings-page {
  max-width: 600px;
  margin: 0 auto;
}
</style>
```

- [ ] **Step 2: Verify frontend builds**

```bash
cd /mnt/software/xray-go/web/frontend
/mnt/nodejs/bin/npx vite build
```

- [ ] **Step 3: Commit**

```bash
cd /mnt/software/xray-go
git add web/frontend/
git commit -m "feat(web): implement Settings page with language toggle and logout"
```

---

## Task 11: Integration build and Go embed verification

**Files:**
- Verify: `web/static/` contains built frontend
- Verify: Go binary embeds and serves frontend correctly
- Verify: CLI mode still works independently

- [ ] **Step 1: Clean build frontend**

```bash
cd /mnt/software/xray-go/web/frontend
rm -rf ../../web/static/*
/mnt/nodejs/bin/npx vite build
```

Expected: `web/static/` contains `index.html` and `assets/` directory

- [ ] **Step 2: Verify Go build with embedded frontend**

```bash
cd /mnt/software/xray-go
GOROOT=/mnt/go PATH=/mnt/go/bin:$PATH go build -o xray-go-test .
```

Expected: Build succeeds

- [ ] **Step 3: Verify CLI mode is unaffected**

```bash
cd /mnt/software/xray-go
GOROOT=/mnt/go PATH=/mnt/go/bin:$PATH go run . --help
```

Expected: Help output shows both CLI and web commands

- [ ] **Step 4: Clean up test binary**

```bash
rm /mnt/software/xray-go/xray-go-test
```

- [ ] **Step 5: Commit any final adjustments**

```bash
cd /mnt/software/xray-go
git add web/
git commit -m "chore(web): verify integration build and embed"
```

---

## Task 12: SPA fallback for Vue Router history mode

**Files:**
- Modify: `web/router.go` (update spaHandler to handle Vue Router pushState)

- [ ] **Step 1: Update spaHandler to serve index.html for all non-API, non-static routes**

In `web/router.go`, replace the `spaHandler()` method. The current implementation already does this (falls back to index.html), but we need to ensure it also handles nested paths like `/subscription`, `/routing`, `/settings` correctly.

The current logic already serves `index.html` as fallback, which is correct for Vue Router history mode. No changes needed if the current fallback works.

Let's verify: the current spaHandler reads the file path, and if it doesn't find it, falls back to `index.html`. This is already correct for SPA routing.

- [ ] **Step 2: Commit if changes were needed (otherwise skip)**

---

## Self-Review Checklist

1. **Spec coverage**: 
   - Vue 3 + Element Plus ✓ (Task 2, 6-10)
   - Top nav + multi-page ✓ (Task 6)
   - WebSocket real-time ✓ (Task 4)
   - Frontend in web/frontend/ + go:embed ✓ (Task 2, 11)
   - Build deps in /mnt subdirs ✓ (Task 1)
   - Backend handler split ✓ (Task 3)
   - New API endpoints ✓ (Task 5)
   - Nodes page ✓ (Task 7)
   - Subscription page ✓ (Task 8)
   - Routing page ✓ (Task 9)
   - Settings page ✓ (Task 10)
   - i18n zh/en ✓ (Task 2)
   - Login/Register ✓ (Task 6)
   - Modularity (CLI vs web) ✓ (Task 3, 11)

2. **Placeholder scan**: No TBD/TODO found.

3. **Type consistency**: All API paths, component names, store methods are consistent across tasks.
