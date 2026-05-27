package subscription

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"crypto/tls"
	"golang.org/x/net/proxy"
)

func Fetch(url string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subscription: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("subscription returned status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read subscription body: %w", err)
	}
	return data, nil
}

func FetchWithProxy(url, socks5ProxyAddr string) ([]byte, error) {
	dialer, err := proxy.SOCKS5("tcp", socks5ProxyAddr, nil, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("socks5 dialer: %w", err)
	}
	ctxDialer, ok := dialer.(proxy.ContextDialer)
	if !ok {
		return nil, fmt.Errorf("socks5 dialer does not support DialContext")
	}
	transport := &http.Transport{
		DialContext:    ctxDialer.DialContext,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Timeout: 30 * time.Second, Transport: transport}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subscription via proxy: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("subscription returned status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read subscription body: %w", err)
	}
	return data, nil
}
