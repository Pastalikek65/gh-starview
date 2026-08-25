package tui

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/Pastalikek65/gh-starview/internal/cache"
	"github.com/Pastalikek65/gh-starview/internal/util"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99")).Padding(0, 1)
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("240"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	normalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	filterStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	faintStyle    = lipgloss.NewStyle().Faint(true)
)

type Model struct {
	all         []cache.Repo
	filtered    []cache.Repo
	cursor      int
	sortBy      string
	filter      string
	filtering   bool
	filterInput string
}

func NewModel(repos []cache.Repo) Model {
	m := Model{all: repos, sortBy: "stars"}
	m.apply()
	return m
}

func (m *Model) apply() {
	var f []cache.Repo
	for _, r := range m.all {
		if m.filter == "" || strings.Contains(strings.ToLower(r.Name), strings.ToLower(m.filter)) {
			f = append(f, r)
		}
	}
	switch m.sortBy {
	case "name":
		sort.Slice(f, func(i, j int) bool { return f[i].Name < f[j].Name })
	case "updated":
		sort.Slice(f, func(i, j int) bool { return f[i].UpdatedAt > f[j].UpdatedAt })
	case "forks":
		sort.Slice(f, func(i, j int) bool { return f[i].Forks > f[j].Forks })
	default:
		sort.Slice(f, func(i, j int) bool { return f[i].Stars > f[j].Stars })
	}
	m.filtered = f
	if m.cursor >= len(m.filtered) {
		m.cursor = 0
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *Model) SortBy(s string) {
	m.sortBy = s
	m.apply()
}

func (m *Model) Filter(s string) {
	m.filter = s
	m.apply()
}

func (m Model) Repos() []cache.Repo {
	return m.filtered
}

func (m Model) IsFiltering() bool { return m.filtering }
func (m Model) FilterInput() string { return m.filterInput }

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		if m.filtering {
			return m.updateFiltering(key)
		}
		return m.updateNormal(key)
	}
	return m, nil
}

func (m *Model) updateFiltering(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		m.filter = m.filterInput
		m.filtering = false
		m.filterInput = ""
		m.apply()
		return *m, nil
	case "esc", "ctrl+c":
		m.filtering = false
		m.filterInput = ""
		return *m, nil
	case "backspace", "ctrl+h":
		if len(m.filterInput) > 0 {
			// rune-safe backspace
			_, size := utf8.DecodeLastRuneInString(m.filterInput)
			m.filterInput = m.filterInput[:len(m.filterInput)-size]
		}
		return *m, nil
	case "ctrl+u":
		m.filterInput = ""
		return *m, nil
	default:
		if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
			m.filterInput += key
			return *m, nil
		}
		return *m, nil
	}
}

func (m *Model) updateNormal(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "q", "ctrl+c":
		return *m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
		}
	case "s":
		m.SortBy("stars")
	case "n":
		m.SortBy("name")
	case "f":
		m.SortBy("forks")
	case "u":
		m.SortBy("updated")
	case "/":
		m.filtering = true
		m.filterInput = ""
		return *m, nil
	case "esc":
		if m.filter != "" {
			m.Filter("")
		}
	}
	return *m, nil
}

func (m Model) View() string {
	s := titleStyle.Render(" gh-starview ") + "\n"
	s += headerStyle.Render(fmt.Sprintf(" %-28s %-10s %4s %4s  %-19s ", "NAME", "LANG", "★", "FORK", "UPDATED")) + "\n"
	s += strings.Repeat("─", 72) + "\n"
	if len(m.filtered) == 0 {
		if m.filter != "" {
			s += fmt.Sprintf("  (no match for %q — esc to clear)\n", m.filter)
		} else {
			s += "  (no repos — try / to filter, q to quit)\n"
		}
	}
	for i, r := range m.filtered {
		line := fmt.Sprintf(" %-28s %-10s %4d %4d  %-19s", util.Truncate(r.Name, 28), util.Truncate(r.Language, 10), r.Stars, r.Forks, util.Truncate(r.UpdatedAt, 19))
		if i == m.cursor {
			s += selectedStyle.Render("▶"+line) + "\n"
		} else {
			s += normalStyle.Render(" "+line) + "\n"
		}
	}
	if m.filtering {
		s += "\n" + filterStyle.Render(fmt.Sprintf("/%s█", m.filterInput)) + faintStyle.Render(" (enter:apply  esc:cancel)") + "\n"
	} else if m.filter != "" {
		s += "\n" + faintStyle.Render(fmt.Sprintf("filter: %q  (/ to change, esc to clear)", m.filter)) + "\n"
	}
	s += "\n" + faintStyle.Render("q:quit  s:stars  n:name  f:forks  u:updated  j/k:nav  /:filter") + "\n"
	return s
}
