package domain_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/Y4NN777/226os/src/api/internal/assistant/domain"
)

// AI-3: a sentence that affirms a guarantee about an API, or presents a solution as verified, is
// removed server-side; negated forms are honest and stay.
func TestRemoveGuarantees(t *testing.T) {
	tests := []struct {
		name, in, want string
		removed        bool
	}{
		{"fr garanti de fonctionner", "Signez le corps [S1]. Cette API est garantie de fonctionner. Réessayez [S2].", "Signez le corps [S1]. Réessayez [S2].", true},
		{"fr garantis", "Les appels sont garantis de fonctionner.", "", true},
		{"fr je vous garantis", "Je vous garantis que cela marche. Voir [S1].", "Voir [S1].", true},
		{"fr nous garantissons", "Nous vous garantissons le résultat [S1].", "", true},
		{"fr coup sûr", "Cela fonctionnera à coup sûr [S1]. Lisez la doc [S2].", "Lisez la doc [S2].", true},
		{"fr solution vérifiée", "C'est une solution vérifiée [S1]. Testez-la [S2].", "Testez-la [S2].", true},
		{"fr 100 % fiable", "Le service est 100 % fiable [S1].", "", true},
		{"fr fonctionne toujours", "Cet endpoint fonctionne toujours [S1].", "", true},
		{"en guaranteed to work", "Sign the body [S1]. This call is guaranteed to work. Retry [S2].", "Sign the body [S1]. Retry [S2].", true},
		{"en I guarantee", "I guarantee it works [S1]!", "", true},
		{"en will definitely work", "It will definitely work [S1]? Read the docs [S2].", "Read the docs [S2].", true},
		{"en verified solution", "This is a verified solution [S1].", "", true},
		{"en 100% reliable", "The API is 100% reliable [S1].", "", true},
		{"en always works", "The endpoint always works [S1].", "", true},
		{"en case and curly apostrophe", "WE GUARANTEE this [S1]. Fine [S2].", "Fine [S2].", true},
		{"one line per sentence keeps the heading", "Answer [S1].\nI guarantee it.\nComment vérifier\n- Test it.", "Answer [S1].\n\nComment vérifier\n- Test it.", true},
		{"fr aucune garantie", "Il n'y a aucune garantie que cela fonctionne [S1].", "Il n'y a aucune garantie que cela fonctionne [S1].", false},
		{"fr pas garanti", "Le résultat n'est pas garanti [S1].", "Le résultat n'est pas garanti [S1].", false},
		{"fr pas garantie de fonctionner", "Elle n'est pas garantie de fonctionner [S1].", "Elle n'est pas garantie de fonctionner [S1].", false},
		{"fr ne garantit pas", "La doc ne garantit pas la disponibilité [S1].", "La doc ne garantit pas la disponibilité [S1].", false},
		{"fr sans garantie", "Utilisez-la sans garantie [S1].", "Utilisez-la sans garantie [S1].", false},
		{"fr ne fonctionne pas toujours", "Elle ne fonctionne pas toujours [S1].", "Elle ne fonctionne pas toujours [S1].", false},
		{"fr pas une solution vérifiée", "Ce n'est pas une solution vérifiée [S1].", "Ce n'est pas une solution vérifiée [S1].", false},
		{"en no guarantee", "There is no guarantee it works [S1].", "There is no guarantee it works [S1].", false},
		{"en not guaranteed", "Success is not guaranteed [S1].", "Success is not guaranteed [S1].", false},
		{"en cannot guarantee", "I cannot guarantee that [S1]. I can't guarantee it either.", "I cannot guarantee that [S1]. I can't guarantee it either.", false},
		{"en does not guarantee curly", "The vendor doesn’t guarantee uptime [S1].", "The vendor doesn’t guarantee uptime [S1].", false},
		{"en not always works", "It does not always work [S1]. It doesn't always work.", "It does not always work [S1]. It doesn't always work.", false},
		{"en not a verified solution", "This is not a verified solution [S1].", "This is not a verified solution [S1].", false},
		{"no claim", "Sign the raw body [S1].", "Sign the raw body [S1].", false},
		{"dots inside a sentence are not boundaries", "Call api.example.test/v1 [S1], it is guaranteed to work.", "", true},
		{"empty", "", "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, removed := domain.RemoveGuarantees(tc.in)
			if strings.TrimSpace(got) != tc.want || removed != tc.removed {
				t.Fatalf("RemoveGuarantees(%q) = %q, %v; want %q, %v", tc.in, got, removed, tc.want, tc.removed)
			}
		})
	}
}

func FuzzRemoveGuarantees(f *testing.F) {
	for _, s := range []string{"", "I guarantee it.", "n'est pas garanti.\n\n", "...!!!???", "a\x00b\xff", "[S1] garanti de fonctionner", "100 % fiable"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, in string) {
		out, removed := domain.RemoveGuarantees(in)
		if len(out) > len(in) || (!removed && out != in) {
			t.Fatalf("out %q from %q (removed %v)", out, in, removed)
		}
		if utf8.ValidString(in) && !utf8.ValidString(out) {
			t.Fatalf("invalid UTF-8 out of valid input: %q", out)
		}
		again, removedAgain := domain.RemoveGuarantees(out)
		if removedAgain || again != out {
			t.Fatalf("not idempotent: %q then %q", out, again)
		}
	})
}
