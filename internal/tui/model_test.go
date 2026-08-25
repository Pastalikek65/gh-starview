package tui

import (
	"testing"

	"github.com/Pastalikek65/gh-starview/internal/cache"
)

func TestModelSort(t *testing.T) {
	repos := []cache.Repo{{Name: "b", Stars: 5}, {Name: "a", Stars: 10}}
	m := NewModel(repos)
	m.SortBy("stars")
	if m.Repos()[0].Name != "a" {
		t.Fatalf("want a first got %s", m.Repos()[0].Name)
	}
	m.SortBy("name")
	if m.Repos()[0].Name != "a" {
		t.Fatalf("sort name want a got %s", m.Repos()[0].Name)
	}
	// forks
	repos2 := []cache.Repo{{Name: "x", Forks: 1}, {Name: "y", Forks: 10}}
	m2 := NewModel(repos2)
	m2.SortBy("forks")
	if m2.Repos()[0].Name != "y" {
		t.Fatalf("forks want y got %s", m2.Repos()[0].Name)
	}
	// updated
	repos3 := []cache.Repo{{Name: "old", UpdatedAt: "2026-01-01T00:00:00Z"}, {Name: "new", UpdatedAt: "2026-01-03T00:00:00Z"}}
	m3 := NewModel(repos3)
	m3.SortBy("updated")
	if m3.Repos()[0].Name != "new" {
		t.Fatalf("updated want new got %s", m3.Repos()[0].Name)
	}
}

func TestFilter(t *testing.T) {
	repos := []cache.Repo{{Name: "foo"}, {Name: "bar"}, {Name: "foobar"}}
	m := NewModel(repos)
	m.Filter("foo")
	if len(m.Repos()) != 2 {
		t.Fatalf("filter foo want 2 got %d %+v", len(m.Repos()), m.Repos())
	}
	m.Filter("bar")
	// substring match: both "bar" and "foobar" contain "bar"
	if len(m.Repos()) != 2 {
		t.Fatalf("filter bar want 2 got %d %+v", len(m.Repos()), m.Repos())
	}
	m.Filter("foobar")
	if len(m.Repos()) != 1 || m.Repos()[0].Name != "foobar" {
		t.Fatalf("filter foobar failed %+v", m.Repos())
	}
	m.Filter("")
	if len(m.Repos()) != 3 {
		t.Fatalf("filter empty want 3 got %d", len(m.Repos()))
	}
}

func TestModelCursorAndView(t *testing.T) {
	repos := []cache.Repo{{Name: "a", Stars: 1, Language: "Go", UpdatedAt: "2026-01-01T00:00:00Z"}}
	m := NewModel(repos)
	v := m.View()
	if v == "" {
		t.Fatal("view empty")
	}
	if !contains(v, "a") {
		t.Fatalf("view should contain repo name, got %q", v)
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
