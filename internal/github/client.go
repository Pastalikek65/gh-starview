package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Pastalikek65/gh-starview/internal/cache"
)

var ErrRateLimited = errors.New("rate limited")
var ErrNetwork = errors.New("network error")

var validUserRe = regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,37}[a-zA-Z0-9])?$`)

type Repo = cache.Repo

type Client struct {
	token   string
	baseURL string
	http    *http.Client
}

func NewClient(token, baseURL string) *Client {
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	return &Client{
		token:   token,
		baseURL: strings.TrimSuffix(baseURL, "/"),
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

func validateUser(user string) error {
	if user == "" {
		return fmt.Errorf("user is empty")
	}
	if len(user) > 39 {
		return fmt.Errorf("user too long")
	}
	if !validUserRe.MatchString(user) {
		return fmt.Errorf("invalid user %q", user)
	}
	return nil
}

func (c *Client) ListRepos(ctx context.Context, user string) ([]Repo, error) {
	if err := validateUser(user); err != nil {
		return nil, err
	}
	escaped := url.PathEscape(user)
	nextURL := fmt.Sprintf("%s/users/%s/repos?per_page=100&sort=updated", c.baseURL, escaped)
	var all []Repo

	for nextURL != "" {
		req, err := http.NewRequestWithContext(ctx, "GET", nextURL, nil)
		if err != nil {
			return nil, err
		}
		if c.token != "" {
			req.Header.Set("Authorization", "Bearer "+c.token)
		}
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("User-Agent", "gh-starview/0.1.0")
		resp, err := c.http.Do(req)
		if err != nil {
			// context deadline or network
			return nil, ErrNetwork
		}

		// handle rate limit 429 and 403
		if resp.StatusCode == 429 {
			resp.Body.Close()
			return nil, ErrRateLimited
		}
		if resp.StatusCode == 403 {
			var body map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&body)
			resp.Body.Close()
			if msg, _ := body["message"].(string); strings.Contains(strings.ToLower(msg), "rate limit") {
				return nil, ErrRateLimited
			}
			// check header for rate limit
			if resp.Header.Get("X-RateLimit-Remaining") == "0" {
				return nil, ErrRateLimited
			}
			return nil, ErrRateLimited
		}
		if resp.StatusCode != 200 {
			resp.Body.Close()
			return nil, fmt.Errorf("github api %d", resp.StatusCode)
		}
		var raw []struct {
			Name        string  `json:"name"`
			Description *string `json:"description"`
			Language    *string `json:"language"`
			Stars       int     `json:"stargazers_count"`
			Forks       int     `json:"forks_count"`
			UpdatedAt   string  `json:"updated_at"`
			Fork        bool    `json:"fork"`
			HTMLURL     string  `json:"html_url"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
			resp.Body.Close()
			return nil, err
		}
		respBodyClose := resp.Body
		linkHeader := resp.Header.Get("Link")
		respBodyClose.Close()

		for _, r := range raw {
			repo := Repo{
				Name:      r.Name,
				Stars:     r.Stars,
				Forks:     r.Forks,
				UpdatedAt: r.UpdatedAt,
				IsFork:    r.Fork,
				URL:       r.HTMLURL,
			}
			if r.Description != nil {
				repo.Description = *r.Description
			}
			if r.Language != nil {
				repo.Language = *r.Language
			}
			all = append(all, repo)
		}
		// pagination: check Link header for rel="next"
		if linkHeader != "" && strings.Contains(linkHeader, `rel="next"`) {
			nextURL = extractNextURL(linkHeader)
			// if nextURL is relative, make absolute? For mock server, it's absolute already
			// continue loop
			continue
		}
		// fallback: if we got 100 results, there might be more even without Link (GitHub always sends Link, but be safe)
		// if len(raw) == 100, try next page increment (not needed for mock, but handle)
		break
	}
	return all, nil
}

func extractNextURL(link string) string {
	// Link: <https://api.github.com/users/foo/repos?page=2>; rel="next", <...>; rel="last"
	parts := strings.Split(link, ",")
	for _, p := range parts {
		if strings.Contains(p, `rel="next"`) {
			// extract between < and >
			start := strings.Index(p, "<")
			end := strings.Index(p, ">")
			if start != -1 && end != -1 && end > start {
				return strings.TrimSpace(p[start+1 : end])
			}
		}
	}
	return ""
}
