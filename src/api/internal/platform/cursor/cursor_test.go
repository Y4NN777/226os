package cursor_test

import (
	"bytes"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
)

// key returns a distinct 32-byte key per b (not a repeated byte: New refuses those).
func key(b byte) []byte {
	k := make([]byte, 32)
	for i := range k {
		k[i] = b + byte(i)
	}
	return k
}

func newCodec(t *testing.T, b byte) *cursor.Codec {
	t.Helper()
	c, err := cursor.New(key(b))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func assertInvalid(t *testing.T, err error) {
	t.Helper()
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != apperr.CodeInvalidCursor {
		t.Fatalf("err = %v, want INVALID_CURSOR *apperr.Error", err)
	}
}

var sample = cursor.Payload{Sort: "created_at:desc", ID: "0190a1b2-0000-7000-8000-000000000001", FilterHash: "f00d"}

func TestNewRejectsShortKey(t *testing.T) {
	secret := bytes.Repeat([]byte{'s'}, 31)
	_, err := cursor.New(secret)
	if err == nil {
		t.Fatal("expected error for a 31-byte key")
	}
	if strings.Contains(err.Error(), string(secret)) {
		t.Fatal("error echoes the key")
	}
}

func TestRoundTrip(t *testing.T) {
	c := newCodec(t, 1)
	for _, p := range []cursor.Payload{sample, {}, {Sort: "é\x00/+", ID: "x", FilterHash: ""}} {
		got, err := c.Decode(c.Encode(p), p.FilterHash)
		if err != nil {
			t.Fatalf("Decode: %v", err)
		}
		if got != p {
			t.Fatalf("got %+v, want %+v", got, p)
		}
	}
}

func TestEncodeIsOpaqueBase64URL(t *testing.T) {
	s := newCodec(t, 1).Encode(sample)
	if strings.ContainsAny(s, "+/=") {
		t.Fatalf("not base64url-unpadded: %q", s)
	}
	if strings.Contains(s, sample.ID) {
		t.Fatal("cursor leaks the id in clear text")
	}
}

func TestKeyIsCopied(t *testing.T) {
	k := key(2)
	c, err := cursor.New(k)
	if err != nil {
		t.Fatal(err)
	}
	s := c.Encode(sample)
	k[0] ^= 0xff
	if _, err := c.Decode(s, sample.FilterHash); err != nil {
		t.Fatalf("mutating caller key broke the codec: %v", err)
	}
}

func TestDecodeRejections(t *testing.T) {
	c := newCodec(t, 1)
	good := c.Encode(sample)
	raw, err := base64.RawURLEncoding.DecodeString(good)
	if err != nil {
		t.Fatal(err)
	}
	flip := func(i int) string {
		b := append([]byte(nil), raw...)
		b[i] ^= 0x01
		return base64.RawURLEncoding.EncodeToString(b)
	}
	tests := []struct {
		name   string
		in     string
		filter string
		codec  *cursor.Codec
	}{
		{"empty", "", sample.FilterHash, c},
		{"not base64", "!!!not-base64!!!", sample.FilterHash, c},
		{"padded base64 is not canonical", good + "=", sample.FilterHash, c},
		{"truncated", good[:len(good)-4], sample.FilterHash, c},
		{"too short for a signature", base64.RawURLEncoding.EncodeToString([]byte{1, 2, 3}), sample.FilterHash, c},
		{"flipped payload byte", flip(3), sample.FilterHash, c},
		{"flipped signature byte", flip(len(raw) - 1), sample.FilterHash, c},
		{"flipped version byte", flip(0), sample.FilterHash, c},
		{"wrong key", good, sample.FilterHash, newCodec(t, 9)},
		{"mismatched filter hash", good, "other", c},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.codec.Decode(tc.in, tc.filter)
			assertInvalid(t, err)
		})
	}
}

func TestDecodeRejectsValidlySignedGarbage(t *testing.T) {
	c := newCodec(t, 1)
	for name, body := range map[string][]byte{
		"unsupported version":       {3, 0, 0, 0},
		"no fields":                 {2},
		"field length past the end": {2, 5, 'a'},
		"bad varint":                {2, 0x80},
		"trailing bytes":            {2, 0, 0, 0, 'x'},
		"only two fields":           {2, 0, 0},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := c.Decode(cursor.SignForTest(c, body), "")
			assertInvalid(t, err)
		})
	}
}

func TestNewRejectsAKeyWithNoEntropy(t *testing.T) {
	if _, err := cursor.New(make([]byte, 32)); err == nil {
		t.Fatal("an all-zero key must be refused")
	}
}

// clock is a mutable test clock.
type clock struct{ now time.Time }

