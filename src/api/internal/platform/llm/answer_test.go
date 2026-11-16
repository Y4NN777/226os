package llm_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/llm"
	"github.com/Y4NN777/226os/src/api/internal/platform/llm/aistest"
)

// Both Gemma 4 models put their reasoning in the answer (V4, ROADMAP). The adapter must hand the
// assistant only the answer: a `thought` part and a reasoning block never reach the user.
// Fixtures are synthetic (no real provider sample is committed); the shapes are the documented ones.

func TestCleanAnswer(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"plain text is kept", "Use the sandbox first.", "Use the sandbox first."},
		{"surrounding space is trimmed", "  \n answer \n\n", "answer"},
		{"think block", "<think>Let me reason.\nStep 2.</think>The answer.", "The answer."},
		{"thinking block, any case", "<THINKING>x</Thinking>Answer", "Answer"},
		{"thought block", "<thought>x</thought>\n\nAnswer", "Answer"},
		{"several blocks", "<think>a</think>One <think>b</think>Two", "One Two"},
		{"reasoning in the middle", "Start\n<think>\nhidden\n</think>\nEnd", "Start\n\nEnd"},
		{"unterminated block drops the rest", "<think>never closed and more text", ""},
		{"line endings are normalised", "a\r\nb", "a\nb"},
		{"blank runs collapse to one", "a\n\n\n\n\nb", "a\n\nb"},
		{"code with angle brackets is untouched", "Use `List<T>` and <b>bold</b>.", "Use `List<T>` and <b>bold</b>."},
	}
	for _, c := range cases {
		if got := llm.CleanAnswer(c.in); got != c.want {
			t.Errorf("%s: CleanAnswer(%q) = %q, want %q", c.name, c.in, got, c.want)
		}
	}
}

func TestGenerateDropsThoughtPartsAndReasoningBlocks(t *testing.T) {
	body := `{"candidates":[{"finishReason":"STOP","content":{"role":"model","parts":[
		{"text":"I should check the sources first.","thought":true},
		{"text":"<think>internal</think>Check the provider docs [1]."}]}}],
		"usageMetadata":{"promptTokenCount":3,"candidatesTokenCount":9}}`
	s := aistest.NewServer(t, aistest.WithGenerate(aistest.Behaviour{Body: body}))
	got, err := newClient(t, s).Generate(context.Background(), llm.Request{User: "q"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Text != "Check the provider docs [1]." {
		t.Fatalf("text = %q, want only the answer", got.Text)
	}
}

func TestGenerateWithOnlyReasoningIsUnavailable(t *testing.T) {
	for name, body := range map[string]string{
		"thought part only": `{"candidates":[{"finishReason":"STOP","content":{"parts":[{"text":"hmm","thought":true}]}}]}`,
		"reasoning only":    aistest.AnswerBody("<think>only reasoning</think>", 1, 1),
		"unterminated":      aistest.AnswerBody("<think>cut off by the token limit", 1, 1),
	} {
		s := aistest.NewServer(t, aistest.WithGenerate(aistest.Behaviour{Body: body}))
		_, err := newClient(t, s).Generate(context.Background(), llm.Request{User: "q"})
		if !errors.Is(err, llm.ErrUnavailable) || !strings.Contains(err.Error(), "empty output") {
			t.Errorf("%s: err = %v, want ErrUnavailable (empty output)", name, err)
		}
	}
}
