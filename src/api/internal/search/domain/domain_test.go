package domain

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// C4 DECIDED option 1, ADR-0005 point 6: the server truncates q to 100 runes at a word boundary.
func TestTruncateQueryC4Option1(t *testing.T) {
	word := strings.Repeat("a", 9) + " " // 10 runes per word
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"under 100 is unchanged", "paiement mobile", "paiement mobile"},
		{"exactly 100 is unchanged", strings.Repeat("a", 100), strings.Repeat("a", 100)},
		{"the 101st rune is a space: keep the 100 runes", strings.Repeat("a", 100) + " tail", strings.Repeat("a", 100)},
		{"cut inside a word: back up to the last space", strings.Repeat("a", 99) + " tail", strings.Repeat("a", 99)},
		{"word boundary: drop the partial last word", strings.Repeat(word, 9) + "abcdefghijklmnop", strings.TrimSpace(strings.Repeat(word, 9))},
		{"no space in the first 100: hard cut at 100 runes", strings.Repeat("b", 150), strings.Repeat("b", 100)},
		{"multi-byte runes are counted as runes, never split", strings.Repeat("é", 150), strings.Repeat("é", 100)},
		{"multi-byte with spaces", strings.Repeat("é", 95) + " " + strings.Repeat("ü", 20), strings.Repeat("é", 95)},
		{"leading space only is a hard cut", " " + strings.Repeat("c", 150), " " + strings.Repeat("c", 99)},
		{"leading spaces only before the first word: hard cut, never empty", "  " + strings.Repeat("x", 150), "  " + strings.Repeat("x", 98)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := TruncateQuery(tc.in)
			if got != tc.want {
				t.Fatalf("TruncateQuery() = %q (%d runes), want %q", got, utf8.RuneCountInString(got), tc.want)
			}
			if utf8.RuneCountInString(got) > MaxTextRunes || !utf8.ValidString(got) {
				t.Fatalf("result exceeds %d runes or is invalid UTF-8", MaxTextRunes)
			}
		})
	}
}

func TestSnippetIsPlainTextOfAtMost300Runes(t *testing.T) {
	long := strings.Repeat("mot ", 200)
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "Paiement mobile", "Paiement mobile"},
		{"empty", "", ""},
		{"html tags are removed whole", "<b>gras</b> <script>alert(1)</script>", "gras alert(1)"},
		{"an unterminated tag does not survive", "texte <img src=x onerror=alert(1)", "texte"},
		{"nested tags do not rebuild a tag", "<<b>script>alert(1)<</b>/script>", "alert(1)"},
		{"a tag hidden by a tag does not survive", "texte <<i>img src=x onerror=alert(1)>", "texte"},
		{"inline code cannot rebuild a tag", "texte <`script`>alert(1)", "texte alert(1)"},
		{"inline code hiding a tag does not survive", "texte <`img src=x onerror=alert(1)`>", "texte"},
		{"emphasis cannot rebuild a tag", "a <*b*>gras", "a gras"},
		{"bold cannot rebuild a tag", "a <**b**>gras", "a gras"},
		{"strikethrough cannot rebuild a tag", "a <~~b~~>gras", "a gras"},
		{"programming text keeps its characters (#118)", "How to use C# records, ~/.config, 5 * 3, a < b, x -> y", "How to use C# records, ~/.config, 5 * 3, a < b, x -> y"},
		{"line start structure goes", "## Titre\n> cité\n- puce\n* autre\n+ encore", "Titre cité puce autre encore"},
		{"emphasis pairs go, lone markers stay", "*it* et **gras** et a*b*c et 2 * 3 * 4", "it et gras et a*b*c et 2 * 3 * 4"},
		{"bidi and control characters are dropped", "a\u202eb\u200bc\u0000d\u2066e\ufeff", "abcde"},
		{"markdown markers are dropped", "# Titre\n\n**gras** et `code` ~~barré~~", "Titre gras et code barré"},
		{"links keep their text", "voir [la doc](https://example.org/x) ici", "voir la doc ici"},
		{"whitespace is collapsed", "a \n\n\t b   c", "a b c"},
		{"fenced code markers", "```go\nfmt.Println()\n```", "go fmt.Println()"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Snippet(tc.in); got != tc.want {
				t.Fatalf("Snippet() = %q, want %q", got, tc.want)
			}
		})
	}
	got := Snippet(long)
	if n := utf8.RuneCountInString(got); n > MaxSnippetRunes || n < MaxSnippetRunes-10 {
		t.Fatalf("long snippet has %d runes", n)
	}
	if strings.ContainsAny(Snippet(strings.Repeat("é<>", 400)), "<>") {
		t.Fatal("a snippet never contains HTML angle brackets")
	}
	multi := Snippet(strings.Repeat("é", 400))
	if utf8.RuneCountInString(multi) != MaxSnippetRunes || !utf8.ValidString(multi) {
		t.Fatalf("multi-byte snippet: %d runes", utf8.RuneCountInString(multi))
	}
}

func TestPlainTextDropsControlAndBidiAndKeepsTheRest(t *testing.T) {
	tests := []struct{ name, in, want string }{
		{"markdown is kept: a title is not markdown", "How to use **C#**", "How to use **C#**"},
		{"bidi override and control are dropped", "re\u202etset\u0007  ok", "retset ok"},
		{"joiners of Persian and emoji text stay", "می\u200cخواهم", "می\u200cخواهم"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := PlainText(tc.in); got != tc.want {
				t.Fatalf("PlainText() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestGroupsAreTheContractEnumInOrder(t *testing.T) {
	got := Groups()
	want := []Group{Questions, Resources, Projects, Learn}
	if len(got) != 4 {
		t.Fatalf("groups = %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("groups = %v, want %v", got, want)
		}
	}
	for _, g := range want {
		if !g.Valid() {
			t.Fatalf("%q must be valid", g)
		}
	}
	if Group("answers").Valid() {
		t.Fatal("answers are not a search group")
	}
}
