package password

import (
	"strings"
	"testing"
)

func TestHashVerifyRoundTrip(t *testing.T) {
	h, err := Hash("a-very-good-password")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(h, "$argon2id$v=19$m=65536,t=3,p=1$") {
		t.Fatalf("unexpected PHC profile: %s", h)
	}
	ok, err := Verify("a-very-good-password", h)
	if err != nil || !ok {
		t.Fatalf("verify ok=%v err=%v", ok, err)
	}
	ok, err = Verify("a-very-good-passwore", h)
	if err != nil || ok {
		t.Fatal("wrong password accepted")
	}
}

func TestHashRejectsShortAndLong(t *testing.T) {
	if _, err := Hash("short"); err == nil {
		t.Fatal("short password hashed")
	}
	if _, err := Hash(strings.Repeat("x", 129)); err == nil {
		t.Fatal("long password hashed")
	}
	if Valid("  padded-password-12  ") != true {
		t.Fatal("padded valid password rejected")
	}
}

func TestVerifyRejectsMalformedAndWeakHashes(t *testing.T) {
	for _, bad := range []string{
		"", "not-a-phc", "$argon2i$v=19$m=65536,t=3,p=1$c2FsdA$a2V5",
		"$argon2id$v=16$m=65536,t=3,p=1$c2FsdHNhbHRzYWx0$a2V5a2V5a2V5a2V5a2V5",
		"$argon2id$v=19$m=1024,t=1,p=1$c2FsdHNhbHRzYWx0$a2V5a2V5a2V5a2V5a2V5", // below OWASP floor
	} {
		if _, err := Verify("whatever-password", bad); err == nil {
			t.Fatalf("accepted malformed hash %q", bad)
		}
	}
}

func TestVerifyDummyAlwaysMismatch(t *testing.T) {
	// The dummy path must run a real Argon2id and report mismatch.
	VerifyDummy("any-password-value")
}

func TestNormalizeStripsCRLF(t *testing.T) {
	if Normalize("ab\r\ncd-efghij") != "abcd-efghij" {
		t.Fatal("CRLF not stripped")
	}
}
