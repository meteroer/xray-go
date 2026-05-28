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

	// "start" subcommand: non-interactive, use last saved config
	args := flag.Args()
	if len(args) > 0 && args[0] == "start" {
		startMode(cfg, *portFlag, *updateFlag)
		return
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
		sub.LastRegion = selectedRegion

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

		// 选择路由模式
		cfg.RouteMode = promptRouteMode(cfg.RouteMode)
		cfg.Save()

		socksPort := *portFlag + 1
		runProxy(bestNode, socksPort, *portFlag, cfg)
		return
	}
}

func startMode(cfg *config.Config, httpPort int, updateFlag bool) {
	if len(cfg.Subscriptions) == 0 {
		fmt.Fprintln(os.Stderr, "No subscriptions configured. Run without 'start' first.")
		os.Exit(1)
	}
	sub := cfg.FindSubscription(cfg.LastUsedSub)
	if sub == nil {
		sub = cfg.Subscriptions[0]
		cfg.LastUsedSub = sub.Name
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

func runProxy(node *subscription.Node, socksPort, httpPort int, cfg *config.Config) {
	fmt.Printf("Starting proxy on 0.0.0.0:%d (HTTP) and 0.0.0.0:%d (SOCKS5) [%s mode]...\n", httpPort, socksPort, cfg.RouteMode)

	var srv ProxyServer
	var err error
	if node.Protocol == "anytls" {
		srv, err = singbox.Start(node, socksPort, httpPort, cfg.RouteMode, cfg.Whitelist, cfg.Blacklist)
	} else {
		srv, err = xrayproxy.Start(node, socksPort, httpPort, cfg.RouteMode, cfg.Whitelist, cfg.Blacklist)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting proxy: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Proxy running at 0.0.0.0:%d (HTTP) and 0.0.0.0:%d (SOCKS5) [%s mode]\n", httpPort, socksPort, cfg.RouteMode)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	fmt.Println("\nShutting down...")
	srv.Stop()
	fmt.Println("Done.")
}

func selectSubscription(cfg *config.Config) *config.Subscription {
	for {
		itemNum := 1
		optionMap := make(map[int]func() *config.Subscription)

		if len(cfg.Subscriptions) > 0 {
			fmt.Println("\nSaved subscriptions:")
			for i, s := range cfg.Subscriptions {
				cached := len(s.Nodes)
				marker := " "
				if s.Name == cfg.LastUsedSub {
					marker = "*"
				}
				fmt.Printf("  %2d. %s%s (%s) [%d cached]\n", itemNum, marker, s.Name, s.URL, cached)
				idx := i
				optionMap[itemNum] = func() *config.Subscription {
					cfg.LastUsedStandalone = false
					return cfg.Subscriptions[idx]
				}
				itemNum++
			}
		}

		if len(cfg.StandaloneNodes) > 0 {
			marker := " "
			if cfg.LastUsedSub == "" && cfg.LastUsedStandalone {
				marker = "*"
			}
			fmt.Printf("\n  %2d. %sManual Nodes (%d nodes)\n", itemNum, marker, len(cfg.StandaloneNodes))
			optionMap[itemNum] = func() *config.Subscription {
				cfg.LastUsedStandalone = true
				cfg.LastUsedSub = ""
				promptStandaloneMenu(cfg)
				return nil
			}
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

		if action, ok := optionMap[choice]; ok {
			result := action()
			if result != nil {
				return result
			}
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
		srv, err = singbox.Start(fallbackNode, socksPort, httpPort, config.RouteModeGlobal, nil, nil)
	} else {
		srv, err = xrayproxy.Start(fallbackNode, socksPort, httpPort, config.RouteModeGlobal, nil, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("start fallback proxy: %w", err)
	}
	defer srv.Stop()
	time.Sleep(200 * time.Millisecond)

	proxyAddr := fmt.Sprintf("0.0.0.0:%d", socksPort)
	data, err = subscription.FetchWithProxy(sub.URL, proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("fallback fetch failed: %w", err)
	}
	return subscription.Parse(data)
}

func promptRouteMode(current config.RouteMode) config.RouteMode {
	fmt.Println("\nSelect route mode:")
	fmt.Println("  1. Global (所有流量走代理)")
	fmt.Println("  2. Whitelist (仅白名单走代理，其他直连)")
	fmt.Println("  3. Blacklist (仅黑名单直连，其他走代理)")
	fmt.Print("Select (默认使用上次配置): ")

	var input string
	fmt.Scanln(&input)
	choice := 0
	fmt.Sscanf(input, "%d", &choice)

	switch choice {
	case 1:
		return config.RouteModeGlobal
	case 2:
		return config.RouteModeWhitelist
	case 3:
		return config.RouteModeBlacklist
	default:
		if current != "" {
			fmt.Printf("Using saved mode: %s\n", current)
			return current
		}
		return config.RouteModeGlobal
	}
}

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
			cfg.LastUsedStandalone = true
			cfg.LastUsedSub = ""
			cfg.Save()

			allNodes := cfg.StandaloneNodes
			groups := region.GroupByRegion(allNodes)
			selectedRegion := region.PromptRegion(groups)
			cfg.LastStandaloneRegion = selectedRegion
			cfg.Save()

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
