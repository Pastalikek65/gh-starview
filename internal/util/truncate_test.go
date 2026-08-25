package util

import "testing"

func TestTruncate(t *testing.T) {
	if got := Truncate("hello world", 5); got != "he..." {
		t.Fatalf("want he... got %q", got)
	}
	if got := Truncate("hi", 10); got != "hi" {
		t.Fatalf("want hi got %q", got)
	}
	if got := Truncate("café☕️", 5); got == "" {
		t.Fatal("empty for unicode")
	}
	// wide chars: café has width 4, truncating to 4 with "..." gives "c..."
	if got := Truncate("café world", 4); got != "c..." {
		t.Fatalf("unicode truncate want c... got %q", got)
	}
	// ensure no mid-rune cut
	if got := Truncate("café", 4); got != "café" {
		t.Fatalf("want café got %q", got)
	}
	if got := Truncate("abc", 2); got != "ab" {
		t.Fatalf("want ab got %q", got)
	}
}
