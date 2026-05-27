package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"xray-go/subscription"
)

type Subscription struct {
	Name        string               `json:"name"`
	URL         string               `json:"url"`
	Nodes       []*subscription.Node `json:"nodes,omitempty"`
	LastNode    string               `json:"last_node"`
	LastRegion  string               `json:"last_region,omitempty"`
	LastFetched time.Time            `json:"last_fetched"`
}

type RouteMode string

const (
	RouteModeGlobal    RouteMode = "global"    // 全部走代理
	RouteModeWhitelist RouteMode = "whitelist" // 白名单走代理，其他直连
	RouteModeBlacklist RouteMode = "blacklist" // 黑名单直连，其他走代理
)

type Config struct {
	Subscriptions   []*Subscription `json:"subscriptions"`
	LastUsedSub     string          `json:"last_used_subscription"`
	SubscriptionURL string          `json:"subscription_url,omitempty"`
	SelectedNode    string          `json:"selected_node,omitempty"`
	RouteMode       RouteMode       `json:"route_mode,omitempty"`
	Whitelist       []string        `json:"whitelist,omitempty"`
	Blacklist       []string        `json:"blacklist,omitempty"`
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".xray-go")
	return dir, os.MkdirAll(dir, 0755)
}

func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func defaultWhitelist() []string {
	return []string{
		"geosite:google",
		"geosite:youtube",
		"geosite:github",
		"geosite:twitter",
		"geosite:telegram",
		"geosite:facebook",
		"geosite:instagram",
	}
}

func defaultBlacklist() []string {
	return []string{
		"geosite:cn",
		"geoip:cn",
		"geoip:private",
	}
}

func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				RouteMode: RouteModeGlobal,
				Whitelist: defaultWhitelist(),
				Blacklist: defaultBlacklist(),
			}, nil
		}
		return nil, err
	}
	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	// Set defaults if empty
	if cfg.RouteMode == "" {
		cfg.RouteMode = RouteModeGlobal
	}
	if len(cfg.Whitelist) == 0 {
		cfg.Whitelist = defaultWhitelist()
	}
	if len(cfg.Blacklist) == 0 {
		cfg.Blacklist = defaultBlacklist()
	}
	// Migration from old single subscription format
	if cfg.SubscriptionURL != "" && len(cfg.Subscriptions) == 0 {
		sub := &Subscription{
			Name:     "default",
			URL:      cfg.SubscriptionURL,
			LastNode: cfg.SelectedNode,
		}
		cfg.Subscriptions = append(cfg.Subscriptions, sub)
		cfg.LastUsedSub = "default"
		cfg.SubscriptionURL = ""
		cfg.SelectedNode = ""
		cfg.Save()
	}
	return cfg, nil
}

func (c *Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}
	cfg := *c
	cfg.SubscriptionURL = ""
	cfg.SelectedNode = ""
	data, err := json.MarshalIndent(&cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (c *Config) AddSubscription(name, url string) *Subscription {
	sub := &Subscription{
		Name: name,
		URL:  url,
	}
	c.Subscriptions = append(c.Subscriptions, sub)
	return sub
}

func (c *Config) RemoveSubscription(name string) bool {
	for i, s := range c.Subscriptions {
		if s.Name == name {
			c.Subscriptions = append(c.Subscriptions[:i], c.Subscriptions[i+1:]...)
			if c.LastUsedSub == name {
				c.LastUsedSub = ""
			}
			return true
		}
	}
	return false
}

func (c *Config) FindSubscription(name string) *Subscription {
	for _, s := range c.Subscriptions {
		if s.Name == name {
			return s
		}
	}
	return nil
}

func (c *Config) FindFallbackSub(excludeName string) *Subscription {
	if c.LastUsedSub != "" && c.LastUsedSub != excludeName {
		sub := c.FindSubscription(c.LastUsedSub)
		if sub != nil && sub.LastNode != "" && len(sub.Nodes) > 0 {
			return sub
		}
	}
	for _, s := range c.Subscriptions {
		if s.Name == excludeName {
			continue
		}
		if s.LastNode != "" && len(s.Nodes) > 0 {
			return s
		}
	}
	return nil
}

func (s *Subscription) FindNode(name string) *subscription.Node {
	for _, n := range s.Nodes {
		if n.Name == name {
			return n
		}
	}
	return nil
}

func PromptSubscriptionURL() (string, error) {
	fmt.Print("Enter subscription URL: ")
	var url string
	_, err := fmt.Scanln(&url)
	if err != nil {
		return "", err
	}
	if url == "" {
		return "", fmt.Errorf("subscription URL cannot be empty")
	}
	return url, nil
}

func Save(cfg *Config) error {
	return cfg.Save()
}