func (c *clock) Now() time.Time { return c.now }

func expiryCodec(t *testing.T) (*cursor.Codec, *clock) {
	t.Helper()
	clk := &clock{now: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)}
	c, err := cursor.New(key(1), cursor.WithClock(clk.Now))
	if err != nil {
		t.Fatal(err)
	}
	return c, clk
}

func TestCursorExpiry(t *testing.T) {
	ttl := 24 * time.Hour
	tests := []struct {
		name    string
		advance time.Duration
		ok      bool
	}{
		{"fresh", 0, true},
		{"exactly at TTL", ttl, true},
		{"TTL plus one second", ttl + time.Second, false},
		{"future within skew", -5 * time.Minute, true},
		{"future beyond skew", -5*time.Minute - time.Second, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, clk := expiryCodec(t)
			s := c.Encode(sample)
			clk.now = clk.now.Add(tc.advance)
			got, err := c.Decode(s, sample.FilterHash)
			if tc.ok {
				if err != nil || got != sample {
					t.Fatalf("got %+v, %v", got, err)
				}
				return
			}
			assertInvalid(t, err)
			if !strings.Contains(err.Error(), "invalid or expired") || strings.Contains(strings.ToLower(err.Error()), "ttl") {
				t.Fatalf("error text = %q", err)
			}
		})
	}
}

func TestWithTTLOverridesTheDefault(t *testing.T) {
	clk := &clock{now: time.Unix(1_800_000_000, 0)}
	c, err := cursor.New(key(1), cursor.WithClock(clk.Now), cursor.WithTTL(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	s := c.Encode(sample)
	clk.now = clk.now.Add(time.Minute)
	if _, err := c.Decode(s, sample.FilterHash); err != nil {
		t.Fatalf("at TTL: %v", err)
	}
	clk.now = clk.now.Add(time.Second)
	_, err = c.Decode(s, sample.FilterHash)
	assertInvalid(t, err)
}

func TestDefaultClockIsTheWallClock(t *testing.T) {
	c := newCodec(t, 1)
	if _, err := c.Decode(c.Encode(sample), sample.FilterHash); err != nil {
		t.Fatalf("fresh cursor on the real clock: %v", err)
	}
}

func TestTamperedTimestampIsRejectedByTheHMAC(t *testing.T) {
	c, _ := expiryCodec(t)
	raw, err := base64.RawURLEncoding.DecodeString(c.Encode(sample))
	if err != nil {
		t.Fatal(err)
	}
	ts := len(raw) - 32 - 8 // last byte of the issued-at
	raw[ts+7] ^= 0x01
	_, err = c.Decode(base64.RawURLEncoding.EncodeToString(raw), sample.FilterHash)
	assertInvalid(t, err)
}

func TestVersionOneCursorIsRejected(t *testing.T) {
	c, _ := expiryCodec(t)
	body := []byte{1}
	for _, f := range []string{sample.Sort, sample.ID, sample.FilterHash} {
		body = append(body, byte(len(f)))
		body = append(body, f...)
	}
	_, err := c.Decode(cursor.SignForTest(c, body), sample.FilterHash)
	assertInvalid(t, err)
}

func TestVersionTwoWithoutTimestampIsRejected(t *testing.T) {
	c, _ := expiryCodec(t)
	_, err := c.Decode(cursor.SignForTest(c, []byte{2, 0, 0, 0, 1, 2, 3}), "")
	assertInvalid(t, err)
}

func TestNegativeIssuedAtDoesNotPanic(t *testing.T) {
	c, _ := expiryCodec(t)
	body := append([]byte{2, 0, 0, 0}, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff)
	_, err := c.Decode(cursor.SignForTest(c, body), "")
	assertInvalid(t, err)
}

// The contract caps a cursor at 512 characters, and the filters a cursor is bound to can be long: q alone is 200
// runes, 400 bytes in French (#118). Whatever the filters, the cursor stays within the cap, and it still refuses
// other filters.
func TestCursorLengthDoesNotGrowWithTheFilters(t *testing.T) {
	c := newCodec(t, 1)
	filters := "200:" + strings.Repeat("é", 200) + "|" + strings.Repeat("x", 400)
	p := cursor.Payload{Sort: "2026-11-13T09:30:00.123456789Z", ID: "019a0000-0000-7000-8000-0000000000a1", FilterHash: filters}
	token := c.Encode(p)
	if len(token) > 512 {
		t.Fatalf("cursor is %d characters, the contract allows 512", len(token))
	}
	got, err := c.Decode(token, filters)
	if err != nil || got != p {
		t.Fatalf("got %+v, %v", got, err)
	}
	_, err = c.Decode(token, filters+"!")
	assertInvalid(t, err)
}
