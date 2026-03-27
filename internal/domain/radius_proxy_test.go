package domain

import "testing"

func TestProxyServer_ValidConfiguration_ShouldPass(t *testing.T) {
	server := &RadiusProxyServer{
		Name:       "Primary Proxy",
		Host:       "192.168.1.10",
		AuthPort:   1812,
		AcctPort:   1813,
		Secret:     "sharedsecret",
		Status:     "enabled",
		MaxConns:   100,
		TimeoutSec: 5,
	}

	err := server.Validate()
	if err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}

func TestProxyRealm_ValidRouting_ShouldPass(t *testing.T) {
	realm := &RadiusProxyRealm{
		Realm:         "example.com",
		ProxyServers:  []int64{1, 2}, // Server IDs
		FallbackOrder: 1,
		Status:        "enabled",
	}

	err := realm.Validate()
	if err != nil {
		t.Fatalf("expected valid realm, got error: %v", err)
	}
}
