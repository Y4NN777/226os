package llm

import (
	"regexp"
	"strings"
)

// Both Gemma 4 models put their reasoning in the answer text (V4, ROADMAP). CleanAnswer removes the
// reasoning blocks and normalises the rest, so the assistant only ever shows the answer.
// Only the documented block tags are removed: other markup, code and angle brackets are untouched.
var (
	reasoningBlock      = regexp.MustCompile(`(?is)<think>.*?</think>|<thinking>.*?</thinking>|<thought>.*?</thought>`)
	reasoningUnclosed   = regexp.MustCompile(`(?is)<(?:think|thinking|thought)>.*`)
	blankRun            = regexp.MustCompile(`\n{3,}`)
	lineEndingsToNormal = strings.NewReplacer("\r\n", "\n", "\r", "\n")
)

// CleanAnswer returns the answer text without reasoning blocks. A block that is never closed (the
// token limit cut the reasoning) drops everything after its opening tag: the answer cannot be told
// apart from the reasoning, so nothing of it is kept. Line endings become \n, runs of blank lines
// collapse to one, and surrounding space is trimmed. The result may be empty.
func CleanAnswer(s string) string {
	s = reasoningBlock.ReplaceAllString(s, "")
	s = reasoningUnclosed.ReplaceAllString(s, "")
	s = lineEndingsToNormal.Replace(s)
	s = blankRun.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}
