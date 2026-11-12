package etag_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/etag"
)

func TestCodeConstants(t *testing.T) {
	if apperr.CodePreconditionRequired != "PRECONDITION_REQUIRED" ||
		apperr.CodePreconditionFailed != "PRECONDITION_FAILED" ||
		apperr.CodeMalformedRequest != "MALFORMED_REQUEST" {
		t.Fatal("apperr code constants drifted from the contract")
	}
}

var tagShape = regexp.MustCompile(`^"[A-Za-z0-9_-]{22}"$`)

func TestTagShapeAndDeterminism(t *testing.T) {
	a := etag.Tag("id-1", "2026-01-02T03:04:05.123456789Z")
	if !tagShape.MatchString(a) {
		t.Fatalf("Tag() = %q, want quoted 22 url-safe base64 chars", a)
	}
	if strings.HasPrefix(a, "W/") {
		t.Fatalf("Tag() = %q is weak", a)
	}
	if b := etag.Tag("id-1", "2026-01-02T03:04:05.123456789Z"); a != b {
		t.Fatalf("Tag not deterministic: %q vs %q", a, b)
	}
	if !tagShape.MatchString(etag.Tag()) {
		t.Fatal("Tag() with no parts has wrong shape")
	}
}

func TestTagDistinguishes(t *testing.T) {
	seen := map[string]string{}
	for name, parts := range map[string][]string{
		"ab,c":      {"ab", "c"},
		"a,bc":      {"a", "bc"},
		"abc":       {"abc"},
		"abc,":      {"abc", ""},
		"none":      {},
		"id,time":   {"id", "time"},
		"a NUL b":   {"a\x00b"}, // not the same as the two parts a and b
		"a,b":       {"a", "b"},
		"empty":     {""},
		"two empty": {"", ""},
	} {
		tag := etag.Tag(parts...)
		if prev, dup := seen[tag]; dup {
			t.Fatalf("%s collides with %s: %s", name, prev, tag)
		}
		seen[tag] = name
	}
}

func TestSet(t *testing.T) {
	rec := httptest.NewRecorder()
	etag.Set(rec, `"abc"`)
	if got := rec.Header().Get("ETag"); got != `"abc"` {
		t.Fatalf("ETag = %q", got)
	}
}

