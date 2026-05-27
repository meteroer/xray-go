package latency

import (
	"context"
	"fmt"
	"sync"
	"time"

	"xray-cli/subscription"
	"xray-cli/singbox"
	"xray-cli/xrayproxy"

	"golang.org/x/net/proxy"
)

type Result struct {
	Node    *subscription.Node
	Latency time.Duration
	Err     error
}

func TestAll(nodes []*subscription.Node, maxConcurrent int) []*Result {
	if maxConcurrent <= 0 {
		maxConcurrent = 5
	}
	results := make([]*Result, len(nodes))
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrent)

	for i, node := range nodes {
		wg.Add(1)
		go func(idx int, n *subscription.Node) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			latency, err := testNode(n)
			results[idx] = &Result{Node: n, Latency: latency, Err: err}
		}(i, node)
	}
	wg.Wait()
	return results
}

func testNode(node *subscription.Node) (time.Duration, error) {
	socksPort, err := getFreePort()
	if err != nil {
		return 0, fmt.Errorf("get free port: %w", err)
	}
	httpPort, err := getFreePort()
	if err != nil {
		return 0, fmt.Errorf("get free port: %w", err)
	}

	if node.Protocol == "anytls" {
		srv, err := singbox.Start(node, socksPort, httpPort)
		if err != nil {
			return 0, fmt.Errorf("start sing-box: %w", err)
		}
		defer srv.Stop()
	} else {
		srv, err := xrayproxy.Start(node, socksPort, httpPort)
		if err != nil {
			return 0, fmt.Errorf("start proxy: %w", err)
		}
		defer srv.Stop()
	}

	time.Sleep(200 * time.Millisecond)

	dialer, err := proxy.SOCKS5("tcp", fmt.Sprintf("127.0.0.1:%d", socksPort), nil, proxy.Direct)
	if err != nil {
		return 0, fmt.Errorf("socks5 dialer: %w", err)
	}
	contextDialer, ok := dialer.(proxy.ContextDialer)
	if !ok {
		return 0, fmt.Errorf("socks5 dialer does not support DialContext")
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := contextDialer.DialContext(ctx, "tcp", "cp.cloudflare.com:80")
	if err != nil {
		return 0, fmt.Errorf("dial: %w", err)
	}
	conn.Close()
	elapsed := time.Since(start)
	return elapsed, nil
}

func getFreePort() (int, error) {
	return xrayproxy.GetFreePort()
}

func FindBest(nodes []*subscription.Node) (*subscription.Node, time.Duration, error) {
	results := TestAll(nodes, 5)
	var bestNode *subscription.Node
	var bestLatency time.Duration
	var lastErr error
	for _, r := range results {
		if r.Err != nil {
			lastErr = r.Err
			fmt.Printf("  x %s: %v\n", r.Node.Name, r.Err)
			continue
		}
		fmt.Printf("  + %s: %v\n", r.Node.Name, r.Latency)
		if bestNode == nil || r.Latency < bestLatency {
			bestNode = r.Node
			bestLatency = r.Latency
		}
	}
	if bestNode == nil {
		return nil, 0, fmt.Errorf("all nodes unreachable: %v", lastErr)
	}
	return bestNode, bestLatency, nil
}