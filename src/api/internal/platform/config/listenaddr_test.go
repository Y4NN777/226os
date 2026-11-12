package config_test

import (
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/config"
)

func TestListenAddrDefaultsAndOverrides(t *testing.T) {
	if got := config.ListenAddr(func(string) string { return "" }); got != ":8080" {
		t.Fatalf("default = %q", got)
	}
	if got := config.ListenAddr(func(k string) string {
		if k == "API_ADDR" {
			return "127.0.0.1:9000"
		}
		return ""
	}); got != "127.0.0.1:9000" {
		t.Fatalf("override = %q", got)
	}
}
