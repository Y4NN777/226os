package app

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/llm"
	"github.com/Y4NN777/226os/src/api/internal/platform/llm/aistest"
)

// aiFixtureKey is built at run time so no key-shaped literal sits in the source.
func aiFixtureKey() string { return fmt.Sprintf("%s%s", "AIza", strings.Repeat("K", 35)) }

func aiCheckEnv(extra map[string]string) func(string) string {
	m := map[string]string{"AI_ENABLED": "true", "AI_API_KEY": aiFixtureKey(), "AI_MODEL": "gemma-test", "AI_TIMEOUT": "5s"}
	for k, v := range extra {
		m[k] = v
	}
	return func(k string) string { return m[k] }
}

// runAICheck runs the command against an aistest server (the production URL is never used).
func runAICheck(s *aistest.Server, env func(string) string, args ...string) (code int, out, errOut string) {
	var o, e bytes.Buffer
	code = aiCheck(context.Background(), args, env, &o, &e, llm.WithBaseURL(s.URL()))
	return code, o.String(), e.String()
}

func assertNoKey(t *testing.T, outs ...string) {
	t.Helper()
	for _, o := range outs {
		if strings.Contains(o, aiFixtureKey()) || strings.Contains(o, "KKKK") {
			t.Fatalf("output leaks the key: %q", o)
		}
	}
}

func TestAICheckModelsListsOnlyGemmaModels(t *testing.T) {
	s := aistest.NewServer(t, aistest.WithModels(
		aistest.Behaviour{Body: aistest.ModelsBody("p2", "gemini-pro", "Gemma-4-test")},
		aistest.Behaviour{Body: aistest.ModelsBody("", "text-embedding", "gemma-3-mini")},
	))
	code, out, errOut := runAICheck(s, aiCheckEnv(nil), "models")
	want := "Gemma-4-test | Display Gemma-4-test | generateContent,countTokens | 131072/8192\n" +
		"gemma-3-mini | Display gemma-3-mini | generateContent,countTokens | 131072/8192\n"
	if code != 0 || out != want || errOut != "" {
		t.Fatalf("code = %d out = %q err = %q", code, out, errOut)
	}
	assertNoKey(t, out, errOut)
}

func TestAICheckModelsWithoutGemma(t *testing.T) {
	for name, body := range map[string]string{"none match": aistest.ModelsBody("", "gemini-a", "gemini-b"), "empty list": `{}`} {
		t.Run(name, func(t *testing.T) {
			s := aistest.NewServer(t, aistest.WithModels(aistest.Behaviour{Body: body}))
			code, out, errOut := runAICheck(s, aiCheckEnv(nil), "models")
			n := 2
			if name == "empty list" {
				n = 0
			}
			if code != 0 || errOut != "" || out != fmt.Sprintf("no model name contains \"gemma\" (%d models seen)\n", n) {
				t.Fatalf("code = %d out = %q err = %q", code, out, errOut)
			}
		})
	}
}

var pingLine = regexp.MustCompile(`^(en|fr) \| \d+ ms \| (\d+) \| (\d+) \| (.*)\n$`)

func TestAICheckPingBothLanguagesInSequence(t *testing.T) {
	long := strings.Repeat("é", 200) + "\nsecond line"
	s := aistest.NewServer(t, aistest.WithGenerate(
		aistest.Behaviour{Body: aistest.AnswerBody("The page was not found.\nMore text.", 9, 5)},
		aistest.Behaviour{Body: aistest.AnswerBody(long, 12, 40)},
	))
	code, out, errOut := runAICheck(s, aiCheckEnv(nil), "ping")
	lines := strings.SplitAfter(out, "\n")
	lines = lines[:len(lines)-1] // SplitAfter leaves an empty tail after the final newline
	if code != 0 || errOut != "" || len(lines) != 2 {
		t.Fatalf("code = %d out = %q err = %q", code, out, errOut)
	}
	first := pingLine.FindStringSubmatch(lines[0])
	second := pingLine.FindStringSubmatch(lines[1])
	if first == nil || first[1] != "en" || first[2] != "9" || first[3] != "5" || first[4] != "The page was not found. More text." {
		t.Fatalf("first line = %q", lines[0])
	}
	if second == nil || second[1] != "fr" || second[2] != "12" || second[3] != "40" || len([]rune(second[4])) != 160 || strings.Contains(second[4], "\n") {
		t.Fatalf("second line = %q", lines[1])
	}
	reqs := s.Requests()
	if len(reqs) != 2 || !strings.Contains(reqs[0].Body, "HTTP status code") || !strings.Contains(reqs[1].Body, "code d'état HTTP") {
		t.Fatalf("requests = %+v", reqs)
	}
	assertNoKey(t, out, errOut)
}

