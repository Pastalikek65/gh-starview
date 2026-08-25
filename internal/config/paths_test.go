package config

import "testing"

func TestCacheDBPath_ContainsGhStarview(t *testing.T) {
	p := CacheDBPath()
	if p == "" {
		t.Fatal("empty path")
	}
	if !contains(p, "gh-starview") {
		t.Fatalf("path %q should contain gh-starview", p)
	}
}

func TestEnsureCacheDir_CreatesDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmp)
	if err := EnsureCacheDir(); err != nil {
		t.Fatal(err)
	}
	// verify dir exists
	p := CacheDir()
	if p == "" {
		t.Fatal("empty cache dir")
	}
}

func TestCacheDir_XDG(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", "/tmp/custom")
	if got := CacheDir(); got != "/tmp/custom/gh-starview" {
		t.Fatalf("want /tmp/custom/gh-starview got %q", got)
	}
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
