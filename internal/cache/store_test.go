package cache

import (
	"path/filepath"
	"testing"
)

func TestOpenAndUpsertAndList(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	s, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	repos := []Repo{
		{Name: "b", Stars: 5, Language: "Go", UpdatedAt: "2026-01-01T00:00:00Z", URL: "https://github.com/u/b"},
		{Name: "a", Stars: 10, Language: "Python", UpdatedAt: "2026-01-02T00:00:00Z", URL: "https://github.com/u/a"},
	}
	if err := s.Upsert(repos); err != nil {
		t.Fatal(err)
	}

	got, err := s.List("stars")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 got %d", len(got))
	}
	if got[0].Name != "a" {
		t.Fatalf("sort stars: want a first got %s", got[0].Name)
	}

	// update
	repos[0].Stars = 20
	if err := s.Upsert(repos[:1]); err != nil {
		t.Fatal(err)
	}
	got, _ = s.List("name")
	if got[0].Name != "a" {
		t.Fatalf("sort name want a first got %s", got[0].Name)
	}
	// check updated value persisted
	got, _ = s.List("stars")
	if got[0].Name != "b" || got[0].Stars != 20 {
		t.Fatalf("want b with 20 stars first, got %+v", got[0])
	}
}

func TestListEmpty(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "e.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	got, err := s.List("stars")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("want 0 got %d", len(got))
	}
}

func TestListSortVariants(t *testing.T) {
	s, _ := Open(filepath.Join(t.TempDir(), "v.db"))
	defer s.Close()
	repos := []Repo{
		{Name: "c", Stars: 1, Forks: 10, UpdatedAt: "2026-01-01T00:00:00Z"},
		{Name: "a", Stars: 5, Forks: 1, UpdatedAt: "2026-01-03T00:00:00Z"},
		{Name: "b", Stars: 3, Forks: 5, UpdatedAt: "2026-01-02T00:00:00Z"},
	}
	if err := s.Upsert(repos); err != nil {
		t.Fatal(err)
	}
	// forks
	got, _ := s.List("forks")
	if got[0].Name != "c" {
		t.Fatalf("forks sort want c got %s", got[0].Name)
	}
	// updated
	got, _ = s.List("updated")
	if got[0].Name != "a" {
		t.Fatalf("updated sort want a got %s", got[0].Name)
	}
}

func TestUpsertPreservesFields(t *testing.T) {
	s, _ := Open(filepath.Join(t.TempDir(), "p.db"))
	defer s.Close()
	r := Repo{Name: "x", Description: "desc", Language: "Go", Stars: 1, Forks: 2, UpdatedAt: "2026-01-01T00:00:00Z", IsFork: true, URL: "https://github.com/u/x"}
	if err := s.Upsert([]Repo{r}); err != nil {
		t.Fatal(err)
	}
	got, _ := s.List("name")
	if len(got) != 1 {
		t.Fatalf("want 1 got %d", len(got))
	}
	if got[0].Description != "desc" || !got[0].IsFork || got[0].URL != r.URL {
		t.Fatalf("fields not preserved %+v", got[0])
	}
}
