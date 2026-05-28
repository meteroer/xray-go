# 独立节点管理功能实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 允许用户通过交互式菜单手动添加、删除和选择独立节点。

**Architecture:** 在配置中新增独立节点池，复用现有解析和代理启动逻辑。主菜单插入"Manual Nodes"选项，选中后进入独立节点子菜单。

**Tech Stack:** Go 1.24+

---

## 文件结构

| 文件 | 责任 |
|------|------|
| `config/store.go` | 配置模型：新增 `StandaloneNodes` 字段，新增增删方法 |
| `subscription/parser.go` | 将 `parseLine` 暴露为公开函数 `ParseNode` |
| `subscription/parser_test.go` | `ParseNode` 的单元测试 |
| `main.go` | 交互逻辑：主菜单插入独立节点选项，新增子菜单、添加/删除节点逻辑 |

---

### Task 1: 暴露 `ParseNode` 函数并添加测试

**Files:**
- Modify: `subscription/parser.go:61-78`
- Create: `subscription/parser_test.go`

- [ ] **Step 1: 将 `parseLine` 改名为 `ParseNode` 并暴露**

```go
// subscription/parser.go
func ParseNode(link string) (*Node, error) {
	if strings.HasPrefix(link, "vmess://") {
		return parseVmess(link[8:])
	}
	if strings.HasPrefix(link, "vless://") {
		return parseVless(link[8:])
	}
	if strings.HasPrefix(link, "trojan://") {
		return parseTrojan(link[9:])
	}
	if strings.HasPrefix(link, "ss://") {
		return parseShadowsocks(link[5:])
	}
	if strings.HasPrefix(link, "anytls://") {
		return parseAnyTLS(line[9:])
	}
	return nil, fmt.Errorf("unsupported protocol")
}
```

更新 `Parse` 函数内部调用 `ParseNode`：

```go
func Parse(data []byte) ([]*Node, error) {
	// ... 解码逻辑 ...
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		node, err := ParseNode(line)
		if err != nil {
			continue
		}
		nodes = append(nodes, node)
	}
	// ...
}
```

- [ ] **Step 2: 创建 `ParseNode` 单元测试**

```go
// subscription/parser_test.go
package subscription

import (
	"testing"
)

func TestParseNode_Vmess(t *testing.T) {
	link := "vmess://eyJhZGQiOiJleGFtcGxlLmNvbSIsInBvcnQiOiI0NDMiLCJpZCI6InV1aWQiLCJuZXQiOiJ3cyIsInBzIjoidGVzdCJ9"
	node, err := ParseNode(link)
	if err != nil {
		t.Fatalf("parse vmess failed: %v", err)
	}
	if node.Name != "test" {
		t.Errorf("expected name 'test', got '%s'", node.Name)
	}
	if node.Protocol != "vmess" {
		t.Errorf("expected protocol 'vmess', got '%s'", node.Protocol)
	}
	if node.Address != "example.com" {
		t.Errorf("expected address 'example.com', got '%s'", node.Address)
	}
	if node.Port != 443 {
		t.Errorf("expected port 443, got %d", node.Port)
	}
}

func TestParseNode_Vless(t *testing.T) {
	link := "vless://uuid@example.com:443?type=tcp#test-node"
	node, err := ParseNode(link)
	if err != nil {
		t.Fatalf("parse vless failed: %v", err)
	}
	if node.Name != "test-node" {
		t.Errorf("expected name 'test-node', got '%s'", node.Name)
	}
	if node.Protocol != "vless" {
		t.Errorf("expected protocol 'vless', got '%s'", node.Protocol)
	}
	if node.UUID != "uuid" {
		t.Errorf("expected uuid 'uuid', got '%s'", node.UUID)
	}
}

func TestParseNode_Unsupported(t *testing.T) {
	_, err := ParseNode("unknown://data")
	if err == nil {
		t.Error("expected error for unsupported protocol")
	}
}
```

- [ ] **Step 3: 运行测试**

