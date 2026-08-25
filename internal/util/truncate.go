package util

import "github.com/mattn/go-runewidth"

// Truncate returns s truncated to n display width, adding "..." if needed.
// It is rune-aware and uses go-runewidth to handle wide chars and emojis.
func Truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	w := runewidth.StringWidth(s)
	if w <= n {
		return s
	}
	if n <= 3 {
		return runewidth.Truncate(s, n, "")
	}
	return runewidth.Truncate(s, n, "...")
}