func TestCheck(t *testing.T) {
	cur := etag.Tag("q1", "t1")
	other := etag.Tag("q1", "t2")
	tests := []struct {
		name    string
		values  []string // nil = header absent
		current string
		want    apperr.Code // "" = nil error
		viol    bool
	}{
		{"missing", nil, cur, apperr.CodePreconditionRequired, false},
		{"empty", []string{""}, cur, apperr.CodePreconditionRequired, false},
		{"blank", []string{"   "}, cur, apperr.CodePreconditionRequired, false},
		{"two empty lines", []string{"", " "}, cur, apperr.CodePreconditionRequired, false},
		{"star with current", []string{"*"}, cur, "", false},
		{"star without current", []string{"*"}, "", apperr.CodePreconditionFailed, false},
		{"star padded", []string{" * "}, cur, "", false},
		{"star mixed in list", []string{"*, " + cur}, cur, apperr.CodeMalformedRequest, true},
		{"single match", []string{cur}, cur, "", false},
		{"list with one match", []string{other + ", " + cur}, cur, "", false},
		{"list match first no space", []string{cur + "," + other}, cur, "", false},
		{"no match", []string{other}, cur, apperr.CodePreconditionFailed, false},
		{"weak never matches", []string{"W/" + cur}, cur, apperr.CodePreconditionFailed, false},
		{"weak then strong match", []string{"W/" + other + ", " + cur}, cur, "", false},
		{"weak current never matches", []string{cur}, "W/" + cur, apperr.CodePreconditionFailed, false},
		{"whitespace around", []string{"  " + cur + " \t"}, cur, "", false},
		{"whitespace around list", []string{" " + other + " ,\t" + cur + "  "}, cur, "", false},
		{"two header lines", []string{other, cur}, cur, "", false},
		{"two header lines no match", []string{other, `"zzz"`}, cur, apperr.CodePreconditionFailed, false},
		// RFC 9110 section 5.6.1.2: "a recipient MUST parse and ignore a reasonable number of
		// empty list elements". Empty members are ignored; a header with only empty members
		// carries no tag, so it is missing (428). Invalid members are still 400.
		{"empty line plus value", []string{"", cur}, cur, "", false},
		{"only commas", []string{",,"}, cur, apperr.CodePreconditionRequired, false},
		{"commas and spaces only", []string{" , ,\t, "}, cur, apperr.CodePreconditionRequired, false},
		{"star among empty members", []string{", * ,"}, cur, "", false},
		{"empty members around a mismatch", []string{",," + other + ",,"}, cur, apperr.CodePreconditionFailed, false},
		{"empty member before invalid one", []string{",," + `abc`}, cur, apperr.CodeMalformedRequest, true},
		{"star mixed with a tag among empty members", []string{",*,," + cur}, cur, apperr.CodeMalformedRequest, true},
		{"unterminated quote", []string{`"abc`}, cur, apperr.CodeMalformedRequest, true},
		{"unquoted", []string{"abc"}, cur, apperr.CodeMalformedRequest, true},
		{"junk after tag", []string{cur + " junk"}, cur, apperr.CodeMalformedRequest, true},
		{"junk after weak prefix", []string{"W/abc"}, cur, apperr.CodeMalformedRequest, true},
		{"bare W/", []string{"W/"}, cur, apperr.CodeMalformedRequest, true},
		{"trailing comma is an ignorable empty member", []string{cur + ","}, cur, "", false},
		{"leading comma is an ignorable empty member", []string{"," + cur}, cur, "", false},
		{"empty member between tags", []string{cur + ",," + other}, cur, "", false},
		{"only comma", []string{","}, cur, apperr.CodePreconditionRequired, false},
		{"space inside tag", []string{`"a b"`}, cur, apperr.CodeMalformedRequest, true},
		{"control char inside tag", []string{"\"a\x01b\""}, cur, apperr.CodeMalformedRequest, true},
		{"comma inside tag is data", []string{`"a,b"`}, `"a,b"`, "", false},
		{"empty quoted tag", []string{`""`}, cur, apperr.CodePreconditionFailed, false},
		{"empty current never matches", []string{`""`}, "", apperr.CodePreconditionFailed, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequestWithContext(context.Background(), http.MethodPatch, "/x", nil)
			for _, v := range tc.values {
				r.Header.Add("If-Match", v)
			}
			err := etag.Check(r, tc.current)
			if tc.want == "" {
				if err != nil {
					t.Fatalf("Check() = %v, want nil", err)
				}
				return
			}
			var ae *apperr.Error
			if !errors.As(err, &ae) {
				t.Fatalf("Check() = %v, want *apperr.Error %s", err, tc.want)
			}
			if ae.Code != tc.want {
				t.Fatalf("code = %s, want %s", ae.Code, tc.want)
			}
			hasViol := len(ae.Violations) == 1 &&
				ae.Violations[0] == (apperr.Violation{Field: "If-Match", Code: "invalid_format"})
			if tc.viol != hasViol || (!tc.viol && len(ae.Violations) != 0) {
				t.Fatalf("violations = %+v, want invalid_format=%v", ae.Violations, tc.viol)
			}
		})
	}
}

func TestCheckNeverEchoesValues(t *testing.T) {
	const secret = "SUPERSECRETVALUE"
	cur := `"CURRENTTAGVALUE"`
	for _, h := range []string{`"` + secret + `"`, secret, `"` + secret, `W/"` + secret + `"`} {
		r := httptest.NewRequestWithContext(context.Background(), http.MethodPatch, "/x", nil)
		r.Header.Set("If-Match", h)
		err := etag.Check(r, cur)
		if err == nil {
			t.Fatalf("%q: want error", h)
		}
		var ae *apperr.Error
		if !errors.As(err, &ae) {
			t.Fatalf("%q: not apperr", h)
		}
		all := err.Error() + ae.Detail
		for _, v := range ae.Violations {
			all += v.Field + v.Code
		}
		if strings.Contains(all, secret) || strings.Contains(all, "CURRENTTAGVALUE") {
			t.Fatalf("%q: error leaks a value: %q", h, all)
		}
	}
}