```bash
cd /mnt/software/xray-go
go test ./subscription -v
```

Expected: 全部 PASS

- [ ] **Step 4: 提交**

```bash
git add subscription/parser.go subscription/parser_test.go
git commit -m "feat: expose ParseNode and add unit tests"
```

---

### Task 2: 扩展配置模型支持独立节点

**Files:**
- Modify: `config/store.go`

- [ ] **Step 1: 新增 `StandaloneNodes` 字段和辅助方法**

```go
// config/store.go

type Config struct {
	Subscriptions      []*Subscription      `json:"subscriptions"`
	LastUsedSub        string               `json:"last_used_subscription"`
	LastUsedStandalone bool                 `json:"last_used_standalone"`
	SubscriptionURL    string               `json:"subscription_url,omitempty"`
	SelectedNode       string               `json:"selected_node,omitempty"`
	RouteMode          RouteMode            `json:"route_mode,omitempty"`
	Whitelist          []string             `json:"whitelist,omitempty"`
	Blacklist          []string             `json:"blacklist,omitempty"`
	StandaloneNodes    []*subscription.Node `json:"standalone_nodes,omitempty"`
}
```

- [ ] **Step 2: 新增增删方法**

```go
// config/store.go

func (c *Config) AddStandaloneNode(node *subscription.Node) {
	c.StandaloneNodes = append(c.StandaloneNodes, node)
}

func (c *Config) RemoveStandaloneNode(index int) bool {
	if index < 0 || index >= len(c.StandaloneNodes) {
		return false
	}
	c.StandaloneNodes = append(c.StandaloneNodes[:index], c.StandaloneNodes[index+1:]...)
	return true
}

func (c *Config) FindStandaloneNode(name string) *subscription.Node {
	for _, n := range c.StandaloneNodes {
		if n.Name == name {
			return n
		}
	}
	return nil
}
```

- [ ] **Step 3: 运行编译检查**

```bash
cd /mnt/software/xray-go
go build ./...
```

Expected: 编译通过，无错误

- [ ] **Step 4: 提交**

```bash
git add config/store.go
git commit -m "feat: add standalone nodes to config model"
```

---

### Task 3: 主菜单插入独立节点选项

**Files:**
- Modify: `main.go:194-239`

- [ ] **Step 1: 修改 `selectSubscription` 显示独立节点**

在订阅列表输出后、"+ Add new subscription" 前，插入：

```go
func selectSubscription(cfg *config.Config) *config.Subscription {
	if len(cfg.Subscriptions) == 0 && len(cfg.StandaloneNodes) == 0 {
		fmt.Println("No saved subscriptions or nodes.")
		// 提供选项让用户选择添加订阅或节点
		fmt.Println("\n  1. + Add new subscription")
		fmt.Println("  2. + Add manual node")
		fmt.Println("  3. Exit")
		fmt.Print("\nSelect option: ")
		var input string
		fmt.Scanln(&input)
		choice := 0
		fmt.Sscanf(input, "%d", &choice)
		switch choice {
		case 1:
			sub := promptAddSub(cfg)
			if sub != nil {
				return sub
			}
		case 2:
			promptAddStandaloneNode(cfg)
			return nil
		case 3:
			return nil
		}
		return nil
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
		
		// 显示独立节点选项
		standaloneCount := len(cfg.StandaloneNodes)
		if standaloneCount > 0 {
			marker := " "
			if cfg.LastUsedSub == "" && cfg.LastUsedStandalone {
				marker = "*"
			}
			fmt.Printf("\n  %sManual Nodes (%d nodes)\n", marker, standaloneCount)
		}
		
		offset := len(cfg.Subscriptions) + 1
		fmt.Printf("\n  %2d. + Add new subscription\n", offset)
		fmt.Printf("  %2d. + Add manual node\n", offset+1)
		fmt.Printf("  %2d. - Delete a subscription\n", offset+2)
		fmt.Printf("  %2d. - Delete a manual node\n", offset+3)
		fmt.Printf("  %2d. Exit\n", offset+4)

		fmt.Print("\nSelect option: ")
		var input string
		fmt.Scanln(&input)
		choice := 0
		fmt.Sscanf(input, "%d", &choice)

		if choice >= 1 && choice <= len(cfg.Subscriptions) {
			cfg.LastUsedStandalone = false
			return cfg.Subscriptions[choice-1]
		}
		
		// 选择 Manual Nodes
		if standaloneCount > 0 && choice == len(cfg.Subscriptions)+1 {
			cfg.LastUsedStandalone = true
			cfg.LastUsedSub = ""
			promptStandaloneMenu(cfg)
			// 如果用户从子菜单返回，继续显示主菜单
			continue
		}
		
		// 处理选项偏移
		actualChoice := choice
		if standaloneCount > 0 {
			actualChoice = choice - 1 // 因为 Manual Nodes 占了一个位置
		}
		
		if actualChoice == offset {
			sub := promptAddSub(cfg)
			if sub != nil {
				cfg.LastUsedStandalone = false
				return sub
			}
			continue
		}
		if actualChoice == offset+1 {
			promptAddStandaloneNode(cfg)
			continue
		}
		if actualChoice == offset+2 {
			promptDeleteSub(cfg)
			continue
		}
		if actualChoice == offset+3 {
			promptDeleteStandaloneNode(cfg)
			continue
		}
		if actualChoice == offset+4 {
			return nil
		}
		fmt.Println("Invalid choice")
	}
}
```

