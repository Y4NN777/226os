package app_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/app"
)

// The exported entry points never reach the production endpoint in tests: these cases stop before
// any request is built.
func TestAICheckHelpIsSuccessAndNeedsNoConfiguration(t *testing.T) {
	var o, e bytes.Buffer
	code := app.AICheck(context.Background(), []string{"help"}, func(string) string { return "" }, &o, &e)
	if code != 0 || e.Len() != 0 || !strings.Contains(o.String(), "usage: aicheck") {
		t.Fatalf("code = %d out = %q err = %q", code, o.String(), e.String())
	}
}

func TestAICheckRefusesWhenAIIsNotEnabled(t *testing.T) {
	var o, e bytes.Buffer
	code := app.AICheck(context.Background(), []string{"models"}, func(string) string { return "" }, &o, &e)
	if code != 2 || o.Len() != 0 || !strings.Contains(e.String(), "AI is not enabled") {
		t.Fatalf("code = %d out = %q err = %q", code, o.String(), e.String())
	}
}

func TestAICheckMainExitsOnlyOnFailure(t *testing.T) {
	var o, e bytes.Buffer
	env := func(string) string { return "" }
	exit := func(c int) { t.Fatalf("exit(%d) called on success", c) }
	app.AICheckMain(context.Background(), []string{"help"}, env, &o, &e, exit)

	var got []int
	app.AICheckMain(context.Background(), []string{"nope"}, env, &o, &e, func(c int) { got = append(got, c) })
	if len(got) != 1 || got[0] != 2 {
		t.Fatalf("exit calls = %v", got)
	}
}
