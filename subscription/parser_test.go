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
