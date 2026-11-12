package config_test

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/config"
)

// aiKey is built at run time so no key-shaped literal sits in the source.
func aiKey() string { return fmt.Sprintf("%s%s", "AIza", strings.Repeat("Z", 35)) }

func aiEnv(extra map[string]string) map[string]string {
	m := map[string]string{"AI_ENABLED": "true", "AI_API_KEY": aiKey(), "AI_MODEL": "gemma-test"}
	for k, v := range extra {
		m[k] = v
	}
	return m
}

func TestLoadAIIsNilUnlessEnabled(t *testing.T) {
	for _, m := range []map[string]string{
		{},
		{"AI_ENABLED": "false", "AI_API_KEY": aiKey(), "AI_MODEL": "gemma-test"},
		{"AI_API_KEY": aiKey(), "AI_MODEL": "gemma-test"},
	} {
		ai, err := config.LoadAI(env(m))
		if ai != nil || err != nil {
			t.Fatalf("env %v: ai = %v err = %v", m, ai, err)
		}
	}
}

func TestLoadAIValid(t *testing.T) {
	ai, err := config.LoadAI(env(aiEnv(nil)))
	if err != nil {
		t.Fatal(err)
	}
	want := config.AI{Enabled: true, APIKey: aiKey(), Model: "gemma-test", Timeout: 30 * time.Second}
	if *ai != want {
		t.Fatalf("ai = %+v", *ai)
	}
	for in, d := range map[string]time.Duration{"1s": time.Second, "120s": 120 * time.Second, "2m": 2 * time.Minute, "45s": 45 * time.Second} {
		ai, err := config.LoadAI(env(aiEnv(map[string]string{"AI_TIMEOUT": in})))
		if err != nil || ai.Timeout != d {
			t.Fatalf("AI_TIMEOUT=%s: %+v %v", in, ai, err)
		}
	}
}

func TestLoadAIRejects(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want string // the variable the error must name
	}{
		{"bad switch", aiEnv(map[string]string{"AI_ENABLED": "yes"}), "AI_ENABLED"},
		{"no key", aiEnv(map[string]string{"AI_API_KEY": ""}), "AI_API_KEY"},
		{"key with space", aiEnv(map[string]string{"AI_API_KEY": "ab cd"}), "AI_API_KEY"},
		{"key with newline", aiEnv(map[string]string{"AI_API_KEY": "abcd\n"}), "AI_API_KEY"},
		{"key with DEL", aiEnv(map[string]string{"AI_API_KEY": "abcd\x7f"}), "AI_API_KEY"},
		{"key non-ASCII", aiEnv(map[string]string{"AI_API_KEY": "abcdé"}), "AI_API_KEY"},
		{"key too long", aiEnv(map[string]string{"AI_API_KEY": strings.Repeat("a", 513)}), "AI_API_KEY"},
		{"no model", aiEnv(map[string]string{"AI_MODEL": ""}), "AI_MODEL"},
		{"model with slash", aiEnv(map[string]string{"AI_MODEL": "a/b"}), "AI_MODEL"},
		{"model with query", aiEnv(map[string]string{"AI_MODEL": "a?x=1"}), "AI_MODEL"},
		{"timeout not a duration", aiEnv(map[string]string{"AI_TIMEOUT": "soon"}), "AI_TIMEOUT"},
		{"timeout too short", aiEnv(map[string]string{"AI_TIMEOUT": "999ms"}), "AI_TIMEOUT"},
		{"timeout too long", aiEnv(map[string]string{"AI_TIMEOUT": "121s"}), "AI_TIMEOUT"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ai, err := config.LoadAI(env(tc.env))
			if ai != nil || err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("ai = %v err = %v, want an error naming %s", ai, err, tc.want)
			}
			for _, secret := range []string{aiKey(), tc.env["AI_API_KEY"]} {
				if secret != "" && strings.Contains(err.Error(), secret) {
					t.Fatalf("error echoes the key: %v", err)
				}
			}
		})
	}
}

func TestAIIsRedacted(t *testing.T) {
	ai, err := config.LoadAI(env(aiEnv(nil)))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(env(base(aiEnv(nil))))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	slog.New(slog.NewJSONHandler(&buf, nil)).Info("x", "ai", *ai, "ai_ptr", ai, "config", cfg)
	outs := map[string]string{"slog": buf.String()}
	for name, v := range map[string]any{"ai": *ai, "ai ptr": ai, "config": cfg} {
		outs[name+" %v"] = fmt.Sprintf("%v", v)
		outs[name+" %+v"] = fmt.Sprintf("%+v", v)
		outs[name+" %#v"] = fmt.Sprintf("%#v", v)
	}
	for name, out := range outs {
		if strings.Contains(out, aiKey()) || strings.Contains(out, "ZZZZ") {
			t.Errorf("%s leaks the key: %s", name, out)
		}
		if !strings.Contains(out, "gemma-test") || !strings.Contains(out, "redacted") {
			t.Errorf("%s lost its content: %s", name, out)
		}
	}
}

func TestLoadIncludesAI(t *testing.T) {
	cfg, err := config.Load(env(base(aiEnv(nil))))
	if err != nil || cfg.AI == nil || cfg.AI.Model != "gemma-test" {
		t.Fatalf("cfg = %+v err = %v", cfg, err)
	}
	cfg, err = config.Load(env(base(nil)))
	if err != nil || cfg.AI != nil {
		t.Fatalf("cfg.AI = %v err = %v", cfg.AI, err)
	}
	if s := cfg.String(); strings.Contains(s, "AI:") {
		t.Fatalf("String without AI mentions it: %s", s)
	}
	var buf bytes.Buffer
	slog.New(slog.NewJSONHandler(&buf, nil)).Info("x", "config", cfg)
	if !strings.Contains(buf.String(), `"AI":"none"`) {
		t.Fatalf("slog without AI: %s", buf.String())
	}
	if _, err := config.Load(env(base(aiEnv(map[string]string{"AI_MODEL": ""})))); err == nil || !strings.Contains(err.Error(), "AI_MODEL") {
		t.Fatalf("Load must surface the AI error, got %v", err)
	}
}
