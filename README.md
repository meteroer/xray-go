# xray-go

单一二进制代理工具，内嵌 xray-core 和 sing-box 双核心，自动测速选节点，按地区筛选。

## 特性

- **单一二进制**，无外部依赖，直接运行
- **双核心架构**：vmess/vless/trojan/shadowsocks 走 xray-core，anytls 走 sing-box
- **地区选择**：解析订阅后列出所有地区，选择目标地区再测速
- **自动测速**：并发 TCP 连接测试，选出延迟最低的节点
- **订阅地址持久化**：首次输入后保存到 `~/.xray-go/config.json`，下次自动使用
- 支持 Reality、XTLS、WebSocket、gRPC 等传输方式

## 支持协议

| 协议 | 核心 | 状态 |
|------|------|------|
| VMess | xray-core | ✓ |
| VLESS + Reality | xray-core | ✓ |
| Trojan | xray-core | ✓ |
| Shadowsocks | xray-core | ✓ |
| AnyTLS | sing-box | ✓ |

## 安装

从源码编译（需要 Go 1.24+）：

```bash
git clone <repo-url> xray-go
cd xray-go
go build -o xray-go .
```

或直接下载预编译二进制。

## 使用

### 首次运行

```bash
./xray-go
```

程序会提示输入订阅地址，然后：

1. 获取订阅 → 解析节点
2. 列出所有地区供选择
3. 对选中地区节点测速
4. 自动连接延迟最低的节点
5. 启动本地代理

输出示例：

```
Fetching subscription...
Found 47 nodes

Available regions:
   1. 香港 (9 nodes)
   2. 台湾 (4 nodes)
   3. 日本 (6 nodes)
   4. 新加坡 (6 nodes)
   ...
  22. All regions

Select region number: 1

Selected region: 香港 (9 nodes)

Testing latency...
  + 香港 01: 167ms
  + 香港 02: 165ms
  + 香港 03: 170ms
  ...
Best node: 香港 02 (165ms)

Starting proxy on 0.0.0.0:16708 (HTTP) and 0.0.0.0:16709 (SOCKS5) [blacklist mode]...
Proxy running at 0.0.0.0:16708 (HTTP) and 0.0.0.0:16709 (SOCKS5) [blacklist mode]
```

按 `Ctrl+C` 退出。

### 无交互启动（后台运行）

```bash
./xray-go start
```

跳过地区选择，自动使用上次保存的节点和地区，直接启动代理。

```bash
./xray-go start --port 1080
./xray-go start --update  # 强制更新订阅后启动
```

### 指定订阅地址

```bash
./xray-go --url "https://example.com/sub"
```

### 指定代理端口

```bash
./xray-go --port 1080
```

HTTP 代理监听 `0.0.0.0:<port>`，SOCKS5 代理监听 `0.0.0.0:<port+1>`（所有网卡，局域网可访问）。

默认端口：HTTP `16708`，SOCKS5 `16709`。

### 强制更新订阅

```bash
./xray-go --update
```

重新提示输入订阅地址并测速。

### 使用代理

```bash
# HTTP 代理
curl -x http://0.0.0.0:16708 https://api.ipify.org

# SOCKS5 代理
curl -x socks5h://0.0.0.0:16709 https://api.ipify.org

# 环境变量
export http_proxy=http://0.0.0.0:16708
export https_proxy=http://0.0.0.0:16708
```

## 命令行参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `start` | - | 无交互启动（使用上次保存的节点和地区） |
| `--url` | 空 | 订阅地址（覆盖已保存配置） |
| `--port` | 16708 | HTTP 代理端口 |
| `--update` | false | 强制重新获取订阅并测速 |

## 配置文件

路径：`~/.xray-go/config.json`

```json
{
  "subscription_url": "https://example.com/sub",
  "selected_node": "香港 02"
}
```

## 项目结构

```
xray-go/
├── main.go            -- 入口，地区选择，启动代理
├── config/store.go    -- 配置读写，订阅地址提示
├── region/region.go   -- 地区检测，分组，交互选择
├── subscription/
│   ├── fetcher.go     -- HTTP 获取订阅内容
│   └── parser.go      -- 解析 vmess/vless/trojan/ss/anytls
├── latency/tester.go  -- 并发测速，选择最优节点
├── xrayproxy/server.go -- xray-core 代理启动/停止
├── singbox/server.go  -- sing-box 代理启动/停止
├── go.mod / go.sum
└── README.md
```

## 核心版本

- xray-core: v26.3.27
- sing-box: v1.13.12