注意：上面的代码逻辑中 choice 偏移需要仔细处理。当 `standaloneCount > 0` 时，Manual Nodes 选项占据了 `len(cfg.Subscriptions)+1` 的位置。

更清晰的做法是使用显式的菜单项编号：

```go
func selectSubscription(cfg *config.Config) *config.Subscription {
	for {
		itemNum := 1
		
		if len(cfg.Subscriptions) > 0 {
			fmt.Println("\nSaved subscriptions:")
			for _, s := range cfg.Subscriptions {
				cached := len(s.Nodes)
				marker := " "
				if s.Name == cfg.LastUsedSub {
					marker = "*"
				}
				fmt.Printf("  %2d. %s%s (%s) [%d cached]\n", itemNum, marker, s.Name, s.URL, cached)
				itemNum++
			}
		}
		
		standaloneOption := 0
		if len(cfg.StandaloneNodes) > 0 {
			marker := " "
			if cfg.LastUsedSub == "" && cfg.LastUsedStandalone {
				marker = "*"
			}
			fmt.Printf("\n  %2d. %sManual Nodes (%d nodes)\n", itemNum, marker, len(cfg.StandaloneNodes))
			standaloneOption = itemNum
			itemNum++
		}
		
		addSubOption := itemNum
		fmt.Printf("\n  %2d. + Add new subscription\n", itemNum)
		itemNum++
		
		addNodeOption := itemNum
		fmt.Printf("  %2d. + Add manual node\n", itemNum)
		itemNum++
		
		delSubOption := itemNum
		fmt.Printf("  %2d. - Delete a subscription\n", itemNum)
		itemNum++
		
		delNodeOption := itemNum
		fmt.Printf("  %2d. - Delete a manual node\n", itemNum)
		itemNum++
		
		exitOption := itemNum
		fmt.Printf("  %2d. Exit\n", itemNum)

		fmt.Print("\nSelect option: ")
		var input string
		fmt.Scanln(&input)
		choice := 0
		fmt.Sscanf(input, "%d", &choice)

		for i, s := range cfg.Subscriptions {
			if choice == i+1 {
				cfg.LastUsedStandalone = false
				return s
			}
		}
		
		if standaloneOption > 0 && choice == standaloneOption {
			cfg.LastUsedStandalone = true
			cfg.LastUsedSub = ""
			promptStandaloneMenu(cfg)
			continue
		}
		
		if choice == addSubOption {
			sub := promptAddSub(cfg)
			if sub != nil {
				cfg.LastUsedStandalone = false
				return sub
			}
			continue
		}
		if choice == addNodeOption {
			promptAddStandaloneNode(cfg)
			continue
		}
		if choice == delSubOption {
			promptDeleteSub(cfg)
			continue
		}
		if choice == delNodeOption {
			promptDeleteStandaloneNode(cfg)
			continue
		}
		if choice == exitOption {
			return nil
		}
		fmt.Println("Invalid choice")
	}
}
```

