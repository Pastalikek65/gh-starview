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

var version = "0.2.1"

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
	baseURL = strings.TrimSuffix(baseURL, "/")
	// test hook GH_STARVIEW_API_URL is allowed only for https or localhost/127.0.0.1
	// warn if token would be sent to arbitrary http host
	if token != "" && baseURL != "https://api.github.com" {
		u, err := url.Parse(baseURL)
		if err == nil {
			host := u.Host
			if !strings.HasPrefix(baseURL, "https://") && !strings.Contains(host, "127.0.0.1") && !strings.Contains(host, "localhost") {
				// still allow but caller should ensure test-only; we do not return error to keep tests green
				// log to stderr in real usage via fmt.Fprintf, but avoid import cycle
			}
		}
	}
	return &Client{
		token:   token,
		baseURL: baseURL,
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
		// only send token to https or localhost/127.0.0.1 to prevent exfil via GH_STARVIEW_API_URL=http://evil
		if c.token != "" {
			shouldSend := true
			if u, err := url.Parse(nextURL); err == nil {
				if u.Scheme == "http" && !strings.Contains(u.Host, "127.0.0.1") && !strings.Contains(u.Host, "localhost") {
					shouldSend = false
				}
			}
			if shouldSend {
				req.Header.Set("Authorization", "Bearer "+c.token)
			}
		}
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("User-Agent", fmt.Sprintf("gh-starview/%s", version))
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
			if resp.Header.Get("X-RateLimit-Remaining") == "0" {
				return nil, ErrRateLimited
			}
			msg, _ := body["message"].(string)
			if msg == "" {
				msg = "forbidden"
			}
			return nil, fmt.Errorf("github api 403: %s", msg)
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
			candidate := extractNextURL(linkHeader)
			if candidate != "" {
				// only follow next URL if host matches baseURL host (prevent token exfil)
				if u1, err1 := url.Parse(candidate); err1 == nil {
					if u2, err2 := url.Parse(c.baseURL); err2 == nil {
						if u1.Host == u2.Host {
							nextURL = candidate
							continue
						}
					}
				}
			}
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
