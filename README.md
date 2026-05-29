# xray-go

单一二进制代理工具，内嵌 xray-core 和 sing-box 双核心，支持自动测速选节点、Web 管理界面。

## 快速开始

```bash
./xray-go
```

首次运行输入订阅地址，程序自动获取节点、测速、启动代理。

## Web 管理界面

![Web 界面](web.png)

运行后访问 `http://localhost:16708`，支持：

- 节点浏览、按地区筛选、手动测速
- 一键切换节点（热切换，无需重启代理）
- 订阅管理（添加、刷新、删除）
- 路由模式切换（全局/白名单/黑名单）
- 端口、语言等系统设置

## 使用方式

### 交互模式（默认）

```bash
./xray-go
```

列出所有地区 → 选择 → 测速 → 自动连接最优节点。

### 命令行模式

```bash
./xray-go start                          # 无交互启动（使用上次配置）
./xray-go --url "https://example.com/sub" # 指定订阅地址
./xray-go --port 1080                     # 指定端口
./xray-go --update                        # 强制更新订阅
```

### 代理地址

| 类型 | 地址 |
|------|------|
| HTTP | `0.0.0.0:16708` |
| SOCKS5 | `0.0.0.0:16709` |

```bash
curl -x http://127.0.0.1:16708 https://api.ipify.org
curl -x socks5h://127.0.0.1:16709 https://api.ipify.org
```

## 支持协议

VMess / VLESS + Reality / Trojan / Shadowsocks（xray-core）· AnyTLS（sing-box）

## 配置文件

`~/.xray-go/config.json` — 订阅、节点、路由模式等持久化存储。
