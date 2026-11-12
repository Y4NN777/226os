package httpx_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
)

type payload struct {
	Name string `json:"name"`
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("connection reset") }

const jsonCT = "application/json"

func decode(t *testing.T, ct string, body io.Reader) (payload, error) {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/x", body)
	if ct != "" {
		req.Header.Set("Content-Type", ct)
	}
	var dst payload
	err := httpx.DecodeJSON(httptest.NewRecorder(), req, &dst)
	return dst, err
}

func TestDecodeJSONSuccess(t *testing.T) {
	for _, ct := range []string{jsonCT, "application/json; charset=utf-8", "Application/JSON;charset=UTF-8"} {
		got, err := decode(t, ct, strings.NewReader(` {"name":"ada"} `+"\n"))
		if err != nil || got.Name != "ada" {
			t.Fatalf("ct %q: got %+v, %v", ct, got, err)
		}
	}
}

func TestDecodeJSONRejections(t *testing.T) {
	huge := `{"name":"` + strings.Repeat("a", 1<<20) + `"}`
	paddedTrailing := `{"name":"a"}` + strings.Repeat(" ", 1<<20)
	tests := []struct {
		name      string
		ct        string
		body      io.Reader
		code      apperr.Code
		wantField string
		wantVCode string
	}{
		{"missing content-type", "", strings.NewReader(`{}`), apperr.CodeValidation, "Content-Type", "invalid_value"},
		{"text/plain", "text/plain", strings.NewReader(`{}`), apperr.CodeValidation, "Content-Type", "invalid_value"},
		{"unparseable content-type", "application/json;;=", strings.NewReader(`{}`), apperr.CodeValidation, "Content-Type", "invalid_value"},
		{"non-utf8 charset", "application/json; charset=latin1", strings.NewReader(`{}`), apperr.CodeValidation, "Content-Type", "invalid_value"},
		{"oversized body", jsonCT, strings.NewReader(huge), apperr.CodePayloadTooLarge, "", ""},
		{"oversized trailing data", jsonCT, strings.NewReader(paddedTrailing), apperr.CodePayloadTooLarge, "", ""},
		{"syntax error", jsonCT, strings.NewReader(`{"name":`), apperr.CodeMalformedRequest, "", ""},
		{"invalid json", jsonCT, strings.NewReader(`{name: 1}`), apperr.CodeMalformedRequest, "", ""},
		{"type error", jsonCT, strings.NewReader(`{"name": 5}`), apperr.CodeMalformedRequest, "", ""},
		{"empty body", jsonCT, strings.NewReader(``), apperr.CodeMalformedRequest, "", ""},
		{"whitespace-only body", jsonCT, strings.NewReader("  \n"), apperr.CodeMalformedRequest, "", ""},
		{"trailing value", jsonCT, strings.NewReader(`{"name":"a"} {"name":"b"}`), apperr.CodeMalformedRequest, "", ""},
		{"trailing garbage", jsonCT, strings.NewReader(`{"name":"a"} x`), apperr.CodeMalformedRequest, "", ""},
		{"read failure", jsonCT, failingReader{}, apperr.CodeMalformedRequest, "", ""},
		{"unknown field", jsonCT, strings.NewReader(`{"name":"a","bogus":1}`), apperr.CodeValidation, "bogus", "unknown_field"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := decode(t, tc.ct, tc.body)
			var ae *apperr.Error
			if !errors.As(err, &ae) || ae.Code != tc.code {
				t.Fatalf("err = %v, want code %s", err, tc.code)
			}
			if tc.wantField == "" {
				if len(ae.Violations) != 0 {
					t.Fatalf("violations = %+v", ae.Violations)
				}
				return
			}
			if len(ae.Violations) != 1 || ae.Violations[0].Field != tc.wantField || ae.Violations[0].Code != tc.wantVCode {
				t.Fatalf("violations = %+v", ae.Violations)
			}
		})
	}
}

func TestDecodeJSONNeverEchoesInput(t *testing.T) {
	_, err := decode(t, jsonCT, strings.NewReader(`{"name":"top-secret-value`))
	if err == nil || strings.Contains(err.Error(), "top-secret-value") {
		t.Fatalf("err = %v", err)
	}
}

func TestDecodeJSONTruncatesLongUnknownFieldName(t *testing.T) {
	long := strings.Repeat("é", 300)
	_, err := decode(t, jsonCT, strings.NewReader(`{"`+long+`":1}`))
	var ae *apperr.Error
	if !errors.As(err, &ae) || len(ae.Violations) != 1 {
		t.Fatalf("err = %v", err)
	}
	if got := []rune(ae.Violations[0].Field); len(got) != 128 {
		t.Fatalf("field runes = %d, want 128", len(got))
	}
}
