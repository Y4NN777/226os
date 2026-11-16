package domain_test

import (
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/solve/domain"
)

func TestQuestionStateIsDerived(t *testing.T) {
	tests := []struct {
		name string
		q    domain.Question
		want string
	}{
		{"no answer", domain.Question{}, domain.StateOpen},
		{"answers, none accepted", domain.Question{AnswerCount: 2}, domain.StateAnswered},
		{"an accepted answer", domain.Question{AnswerCount: 2, AcceptedAnswerID: "a1"}, domain.StateResolved},
		{"accepted wins over the count", domain.Question{AcceptedAnswerID: "a1"}, domain.StateResolved},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.q.State(); got != tc.want {
				t.Fatalf("State() = %q, want %q", got, tc.want)
			}
		})
	}
}
