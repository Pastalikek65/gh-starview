package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Pastalikek65/gh-starview/internal/cache"
)

var ErrRateLimited = errors.New("rate limited")
var ErrNetwork = errors.New("network error")

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
	return &Client{token: token, baseURL: strings.TrimSuffix(baseURL, "/"), http: &http.Client{}}
}

func (c *Client) ListRepos(ctx context.Context, user string) ([]Repo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/users/%s/repos?per_page=100&sort=updated", c.baseURL, user), nil)
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
		return nil, ErrNetwork
	}
	defer resp.Body.Close()
	if resp.StatusCode == 403 {
		var body map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&body)
		if msg, _ := body["message"].(string); strings.Contains(strings.ToLower(msg), "rate limit") {
			return nil, ErrRateLimited
		}
		return nil, ErrRateLimited
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("github api %d", resp.StatusCode)
	}
	var raw []struct {
		Name        string `json:"name"`
		Description *string `json:"description"`
		Language    *string `json:"language"`
		Stars       int    `json:"stargazers_count"`
		Forks       int    `json:"forks_count"`
		UpdatedAt   string `json:"updated_at"`
		Fork        bool   `json:"fork"`
		HTMLURL     string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	var out []Repo
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
		out = append(out, repo)
	}
	return out, nil
}
