package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/Pastalikek65/gh-starview/internal/cache"
	"github.com/Pastalikek65/gh-starview/internal/config"
	"github.com/Pastalikek65/gh-starview/internal/github"
	"github.com/Pastalikek65/gh-starview/internal/tui"
	"github.com/Pastalikek65/gh-starview/internal/util"
)

var version = "0.1.0"

const (
	defaultFetchTimeout = 15 * time.Second
	ghCommandTimeout    = 2 * time.Second
	maxLimit            = 100
	minLimit            = 1
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	var sortBy string
	var noCache bool
	var jsonOut bool
	var limit int

	root := &cobra.Command{
		Use:     "gh-starview",
		Short:   "TUI for GitHub repo metrics — gh extension with SQLite cache",
		Version: version,
		Long: `gh-starview is a gh CLI extension that shows your GitHub repos
as a sortable TUI table (stars/forks/updated) with offline SQLite cache.

  gh starview              # your repos (from gh auth)
  gh starview Pastalikek65 --sort stars
  gh starview --json       # plain JSON (CI friendly)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			user := ""
			if len(args) == 1 {
				user = args[0]
			}
			if user == "" {
				user = detectUser()
				if user == "" {
					return fmt.Errorf("no user specified and gh auth not found; run gh auth login or pass <user>")
				}
			}
			return run(user, sortBy, noCache, jsonOut, limit)
		},
	}

	starviewCmd := &cobra.Command{
		Use:   "starview [user]",
		Short: "Show repo metrics in TUI",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			user := ""
			if len(args) == 1 {
				user = args[0]
			}
			if user == "" {
				user = detectUser()
				if user == "" {
					return fmt.Errorf("no user specified and gh auth not found; run gh auth login or pass <user>")
				}
			}
			return run(user, sortBy, noCache, jsonOut, limit)
		},
	}

	addStarviewFlags := func(cmd *cobra.Command) {
		cmd.Flags().StringVar(&sortBy, "sort", "stars", "sort by: stars|name|forks|updated")
		cmd.Flags().BoolVar(&noCache, "no-cache", false, "bypass cache, force fetch")
		cmd.Flags().BoolVar(&jsonOut, "json", false, "output JSON instead of TUI")
		cmd.Flags().IntVar(&limit, "limit", 100, "max repos to show (1-100)")
	}

	addStarviewFlags(root)
	addStarviewFlags(starviewCmd)
	root.AddCommand(starviewCmd)

	return root
}

func run(user, sortBy string, noCache, jsonOut bool, limit int) error {
	if err := validateFlags(sortBy, limit); err != nil {
		return err
	}
	repos, err := loadRepos(user, sortBy, noCache)
	if err != nil {
		return err
	}
	if limit > 0 && len(repos) > limit {
		repos = repos[:limit]
	}
	return renderRepos(repos, sortBy, jsonOut)
}

func validateFlags(sortBy string, limit int) error {
	switch sortBy {
	case "stars", "name", "forks", "updated":
	default:
		return fmt.Errorf("invalid --sort %q: must be stars|name|forks|updated", sortBy)
	}
	if limit < minLimit || limit > maxLimit {
		return fmt.Errorf("invalid --limit %d: must be 1-100", limit)
	}
	return nil
}

func loadRepos(user, sortBy string, noCache bool) ([]cache.Repo, error) {
	if err := config.EnsureCacheDir(); err != nil {
		return nil, fmt.Errorf("cache dir: %w", err)
	}
	store, err := cache.Open(config.CacheDBPath())
	if err != nil {
		return nil, fmt.Errorf("open cache: %w", err)
	}
	defer store.Close()

	var repos []cache.Repo
	if !noCache {
		cached, err := store.List(sortBy)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn: cache read failed: %v\n", err)
		} else if len(cached) > 0 {
			repos = cached
		}
	}

	token := resolveToken()
	if token == "" {
		fmt.Fprintln(os.Stderr, "⚠️  no token found (set GITHUB_TOKEN or run gh auth login), using unauthenticated 60 req/hr")
	}
	baseURL := os.Getenv("GH_STARVIEW_API_URL")
	client := github.NewClient(token, baseURL)
	ctx, cancel := context.WithTimeout(context.Background(), defaultFetchTimeout)
	defer cancel()
	fetched, err := client.ListRepos(ctx, user)
	if err != nil {
		if len(repos) > 0 {
			fmt.Fprintf(os.Stderr, "⚠️  %v — showing cached data (%d repos)\n", err, len(repos))
			return repos, nil
		}
		return nil, fmt.Errorf("fetch failed for %q: %w (try --no-cache=false or check network/token)", user, err)
	}
	if err := store.Upsert(fetched); err != nil {
		fmt.Fprintf(os.Stderr, "warn: cache write failed: %v\n", err)
	}
	sorted, err := store.List(sortBy)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warn: cache read failed: %v\n", err)
		return fetched, nil
	}
	if len(sorted) > 0 {
		return sorted, nil
	}
	return fetched, nil
}

func renderRepos(repos []cache.Repo, sortBy string, jsonOut bool) error {
	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(repos)
	}
	if !term.IsTerminal(int(os.Stdout.Fd())) || os.Getenv("GH_STARVIEW_PLAIN") == "1" {
		printPlain(repos, sortBy)
		return nil
	}
	m := tui.NewModel(repos)
	m.SortBy(sortBy)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func printPlain(repos []cache.Repo, sortBy string) {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99")).Render(fmt.Sprintf(" gh-starview — %d repos sorted by %s ", len(repos), sortBy))
	fmt.Println(title)
	fmt.Printf(" %-30s %-12s %4s %4s  %-20s\n", "NAME", "LANG", "★", "FORK", "UPDATED")
	fmt.Println(strings.Repeat("─", 78))
	for _, r := range repos {
		fmt.Printf(" %-30s %-12s %4d %4d  %-20s\n", util.Truncate(r.Name, 30), util.Truncate(r.Language, 12), r.Stars, r.Forks, util.Truncate(r.UpdatedAt, 20))
	}
	fmt.Println("\nTip: gh starview --json | jq | gh starview --sort name")
}

func resolveToken() string {
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return t
	}
	if t := os.Getenv("GH_TOKEN"); t != "" {
		return t
	}
	if t := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN"); t != "" {
		return t
	}
	ctx, cancel := context.WithTimeout(context.Background(), ghCommandTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "gh", "auth", "token").Output()
	if err == nil {
		return strings.TrimSpace(string(out))
	}
	return ""
}

func detectUser() string {
	ctx, cancel := context.WithTimeout(context.Background(), ghCommandTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "gh", "api", "user", "--jq", ".login").Output()
	if err == nil {
		return strings.TrimSpace(string(out))
	}
	return ""
}
