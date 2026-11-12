package cursor_test

import (
	"bytes"
	"encoding/base64"
	"errors"
	"strings"
	"testing"

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
		"unsupported version":       {2, 0, 0, 0},
		"no fields":                 {1},
		"field length past the end": {1, 5, 'a'},
		"bad varint":                {1, 0x80},
		"trailing bytes":            {1, 0, 0, 0, 'x'},
		"only two fields":           {1, 0, 0},
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