- [ ] **Step 2: 编译检查**

```bash
cd /mnt/software/xray-go
go build ./...
```

Expected: 编译通过（可能有未定义函数警告）

- [ ] **Step 3: 提交**

```bash
git add main.go
git commit -m "feat: add standalone nodes menu option"
```

---

### Task 4: 实现独立节点子菜单

**Files:**
- Modify: `main.go`

- [ ] **Step 1: 实现 `promptStandaloneMenu`**

```go
func promptStandaloneMenu(cfg *config.Config) {
	for {
		fmt.Println("\nManual Nodes:")
		for i, n := range cfg.StandaloneNodes {
			fmt.Printf("  %2d. %s [%s]\n", i+1, n.Name, n.Protocol)
		}
		fmt.Printf("\n  %2d. + Add new node\n", len(cfg.StandaloneNodes)+1)
		fmt.Printf("  %2d. - Delete a node\n", len(cfg.StandaloneNodes)+2)
		fmt.Printf("  %2d. Back\n", len(cfg.StandaloneNodes)+3)

		fmt.Print("\nSelect option: ")
		var input string
		fmt.Scanln(&input)
		choice := 0
		fmt.Sscanf(input, "%d", &choice)

		if choice >= 1 && choice <= len(cfg.StandaloneNodes) {
			node := cfg.StandaloneNodes[choice-1]
			// 启动代理流程
			cfg.LastUsedStandalone = true
			cfg.LastUsedSub = ""
			cfg.Save()
			
			// 走和订阅节点一样的流程
			groups := region.GroupByRegion([]*subscription.Node{node})
			// 实际上这里应该对所有独立节点分组测速
			// 但为了和订阅一致，应该对所有独立节点测速选最优
			allNodes := cfg.StandaloneNodes
			groups = region.GroupByRegion(allNodes)
			selectedRegion := region.PromptRegion(groups)
			
			var targetNodes []*subscription.Node
			if selectedRegion == "" {
				targetNodes = allNodes
			} else {
				targetNodes = groups[selectedRegion]
				fmt.Printf("\nSelected region: %s (%d nodes)\n", selectedRegion, len(targetNodes))
			}
			
			fmt.Println("\nTesting latency...")
			bestNode, bestLatency, err := latency.FindBest(targetNodes)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				continue
			}
			fmt.Printf("Best node: %s (%v)\n", bestNode.Name, bestLatency)
			
			// 选择路由模式
			cfg.RouteMode = promptRouteMode(cfg.RouteMode)
			cfg.Save()
			
			socksPort := 16709
			httpPort := 16708
			runProxy(bestNode, socksPort, httpPort, cfg)
			return
		}
		if choice == len(cfg.StandaloneNodes)+1 {
			promptAddStandaloneNode(cfg)
			continue
		}
		if choice == len(cfg.StandaloneNodes)+2 {
			promptDeleteStandaloneNode(cfg)
			continue
		}
		if choice == len(cfg.StandaloneNodes)+3 {
			return
		}
		fmt.Println("Invalid choice")
	}
}
```

- [ ] **Step 2: 实现 `promptAddStandaloneNode`**

```go
func promptAddStandaloneNode(cfg *config.Config) {
	fmt.Print("Enter node share link (vmess:// / vless:// / trojan:// / ss:// / anytls://): ")
	var link string
	fmt.Scanln(&link)
	link = strings.TrimSpace(link)
	if link == "" {
		fmt.Println("Link cannot be empty")
		return
	}
	node, err := subscription.ParseNode(link)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse node: %v\n", err)
		return
	}
	if node.Name == "" {
		fmt.Print("Enter node name: ")
		var name string
		fmt.Scanln(&name)
		node.Name = strings.TrimSpace(name)
		if node.Name == "" {
			node.Name = fmt.Sprintf("node-%d", time.Now().Unix())
		}
	}
	cfg.AddStandaloneNode(node)
	cfg.Save()
	fmt.Printf("Added node '%s' [%s]\n", node.Name, node.Protocol)
}
```

