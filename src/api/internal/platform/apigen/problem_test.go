package apigen_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
)

// Problem is the RFC 9457 body (ADR-0004 point 4): known members are typed and
// extension members are preserved.
func TestProblemRoundTripKeepsExtensionMembers(t *testing.T) {
	in := `{"type":"urn:226os:problem:NOT_FOUND","title":"Not found","status":404,"code":"NOT_FOUND","request_id":"req-1","detail":"d","instance":"/x","violations":[{"field":"a","code":"REQUIRED"}],"extra":{"k":1}}`
	var p apigen.Problem
	if err := json.Unmarshal([]byte(in), &p); err != nil {
		t.Fatal(err)
	}
	if p.Status != 404 || p.Code != apigen.ErrorCode("NOT_FOUND") || p.RequestId != "req-1" {
		t.Fatalf("typed members wrong: %+v", p)
	}
	if v, ok := p.Get("extra"); !ok || v == nil {
		t.Fatalf("extension member lost: %v %v", v, ok)
	}
	if _, ok := p.Get("missing"); ok {
		t.Fatal("Get reported a member that was never set")
	}
	out, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var a, b map[string]any
	if json.Unmarshal([]byte(in), &a) != nil || json.Unmarshal(out, &b) != nil {
		t.Fatal("unmarshal for comparison")
	}
	if len(a) != len(b) {
		t.Fatalf("round trip changed members: %s != %s", in, out)
	}
	p.Set("another", "v")
	out, _ = json.Marshal(p)
	if !strings.Contains(string(out), `"another":"v"`) {
		t.Fatalf("Set not serialized: %s", out)
	}
}

func TestProblemRejectsBadInput(t *testing.T) {
	cases := map[string]string{
		"not an object": `[]`,
		"code":          `{"code":1}`,
		"detail":        `{"detail":1}`,
		"instance":      `{"instance":1}`,
		"request_id":    `{"request_id":1}`,
		"status":        `{"status":"x"}`,
		"title":         `{"title":1}`,
		"type":          `{"type":1}`,
		"violations":    `{"violations":1}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			var p apigen.Problem
			if err := json.Unmarshal([]byte(body), &p); err == nil {
				t.Fatalf("expected an error for %s", body)
			}
		})
	}
}

func TestProblemMarshalReportsUnserializableExtension(t *testing.T) {
	var p apigen.Problem
	p.Set("bad", make(chan int))
	if _, err := json.Marshal(p); err == nil {
		t.Fatal("expected an error for an unserializable extension member")
	}
}