func TestAICheckPingSingleLanguage(t *testing.T) {
	for _, lang := range []string{"en", "fr"} {
		s := aistest.NewServer(t)
		code, out, _ := runAICheck(s, aiCheckEnv(nil), "ping", lang)
		if m := pingLine.FindStringSubmatch(out); code != 0 || m == nil || m[1] != lang || len(s.Requests()) != 1 {
			t.Fatalf("%s: code = %d out = %q", lang, code, out)
		}
	}
}

func TestAICheckLimitsSendsThreeProbesAndNeverRetries(t *testing.T) {
	s := aistest.NewServer(t, aistest.WithGenerate(
		aistest.Behaviour{Body: aistest.AnswerBody("ok", 1, 1)},
		aistest.Behaviour{Status: 429},
		aistest.Behaviour{Status: 503},
	))
	code, out, errOut := runAICheck(s, aiCheckEnv(nil), "limits")
	if code != 1 || errOut != "" {
		t.Fatalf("code = %d err = %q", code, errOut)
	}
	if n := len(s.Requests()); n != 3 {
		t.Fatalf("%d requests sent, want exactly 3", n)
	}
	re := regexp.MustCompile(`(?s)^request 1 \| ok \| \d+ ms\nrequest 2 \| quota exhausted \| \d+ ms\nrequest 3 \| unavailable \| \d+ ms\n.*AI Studio console`)
	if !re.MatchString(out) {
		t.Fatalf("out = %q", out)
	}
	assertNoKey(t, out, errOut)
}

func TestAICheckLimitsAllOk(t *testing.T) {
	s := aistest.NewServer(t)
	code, out, _ := runAICheck(s, aiCheckEnv(nil), "limits")
	if code != 0 || strings.Count(out, "| ok |") != 3 || !strings.Contains(out, "AI Studio console") {
		t.Fatalf("code = %d out = %q", code, out)
	}
}

func TestAICheckLimitsClassifiesRejection(t *testing.T) {
	s := aistest.NewServer(t, aistest.WithGenerate(aistest.Behaviour{Status: 401}))
	code, out, _ := runAICheck(s, aiCheckEnv(nil), "limits")
	if code != 1 || strings.Count(out, "| rejected |") != 3 {
		t.Fatalf("code = %d out = %q", code, out)
	}
}

func TestAICheckLimitsLabelsAnyOtherError(t *testing.T) {
	// A model name that fails validation is refused locally with ErrRejected; an unexpected
	// error class would be labelled "error". Exercise the fallback through the classifier.
	if got := probeClass(context.Canceled); got != "error" {
		t.Fatalf("probeClass(other) = %q", got)
	}
	for err, want := range map[error]string{nil: "ok", llm.ErrQuotaExhausted: "quota exhausted", llm.ErrUnavailable: "unavailable", llm.ErrRejected: "rejected"} {
		if got := probeClass(err); got != want {
			t.Errorf("probeClass(%v) = %q, want %q", err, got, want)
		}
	}
}