- [ ] **Step 3: 实现 `promptDeleteStandaloneNode`**

```go
func promptDeleteStandaloneNode(cfg *config.Config) {
	if len(cfg.StandaloneNodes) == 0 {
		fmt.Println("No manual nodes to delete.")
		return
	}
	fmt.Println("\nSelect node to delete:")
	for i, n := range cfg.StandaloneNodes {
		fmt.Printf("  %2d. %s [%s]\n", i+1, n.Name, n.Protocol)
	}
	fmt.Print("Select: ")
	var input string
	fmt.Scanln(&input)
	choice := 0
	fmt.Sscanf(input, "%d", &choice)
	if choice < 1 || choice > len(cfg.StandaloneNodes) {
		fmt.Println("Invalid choice")
		return
	}
	node := cfg.StandaloneNodes[choice-1]
	fmt.Printf("Delete '%s'? (y/N): ", node.Name)
	fmt.Scanln(&input)
	if strings.ToLower(strings.TrimSpace(input)) == "y" {
		cfg.RemoveStandaloneNode(choice - 1)
		cfg.Save()
		fmt.Println("Deleted.")
	}
}
```

- [ ] **Step 4: 编译检查**

```bash
cd /mnt/software/xray-go
go build ./...
```

Expected: 编译通过

- [ ] **Step 5: 提交**

```bash
git add main.go
git commit -m "feat: implement standalone node submenu with add/delete"
```

---

### Task 5: 修改 `start` 模式支持独立节点

**Files:**
- Modify: `main.go:117-168`

- [ ] **Step 1: 修改 `startMode` 函数**

```go
func startMode(cfg *config.Config, httpPort int, updateFlag bool) {
	if len(cfg.Subscriptions) == 0 && len(cfg.StandaloneNodes) == 0 {
		fmt.Fprintln(os.Stderr, "No subscriptions or nodes configured. Run without 'start' first.")
		os.Exit(1)
	}
	
	// 如果上次使用的是独立节点
	if cfg.LastUsedStandalone && cfg.LastUsedSub == "" {
		if len(cfg.StandaloneNodes) == 0 {
			fmt.Fprintln(os.Stderr, "No standalone nodes available.")
			os.Exit(1)
		}
		
		var targetNodes []*subscription.Node
		if cfg.LastRegion != "" {
			groups := region.GroupByRegion(cfg.StandaloneNodes)
			targetNodes = groups[cfg.LastRegion]
			if len(targetNodes) == 0 {
				targetNodes = cfg.StandaloneNodes
			}
		} else {
			targetNodes = cfg.StandaloneNodes
		}
		
		fmt.Println("Testing latency...")
		bestNode, bestLatency, err := latency.FindBest(targetNodes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Best node: %s (%v)\n", bestNode.Name, bestLatency)
		
		socksPort := httpPort + 1
		runProxy(bestNode, socksPort, httpPort, cfg)
		return
	}
	
	// 原有订阅逻辑
	sub := cfg.FindSubscription(cfg.LastUsedSub)
	if sub == nil {
		if len(cfg.Subscriptions) > 0 {
			sub = cfg.Subscriptions[0]
			cfg.LastUsedSub = sub.Name
			cfg.LastUsedStandalone = false
		} else {
			fmt.Fprintln(os.Stderr, "No subscriptions available.")
			os.Exit(1)
		}
	}
	cfg.Save()

	nodes := sub.Nodes
	if len(nodes) == 0 || updateFlag {
		fmt.Println("Fetching subscription...")
		fetchedNodes, err := fetchSubOrFallback(sub, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		nodes = fetchedNodes
		sub.Nodes = nodes
		sub.LastFetched = time.Now()
		cfg.Save()
	} else {
		fmt.Printf("Using cached nodes (%d nodes)\n", len(nodes))
	}

	groups := region.GroupByRegion(nodes)
	var targetNodes []*subscription.Node
	if sub.LastRegion == "" {
		targetNodes = nodes
	} else {
		targetNodes = groups[sub.LastRegion]
		if len(targetNodes) == 0 {
			targetNodes = nodes
		}
	}

	fmt.Println("Testing latency...")
	bestNode, bestLatency, err := latency.FindBest(targetNodes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Best node: %s (%v)\n", bestNode.Name, bestLatency)
	sub.LastNode = bestNode.Name
	cfg.Save()

	socksPort := httpPort + 1
	runProxy(bestNode, socksPort, httpPort, cfg)
}
```

