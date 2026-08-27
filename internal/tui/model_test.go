package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Pastalikek65/gh-starview/internal/cache"
	"github.com/Pastalikek65/gh-starview/internal/util"
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

func TestInteractiveFilter(t *testing.T) {
	repos := []cache.Repo{{Name: "foo"}, {Name: "bar"}, {Name: "foobar"}}
	m := NewModel(repos)
	// press /
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = newM.(Model)
	if !m.IsFiltering() {
		t.Fatal("want filtering after /")
	}
	// type f o o
	for _, r := range "foo" {
		newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = newM.(Model)
	}
	if m.FilterInput() != "foo" {
		t.Fatalf("want filterInput foo got %q", m.FilterInput())
	}
	// enter to apply
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newM.(Model)
	if m.IsFiltering() {
		t.Fatal("should not be filtering after enter")
	}
	if len(m.Repos()) != 2 {
		t.Fatalf("want 2 after filter foo got %d %+v", len(m.Repos()), m.Repos())
	}
}

func TestFilterCancel(t *testing.T) {
	repos := []cache.Repo{{Name: "foo"}, {Name: "bar"}}
	m := NewModel(repos)
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = newM.(Model)
	for _, r := range "foo" {
		newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = newM.(Model)
	}
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = newM.(Model)
	if m.IsFiltering() {
		t.Fatal("should cancel filtering on esc")
	}
	if len(m.Repos()) != 2 {
		t.Fatalf("want 2 after cancel got %d", len(m.Repos()))
	}
}

func TestFilterCancelCtrlC(t *testing.T) {
	repos := []cache.Repo{{Name: "foo"}, {Name: "bar"}}
	m := NewModel(repos)
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = newM.(Model)
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m = newM.(Model)
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m = newM.(Model)
	if m.IsFiltering() {
		t.Fatal("ctrl+c should cancel filtering")
	}
}

func TestNavigationAndSortKeys(t *testing.T) {
	repos := []cache.Repo{{Name: "b", Stars: 5}, {Name: "a", Stars: 10}, {Name: "c", Stars: 1}}
	m := NewModel(repos)
	if m.Init() != nil {
		t.Fatal("Init should return nil")
	}
	// sort by stars: a,b,c
	if m.Repos()[0].Name != "a" {
		t.Fatalf("want a first")
	}
	// j down
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newM.(Model)
	if m.Repos()[1].Name != "b" {
		// cursor moved but repos order same; check view contains cursor?
	}
	// s sort stars (already)
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = newM.(Model)
	if m.Repos()[0].Name != "a" {
		t.Fatalf("s sort failed")
	}
	// n sort name
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = newM.(Model)
	if m.Repos()[0].Name != "a" {
		t.Fatalf("n sort want a got %s", m.Repos()[0].Name)
	}
	// q should quit
	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("q should return quit cmd")
	}
	_ = newM
}

func TestFilterInputBackspaceAndClear(t *testing.T) {
	repos := []cache.Repo{{Name: "foo"}, {Name: "bar"}}
	m := NewModel(repos)
	// enter filter mode
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = newM.(Model)
	// type ab
	for _, r := range "ab" {
		newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = newM.(Model)
	}
	// backspace
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = newM.(Model)
	if m.FilterInput() != "a" {
		t.Fatalf("want a after backspace got %q", m.FilterInput())
	}
	// ctrl+u clear
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	m = newM.(Model)
	if m.FilterInput() != "" {
		t.Fatalf("want empty after ctrl+u got %q", m.FilterInput())
	}
	// view should show filtering
	v := m.View()
	if !contains(v, "/") {
		t.Fatalf("view should show filter input, got %q", v)
	}
}

func TestViewStates(t *testing.T) {
	// empty
	m := NewModel(nil)
	v := m.View()
	if !contains(v, "no repos") {
		t.Fatalf("empty view want no repos, got %q", v)
	}
	// with filter
	repos := []cache.Repo{{Name: "foo"}}
	m = NewModel(repos)
	m.Filter("zzz")
	v = m.View()
	if !contains(v, "no match") {
		t.Fatalf("want no match, got %q", v)
	}
	// truncate via util (rune-aware)
	if util.Truncate("hello world", 5) != "he..." {
		t.Fatalf("truncate want he... got %q", util.Truncate("hello world", 5))
	}
	if util.Truncate("hi", 10) != "hi" {
		t.Fatalf("truncate short want hi got %q", util.Truncate("hi", 10))
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
