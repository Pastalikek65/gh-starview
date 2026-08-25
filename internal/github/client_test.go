package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListReposMockServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			t.Errorf("bad auth %q", r.Header.Get("Authorization"))
		}
		if r.URL.Path != "/users/u/repos" {
			t.Errorf("bad path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"name":"a","description":"desc","language":"Go","stargazers_count":5,"forks_count":1,"updated_at":"2026-01-01T00:00:00Z","fork":false,"html_url":"https://github.com/u/a"}]`))
	}))
	defer srv.Close()
	c := NewClient("tok", srv.URL)
	repos, err := c.ListRepos(context.Background(), "u")
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 1 || repos[0].Name != "a" || repos[0].Stars != 5 {
		t.Fatalf("bad repos %+v", repos)
	}
	if repos[0].Language != "Go" || repos[0].URL != "https://github.com/u/a" {
		t.Fatalf("fields missing %+v", repos[0])
	}
}

func TestListReposMultiple(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"name":"b","stargazers_count":1},{"name":"a","stargazers_count":10}]`))
	}))
	defer srv.Close()
	c := NewClient("", srv.URL)
	repos, err := c.ListRepos(context.Background(), "u")
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 2 {
		t.Fatalf("want 2 got %d", len(repos))
	}
}

func TestRateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		w.Write([]byte(`{"message":"API rate limit exceeded"}`))
	}))
	defer srv.Close()
	c := NewClient("tok", srv.URL)
	_, err := c.ListRepos(context.Background(), "u")
	if err != ErrRateLimited {
		t.Fatalf("want ErrRateLimited got %v", err)
	}
}

func TestNetworkError(t *testing.T) {
	c := NewClient("tok", "http://127.0.0.1:1")
	_, err := c.ListRepos(context.Background(), "u")
	if err != ErrNetwork {
		t.Fatalf("want ErrNetwork got %v", err)
	}
}

func TestNon200Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"message":"oops"}`))
	}))
	defer srv.Close()
	c := NewClient("tok", srv.URL)
	_, err := c.ListRepos(context.Background(), "u")
	if err == nil || err == ErrRateLimited || err == ErrNetwork {
		t.Fatalf("want generic error got %v", err)
	}
}