- [ ] **Step 2: 编译检查**

```bash
cd /mnt/software/xray-go
go build ./...
```

Expected: 编译通过

- [ ] **Step 3: 提交**

```bash
git add main.go
git commit -m "feat: support standalone nodes in start mode"
```

---

### Task 6: 保存独立节点的地区信息

**Files:**
- Modify: `config/store.go`
- Modify: `main.go`

- [ ] **Step 1: 在 Config 中新增 `LastStandaloneRegion`**

```go
type Config struct {
	// ... 其他字段 ...
	LastStandaloneRegion string `json:"last_standalone_region,omitempty"`
}
```

- [ ] **Step 2: 在 `promptStandaloneMenu` 中保存地区**

在调用 `region.PromptRegion` 后，保存选中的地区：

```go
selectedRegion := region.PromptRegion(groups)
cfg.LastStandaloneRegion = selectedRegion
cfg.Save()
```

- [ ] **Step 3: 在 `startMode` 中使用 `LastStandaloneRegion`**

```go
if cfg.LastStandaloneRegion != "" {
	groups := region.GroupByRegion(cfg.StandaloneNodes)
	targetNodes = groups[cfg.LastStandaloneRegion]
	if len(targetNodes) == 0 {
		targetNodes = cfg.StandaloneNodes
	}
} else {
	targetNodes = cfg.StandaloneNodes
}
```

- [ ] **Step 4: 编译和提交**

```bash
cd /mnt/software/xray-go
go build ./...
git add config/store.go main.go
git commit -m "feat: persist standalone node region selection"
```

---

## 验证

1. **编译通过**：`go build ./...` 无错误
2. **测试通过**：`go test ./subscription -v` 全部 PASS
3. **手动测试**：
   - 运行 `./xray-go`，检查主菜单显示 "Manual Nodes" 选项
   - 选择 "+ Add manual node"，粘贴 vless:// 链接，确认添加成功
   - 选择 "Manual Nodes"，确认子菜单显示已添加的节点
   - 选择节点，确认能正常测速并启动代理
   - 按 Ctrl+C 退出，再次运行 `./xray-go start`，确认能自动使用上次选择的独立节点
   - 选择 "- Delete a manual node"，确认删除功能正常

---

## Self-Review

**Spec coverage:**
- [x] 数据模型：`StandaloneNodes` 字段 → Task 2
- [x] 交互设计：主菜单插入选项 → Task 3
- [x] 独立节点子菜单 → Task 4
- [x] 解析逻辑：`ParseNode` 公开 → Task 1
- [x] 代理启动流程 → Task 4
- [x] 无交互启动模式 → Task 5
- [x] 地区信息持久化 → Task 6

**Placeholder scan:** 无 TBD/TODO/"implement later"

**Type consistency:**
- `ParseNode` 在 Task 1 定义，Task 4 中使用 → 一致
- `AddStandaloneNode`/`RemoveStandaloneNode` 在 Task 2 定义，Task 4 中使用 → 一致
- `LastUsedStandalone` 在 Task 2 定义，Task 3/5 中使用 → 一致
