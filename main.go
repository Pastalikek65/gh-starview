package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/Pastalikek65/gh-starview/internal/cache"
	"github.com/Pastalikek65/gh-starview/internal/config"
	"github.com/Pastalikek65/gh-starview/internal/github"
	"github.com/Pastalikek65/gh-starview/internal/tui"
)

var version = "0.1.0"

func main() {
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
				// try gh api user
				user = detectUser()
				if user == "" {
					user = "Pastalikek65"
				}
			}
			return run(user, sortBy, noCache, jsonOut, limit)
		},
	}
	starviewCmd.Flags().StringVar(&sortBy, "sort", "stars", "sort by: stars|name|forks|updated")
	starviewCmd.Flags().BoolVar(&noCache, "no-cache", false, "bypass cache, force fetch")
	starviewCmd.Flags().BoolVar(&jsonOut, "json", false, "output JSON instead of TUI")
	starviewCmd.Flags().IntVar(&limit, "limit", 100, "max repos to show (1-100)")

	// Also add flags to root for direct `gh-starview <user>` invocation
	root.Flags().StringVar(&sortBy, "sort", "stars", "sort by: stars|name|forks|updated")
	root.Flags().BoolVar(&noCache, "no-cache", false, "bypass cache")
	root.Flags().BoolVar(&jsonOut, "json", false, "output JSON")
	root.Flags().IntVar(&limit, "limit", 100, "max repos")

	root.AddCommand(starviewCmd)

	// gh extension support: `gh starview <user>` calls binary as `gh-starview <user>`
	// and `gh-starview starview <user>` should also work. Normalize args.
	args := os.Args[1:]
	// if first arg is not a flag and not "starview" and not "help"/"version", treat as user
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") && args[0] != "starview" && args[0] != "help" && args[0] != "completion" && args[0] != "__complete" {
		// e.g., gh-starview Pastalikek65 --sort name  => convert to starview subcommand
		os.Args = append([]string{os.Args[0], "starview"}, args...)
	} else if len(args) == 0 {
		// no args => default to starview
		os.Args = []string{os.Args[0], "starview"}
	}
	// if called as root with flags but no subcommand, handle directly
	// cobra will dispatch; if root is called without subcommand, run starview logic
	root.RunE = func(cmd *cobra.Command, args []string) error {
		user := ""
		if len(args) == 1 {
			user = args[0]
		}
		if user == "" {
			user = detectUser()
			if user == "" {
				user = "Pastalikek65"
			}
		}
		return run(user, sortBy, noCache, jsonOut, limit)
	}

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(user, sortBy string, noCache, jsonOut bool, limit int) error {
	if err := config.EnsureCacheDir(); err != nil {
		return fmt.Errorf("cache dir: %w", err)
	}
	store, err := cache.Open(config.CacheDBPath())
	if err != nil {
		return fmt.Errorf("open cache: %w", err)
	}
	defer store.Close()

	// If not noCache and cache has data, we will use it as fallback or immediate
	var repos []cache.Repo

	if !noCache {
		// try to show cached data quickly if network fails
		cached, _ := store.List(sortBy)
		if len(cached) > 0 {
			repos = cached
		}
	}

	// Fetch from GitHub
	token := resolveToken()
	client := github.NewClient(token, "")
	fetched, err := client.ListRepos(context.Background(), user)
	if err != nil {
		if len(repos) > 0 {
			fmt.Fprintf(os.Stderr, "⚠️  %v — showing cached data (%d repos)\n", err, len(repos))
		} else {
			return fmt.Errorf("fetch failed for %q: %w (try --no-cache=false or check network/token)", user, err)
		}
	} else {
		if err := store.Upsert(fetched); err != nil {
			fmt.Fprintf(os.Stderr, "warn: cache write failed: %v\n", err)
		}
		// re-read sorted from DB
		sorted, _ := store.List(sortBy)
		if len(sorted) > 0 {
			repos = sorted
		} else {
			repos = fetched
			// sort in memory if DB empty
			// use tui sort logic via store; but fetched is already unsorted, store sorts
		}
		// apply sort if needed (store already sorts, but ensure)
		if len(repos) == 0 {
			repos = fetched
		}
	}

	// apply limit
	if limit > 0 && len(repos) > limit {
		repos = repos[:limit]
	}

	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(repos)
	}

	// If not a TTY, fallback to plain table (CI / Termux non-interactive)
	if !term.IsTerminal(int(os.Stdout.Fd())) || os.Getenv("GH_STARVIEW_PLAIN") == "1" {
		printPlain(repos, sortBy)
		return nil
	}

	m := tui.NewModel(repos)
	m.SortBy(sortBy)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func printPlain(repos []cache.Repo, sortBy string) {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99")).Render(fmt.Sprintf(" gh-starview — %d repos sorted by %s ", len(repos), sortBy))
	fmt.Println(title)
	fmt.Printf(" %-30s %-12s %4s %4s  %-20s\n", "NAME", "LANG", "★", "FORK", "UPDATED")
	fmt.Println(strings.Repeat("─", 78))
	for _, r := range repos {
		fmt.Printf(" %-30s %-12s %4d %4d  %-20s\n", truncate(r.Name, 30), truncate(r.Language, 12), r.Stars, r.Forks, truncate(r.UpdatedAt, 20))
	}
	fmt.Println("\nTip: gh starview --json | jq | gh starview --sort name")
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
	// try gh auth token
	out, err := exec.Command("gh", "auth", "token").Output()
	if err == nil {
		return strings.TrimSpace(string(out))
	}
	return ""
}

func detectUser() string {
	// try gh api user --jq .login
	out, err := exec.Command("gh", "api", "user", "--jq", ".login").Output()
	if err == nil {
		return strings.TrimSpace(string(out))
	}
	return ""
}