func TestAICheckAPIFailuresExitOneWithoutLeaks(t *testing.T) {
	for _, status := range []int{429, 401, 500} {
		for _, args := range [][]string{{"models"}, {"ping"}, {"ping", "fr"}} {
			s := aistest.NewServer(t, aistest.WithGenerate(aistest.Behaviour{Status: status, Body: "BODYMARKER"}),
				aistest.WithModels(aistest.Behaviour{Status: status, Body: "BODYMARKER"}))
			code, out, errOut := runAICheck(s, aiCheckEnv(nil), args...)
			if code != 1 || out != "" || !strings.Contains(errOut, fmt.Sprintf("http %d", status)) || strings.Contains(errOut, "BODYMARKER") {
				t.Fatalf("%v %d: code = %d out = %q err = %q", args, status, code, out, errOut)
			}
			assertNoKey(t, out, errOut)
		}
	}
}

func TestAICheckPingStopsAtTheFirstFailure(t *testing.T) {
	s := aistest.NewServer(t, aistest.WithGenerate(aistest.Behaviour{Status: 429}))
	code, out, errOut := runAICheck(s, aiCheckEnv(nil), "ping")
	if code != 1 || out != "" || !strings.Contains(errOut, "ping en") || len(s.Requests()) != 1 {
		t.Fatalf("code = %d out = %q err = %q requests = %d", code, out, errOut, len(s.Requests()))
	}
}

func TestAICheckUsageAndConfigurationErrors(t *testing.T) {
	s := aistest.NewServer(t)
	cases := []struct {
		name string
		env  func(string) string
		args []string
		want string
	}{
		{"no args", aiCheckEnv(nil), nil, "usage:"},
		{"unknown subcommand", aiCheckEnv(nil), []string{"nope"}, "usage:"},
		{"models with extra arg", aiCheckEnv(nil), []string{"models", "x"}, "usage:"},
		{"limits with extra arg", aiCheckEnv(nil), []string{"limits", "x"}, "usage:"},
		{"ping bad language", aiCheckEnv(nil), []string{"ping", "de"}, "usage:"},
		{"ping extra args", aiCheckEnv(nil), []string{"ping", "fr", "en"}, "usage:"},
		{"AI disabled", aiCheckEnv(map[string]string{"AI_ENABLED": "false"}), []string{"models"}, "AI is not enabled"},
		{"AI unset", func(string) string { return "" }, []string{"ping"}, "AI is not enabled"},
		{"missing model", aiCheckEnv(map[string]string{"AI_MODEL": ""}), []string{"models"}, "AI_MODEL"},
		{"missing key", aiCheckEnv(map[string]string{"AI_API_KEY": ""}), []string{"limits"}, "AI_API_KEY"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, out, errOut := runAICheck(s, tc.env, tc.args...)
			if code != 2 || out != "" || !strings.Contains(errOut, tc.want) {
				t.Fatalf("code = %d out = %q err = %q", code, out, errOut)
			}
			assertNoKey(t, out, errOut)
		})
	}
	if n := len(s.Requests()); n != 0 {
		t.Fatalf("%d requests were sent on a refused run", n)
	}
}

func TestAICheckReadsOnlyAIVariables(t *testing.T) {
	asked := map[string]bool{}
	base := aiCheckEnv(nil)
	env := func(k string) string { asked[k] = true; return base(k) }
	s := aistest.NewServer(t)
	if code, _, _ := runAICheck(s, env, "models"); code != 0 {
		t.Fatal("models failed")
	}
	for k := range asked {
		if !strings.HasPrefix(k, "AI_") {
			t.Fatalf("AICheck read %s", k)
		}
	}
}

func TestAICheckHonoursACancelledContext(t *testing.T) {
	s := aistest.NewServer(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var o, e bytes.Buffer
	code := aiCheck(ctx, []string{"ping", "en"}, aiCheckEnv(nil), &o, &e, llm.WithBaseURL(s.URL()))
	if code != 1 || !strings.Contains(e.String(), "unavailable") {
		t.Fatalf("code = %d err = %q", code, e.String())
	}
}
