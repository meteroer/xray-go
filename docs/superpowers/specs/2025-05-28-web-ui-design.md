# xray-go Web UI Design

## 概述

为 xray-go 增加 `web` 子命令，启动一个本地 Web 服务（端口 18700，监听 0.0.0.0），提供图形化界面进行订阅管理、节点配置和代理控制。支持中英文切换。

## 架构设计

### 技术选型

- **后端**: 标准库 `net/http`，零额外 HTTP 依赖
- **前端**: 原生 HTML/CSS/JS，通过 `go:embed` 嵌入二进制
- **认证**: JWT + bcrypt
- **数据存储**: 复用现有 `~/.xray-go/config.json`

### API 路由

```
GET  /                    → 前端单页面应用
POST /api/auth/init       → 首次创建用户（用户名+密码）
POST /api/auth/login      → 登录，返回 JWT
GET  /api/auth/status     → 检查是否已创建用户/登录状态
POST /api/auth/logout     → 登出
GET  /api/config          → 获取完整配置
POST /api/subscriptions   → 添加订阅
DELETE /api/subscriptions/{name} → 删除订阅
POST /api/subscriptions/{name}/refresh → 刷新订阅
POST /api/nodes           → 添加手动节点
DELETE /api/nodes/{index} → 删除手动节点
GET  /api/nodes           → 获取所有节点
GET  /api/nodes/regions   → 获取地区分组
POST /api/proxy/start     → 启动代理
POST /api/proxy/stop      → 停止代理
GET  /api/proxy/status    → 获取代理状态
POST /api/proxy/test      → 测速（可选：指定节点，否则全部）
```

### 页面结构

单页面应用（SPA），左侧导航栏 + 右侧内容区：

**导航项：**
- 概览（Dashboard）— 代理状态、当前节点、延迟
- 订阅（Subscriptions）— 订阅列表、添加/删除/刷新
- 节点（Nodes）— 节点列表、地区筛选、测速、手动选择
- 设置（Settings）— 路由模式、语言切换

**首次访问流程：**
1. 检测是否已创建用户 → 否 → 显示初始化页面（创建用户名+密码）
2. 已创建用户 → 显示登录页面
3. 登录成功 → 进入主页面

### 安全设计

- **监听地址**: 0.0.0.0:18700（局域网可访问）
- **密码存储**: bcrypt 哈希，不保存明文
- **认证方式**: JWT，存储在浏览器 localStorage
- **CORS**: 限制为同源请求
- **敏感操作**: 启动/停止代理、修改配置均需认证

### 数据流

```
浏览器                  API Server              Proxy Engine
  │                       │                        │
  │  1. 打开页面          │                        │
  │ ─────────────────────>│                        │
  │                       │  2. 检查 JWT           │
  │  3. 未登录 → 登录页    │                        │
  │ <─────────────────────│                        │
  │                       │                        │
  │  4. 登录成功          │                        │
  │ ─────────────────────>│                        │
  │  5. 获取配置          │                        │
  │ <─────────────────────│                        │
  │                       │                        │
  │  6. 点击启动代理       │                        │
  │ ─────────────────────>│  7. 调用现有启动逻辑   │
  │                       │ ──────────────────────>│
  │  8. 返回成功          │                        │
  │ <─────────────────────│  9. 运行中             │
  │                       │                        │
  │ 10. 轮询状态 (3s)     │                        │
  │ ─────────────────────>│  11. 返回当前状态      │
  │ <─────────────────────│                        │
```

### 文件结构

```
web/
├── handler.go           # HTTP 路由和 API handler
├── auth.go              # JWT 和 bcrypt 认证
├── server.go            # HTTP 服务器启动/停止
└── static/
    ├── index.html       # 单页面入口
    ├── style.css        # 样式
    └── app.js           # 前端逻辑
```

### 状态管理

Web 服务维护全局状态：
- `currentProxy`: 当前运行的代理实例（`ProxyServer` 接口）
- `currentNode`: 当前使用的节点
- `isRunning`: 代理是否运行中
- `httpPort`/`socksPort`: 代理端口

### 国际化

- 前端维护中英文两套文本
- 通过按钮切换，保存到 localStorage
- 页面加载时读取语言偏好

## 实现计划

1. 创建 `web/` 目录结构
2. 实现 `auth.go` — JWT 签发/验证 + bcrypt 哈希
3. 实现 `server.go` — HTTP 服务器启动/停止
4. 实现 `handler.go` — API 路由和 handler
5. 编写前端代码（index.html, style.css, app.js）
6. 修改 `main.go` — 添加 `web` 子命令
7. 集成测试

## 兼容性

- 现有 CLI 功能完全保留
- `web` 子命令为独立入口，不影响现有 `start` 等命令
- 配置文件格式不变，Web 和 CLI 共享同一配置
