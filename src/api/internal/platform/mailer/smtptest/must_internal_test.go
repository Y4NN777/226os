package smtptest

import (
	"errors"
	"testing"
)

type failingTB struct {
	testing.TB
	failed bool
}

func (f *failingTB) Helper()               {}
func (f *failingTB) Fatalf(string, ...any) { f.failed = true }

func TestCheckFailsTheTestOnError(t *testing.T) {
	f := &failingTB{TB: t}
	check(f, errors.New("boom"))
	if !f.failed {
		t.Fatal("check did not fail the test")
	}
	f = &failingTB{TB: t}
	check(f, nil)
	if f.failed {
		t.Fatal("check failed on a nil error")
	}
}
