package latency

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"xray-cli/subscription"
	"xray-cli/xrayproxy"
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
	socksPort, err := xrayproxy.GetFreePort()
	if err != nil {
		return 0, fmt.Errorf("get free port: %w", err)
	}
	httpPort, err := xrayproxy.GetFreePort()
	if err != nil {
		return 0, fmt.Errorf("get free port: %w", err)
	}

	srv, err := xrayproxy.Start(node, socksPort, httpPort)
	if err != nil {
		return 0, fmt.Errorf("start proxy: %w", err)
	}
	defer srv.Stop()

	time.Sleep(200 * time.Millisecond)

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return (&net.Dialer{Timeout: 3 * time.Second}).DialContext(ctx, "tcp", fmt.Sprintf("127.0.0.1:%d", httpPort))
			},
		},
	}

	start := time.Now()
	resp, err := client.Get("http://www.gstatic.com/generate_204")
	if err != nil {
		return 0, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()
	elapsed := time.Since(start)

	if resp.StatusCode != 204 && resp.StatusCode != 200 {
		return 0, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return elapsed, nil
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
