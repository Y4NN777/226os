package llm

import (
	"net/http"
	"testing"
	"time"
)

func TestDefaultsOfTheAdapter(t *testing.T) {
	a := NewAIStudio(AIStudioConfig{APIKey: "k", Model: "m"})
	if a.cfg.Timeout != 30*time.Second {
		t.Fatalf("default timeout = %v", a.cfg.Timeout)
	}
	if a.base != "https://generativelanguage.googleapis.com" {
		t.Fatalf("default base = %q", a.base)
	}
	b := NewAIStudio(AIStudioConfig{Timeout: -1}, WithBaseURL("http://127.0.0.1:1///"))
	if b.cfg.Timeout != 30*time.Second || b.base != "http://127.0.0.1:1" {
		t.Fatalf("timeout = %v base = %q", b.cfg.Timeout, b.base)
	}
}

func TestTransportKeepsTLSVerificationAndRefusesRedirects(t *testing.T) {
	hc := newHTTPClient(2 * time.Second)
	tr, ok := hc.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport is %T", hc.Transport)
	}
	if tr.TLSClientConfig != nil && tr.TLSClientConfig.InsecureSkipVerify {
		t.Fatal("TLS verification is disabled")
	}
	if tr.ResponseHeaderTimeout != 3*time.Second || tr.DialContext == nil {
		t.Fatalf("timeouts: header %v", tr.ResponseHeaderTimeout)
	}
	if err := hc.CheckRedirect(nil, nil); err == nil {
		t.Fatal("redirects must be refused")
	}
}
