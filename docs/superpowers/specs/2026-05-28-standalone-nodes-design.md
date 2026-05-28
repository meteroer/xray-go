# 独立节点管理功能设计

日期：2026-05-28

## 背景

当前 xray-go 仅支持通过订阅 URL 批量获取节点。用户需要在不依赖订阅的情况下，手动添加单个节点（通过 `vmess://`、`vless://` 等分享链接）。

## 目标

允许用户通过交互式菜单，手动添加、删除和选择独立节点，流程与订阅节点完全一致（地区分组、测速、启动代理）。

## 架构

### 数据模型

在 `config.Config` 中新增字段：

```go
StandaloneNodes []*subscription.Node `json:"standalone_nodes,omitempty"`
```

完全向后兼容：
- 旧版配置无此字段 → 解析为空切片，不影响任何逻辑
- 新版配置有独立节点 → 订阅逻辑零影响

### 交互设计

#### 主菜单（`selectSubscription`）

在订阅列表后、新增订阅选项前，插入"独立节点"选项：

```
Saved subscriptions:
   1. *sub1 (https://...) [47 cached]
   2. sub2 (https://...) [23 cached]

Manual Nodes (3 nodes)

   3. + Add new subscription
   4. - Delete a subscription
   5. Exit
```

- 带 `*` 标记的是上次使用的项目
- "Manual Nodes" 显示当前独立节点数量

#### 独立节点子菜单

选择"Manual Nodes"后：

```
Manual Nodes (3 nodes):
   1. 香港 01 [vless]
   2. 日本 01 [vmess]
   3. + Add new node
   4. - Delete a node
   5. Back
```

- 选择已有节点 → 直接进入地区选择 → 测速 → 启动代理
- `+ Add new node` → 提示粘贴分享链接 → 解析 → 保存
- `- Delete a node` → 列出节点 → 选择删除 → 保存
- `Back` → 返回主菜单

### 解析逻辑

将 `subscription/parser.go` 中的 `parseLine` 提升为公开函数：

```go
func ParseNode(link string) (*Node, error)
```

复用现有解析逻辑，支持全部协议：vmess、vless、trojan、shadowsocks、anytls。

### 代理启动流程

选中独立节点后，流程与订阅节点完全一致：

1. 地区分组（`region.GroupByRegion`）
2. 交互式地区选择（或直接使用上次保存的地区）
3. 并发测速（`latency.FindBest`）
4. 启动代理（`xrayproxy.Start` 或 `singbox.Start`）

### 无交互启动模式（`start` 子命令）

当前逻辑：使用 `LastUsedSub` 找到对应订阅并启动。

扩展逻辑：如果 `LastUsedSub` 为空，且 `StandaloneNodes` 非空：
- 使用 `LastRegion`（保存在独立节点的上下文中）
- 或直接对所有独立节点测速选最优
- 启动代理

**实现方式**：在 `Config` 中新增 `LastUsedStandalone` bool 字段，标记上次使用的是独立节点还是订阅。当 `LastUsedSub == ""` 且 `LastUsedStandalone == true` 时，从独立节点中启动。

## 文件变更

| 文件 | 变更 |
|------|------|
| `config/store.go` | 新增 `StandaloneNodes` 和 `LastUsedStandalone` 字段；新增 `AddStandaloneNode`、`RemoveStandaloneNode` 方法 |
| `subscription/parser.go` | 将 `parseLine` 暴露为 `ParseNode` |
| `main.go` | 主菜单插入"Manual Nodes"选项；新增独立节点子菜单、添加节点、删除节点交互逻辑；修改 `startMode` 支持独立节点 |

## 错误处理

- 解析分享链接失败 → 提示错误，让用户重新输入
- 独立节点列表为空 → 子菜单只显示 `+ Add new node` 和 `Back`
- 删除最后一个节点后 → 自动返回主菜单

## 边界情况

- 独立节点和订阅可以共存，互不影响
- 订阅更新不会覆盖独立节点
- 用户可以同时使用订阅和独立节点，随时切换
- 独立节点名重复 → 允许重复（按索引区分），或提示重命名
