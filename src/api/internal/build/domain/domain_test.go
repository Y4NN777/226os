package domain

import (
	"strings"
	"testing"
)

func TestNormalizeGitHubURLAcceptsAndCanonicalises(t *testing.T) {
	tests := []struct{ name, in, want, owner, repo string }{
		{"mixed case is lowered", "https://github.com/Y4NN777/226os", "https://github.com/y4nn777/226os", "y4nn777", "226os"},
		{"host case and .git", "https://GitHub.com/Org/Repo.git", "https://github.com/org/repo", "org", "repo"},
		{"trailing slash", "https://github.com/org/repo/", "https://github.com/org/repo", "org", "repo"},
		{"dots, underscores and dashes", "https://github.com/org/my.repo_x-1", "https://github.com/org/my.repo_x-1", "org", "my.repo_x-1"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, owner, repo, ok := NormalizeGitHubURL(tc.in)
			if !ok || got != tc.want || owner != tc.owner || repo != tc.repo {
				t.Fatalf("%q -> (%q %q %q %v), want %q", tc.in, got, owner, repo, ok, tc.want)
			}
		})
	}
}

func TestNormalizeGitHubURLRejects(t *testing.T) {
	tests := []struct{ name, in string }{
		{"not https", "http://github.com/org/repo"},
		{"user info", "https://user@github.com/org/repo"},
		{"port", "https://github.com:443/org/repo"},
		{"query", "https://github.com/org/repo?tab=1"},
		{"empty query", "https://github.com/org/repo?"},
		{"fragment", "https://github.com/org/repo#readme"},
		{"other github host", "https://gist.github.com/org/repo"},
		{"look-alike host", "https://github.com.evil.test/org/repo"},
		{"ip literal", "https://127.0.0.1/org/repo"},
		{"no repository", "https://github.com/org"},
		{"deeper path", "https://github.com/org/repo/tree/main"},
		{"owner starts with a dash", "https://github.com/-org/repo"},
		{"dot-dot segment", "https://github.com/org/.."},
		{"dot segment", "https://github.com/org/."},
		{"empty after .git", "https://github.com/org/.git"},
		{"double .git", "https://github.com/org/repo.git.git"},
		{"escaped space", "https://github.com/org/re%20po"},
		{"opaque", "https:github.com/org/repo"},
		{"bad escape", "https://github.com/org/%zz"},
		{"repository too long", "https://github.com/org/" + strings.Repeat("a", 101)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got, _, _, ok := NormalizeGitHubURL(tc.in); ok {
				t.Fatalf("%q accepted as %q", tc.in, got)
			}
		})
	}
}

// FuzzNormalizeGitHubURL: whatever is accepted is canonical (it normalises to itself) and fits the
// projects_github_url_ck backstop.
func FuzzNormalizeGitHubURL(f *testing.F) {
	for _, s := range []string{"https://github.com/Org/Repo.git", "https://github.com/a/b/", "https://github.com/a/..", "http://x"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, in string) {
		got, owner, repo, ok := NormalizeGitHubURL(in)
		if !ok {
			return
		}
		again, o2, r2, ok2 := NormalizeGitHubURL(got)
		if !ok2 || again != got || o2 != owner || r2 != repo {
			t.Fatalf("%q -> %q is not stable (%q %v)", in, got, again, ok2)
		}
		if got != strings.ToLower(got) || strings.HasSuffix(got, ".git") || got != "https://github.com/"+owner+"/"+repo {
			t.Fatalf("%q -> %q breaks the stored form", in, got)
		}
	})
}
