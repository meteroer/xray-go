package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
	urlFlag := flag.String("url", "", "subscription URL (overrides saved config)")
	portFlag := flag.Int("port", 16708, "local proxy port")
	updateFlag := flag.Bool("update", false, "force re-fetch subscription and re-test latency")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	subURL := *urlFlag
	if subURL == "" {
		subURL = cfg.SubscriptionURL
	}
	if subURL == "" || *updateFlag {
		url, err := config.PromptSubscriptionURL()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		subURL = url
	}

	cfg.SubscriptionURL = subURL
	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save config: %v\n", err)
	}

	fmt.Println("Fetching subscription...")
	data, err := subscription.Fetch(subURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching subscription: %v\n", err)
		os.Exit(1)
	}

	nodes, err := subscription.Parse(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing subscription: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Found %d nodes\n", len(nodes))

	groups := region.GroupByRegion(nodes)
	selectedRegion := region.PromptRegion(groups)

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
		os.Exit(1)
	}
	fmt.Printf("Best node: %s (%v)\n", bestNode.Name, bestLatency)

	cfg.SelectedNode = bestNode.Name
	config.Save(cfg)

	httpPort := *portFlag
	socksPort := httpPort + 1
	fmt.Printf("Starting proxy on 127.0.0.1:%d (HTTP) and 127.0.0.1:%d (SOCKS5)...\n", httpPort, socksPort)

	var srv ProxyServer
	if bestNode.Protocol == "anytls" {
		srv, err = singbox.Start(bestNode, socksPort, httpPort)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error starting sing-box proxy: %v\n", err)
			os.Exit(1)
		}
	} else {
		srv, err = xrayproxy.Start(bestNode, socksPort, httpPort)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error starting xray proxy: %v\n", err)
			os.Exit(1)
		}
	}
	fmt.Printf("Proxy running at 127.0.0.1:%d (HTTP) and 127.0.0.1:%d (SOCKS5)\n", httpPort, socksPort)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	fmt.Println("\nShutting down...")
	srv.Stop()
	fmt.Println("Done.")
}