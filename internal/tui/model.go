package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/Pastalikek65/gh-starview/internal/cache"
)

type Model struct {
	all      []cache.Repo
	filtered []cache.Repo
	cursor   int
	sortBy   string
	filter   string
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

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
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
		}
	}
	return m, nil
}

func (m Model) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99")).Padding(0, 1)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("240"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	s := titleStyle.Render(" gh-starview ") + "\n"
	s += headerStyle.Render(fmt.Sprintf(" %-28s %-10s %4s %4s  %-19s ", "NAME", "LANG", "★", "FORK", "UPDATED")) + "\n"
	s += strings.Repeat("─", 72) + "\n"
	if len(m.filtered) == 0 {
		s += "  (no repos — try / to filter, q to quit)\n"
	}
	for i, r := range m.filtered {
		line := fmt.Sprintf(" %-28s %-10s %4d %4d  %-19s", truncate(r.Name, 28), truncate(r.Language, 10), r.Stars, r.Forks, truncate(r.UpdatedAt, 19))
		if i == m.cursor {
			s += selectedStyle.Render("▶"+line) + "\n"
		} else {
			s += normalStyle.Render(" "+line) + "\n"
		}
	}
	s += "\n" + lipgloss.NewStyle().Faint(true).Render("q:quit  s:stars  n:name  f:forks  u:updated  j/k:nav  /:filter") + "\n"
	return s
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}
