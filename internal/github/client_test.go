package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

func TestListReposPagination(t *testing.T) {
	page := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		if page == 1 {
			w.Header().Set("Link", `<http://`+r.Host+`/users/u/repos?page=2>; rel="next"`)
			w.Write([]byte(`[{"name":"a","stargazers_count":1},{"name":"b","stargazers_count":2}]`))
		} else {
			w.Write([]byte(`[{"name":"c","stargazers_count":3}]`))
		}
	}))
	defer srv.Close()
	c := NewClient("", srv.URL)
	repos, err := c.ListRepos(context.Background(), "u")
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 3 {
		t.Fatalf("want 3 paginated repos got %d %+v", len(repos), repos)
	}
	if repos[2].Name != "c" {
		t.Fatalf("want c last got %+v", repos[2])
	}
}

func TestListReposInvalidUser(t *testing.T) {
	// use mock server that would succeed if not validated
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"name":"a","stargazers_count":1}]`))
	}))
	defer srv.Close()
	c := NewClient("", srv.URL)
	_, err := c.ListRepos(context.Background(), "bad/user")
	if err == nil {
		t.Fatal("want error for invalid user with slash")
	}
	_, err = c.ListRepos(context.Background(), "")
	if err == nil {
		t.Fatal("want error for empty user")
	}
	_, err = c.ListRepos(context.Background(), "a-very-long-username-that-exceeds-thirty-nine-chars-limit-xyz")
	if err == nil {
		t.Fatal("want error for too long user")
	}
}

func TestRateLimit429(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		w.Write([]byte(`{"message":"rate limit"}`))
	}))
	defer srv.Close()
	c := NewClient("tok", srv.URL)
	_, err := c.ListRepos(context.Background(), "u")
	if err != ErrRateLimited {
		t.Fatalf("want ErrRateLimited for 429 got %v", err)
	}
}

func Test403NonRateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		w.Write([]byte(`{"message":"Resource not accessible by integration"}`))
	}))
	defer srv.Close()
	c := NewClient("tok", srv.URL)
	_, err := c.ListRepos(context.Background(), "u")
	if err == nil || err == ErrRateLimited || err == ErrNetwork {
		t.Fatalf("want generic 403 error, got %v", err)
	}
	if !contains(err.Error(), "403") {
		t.Fatalf("want 403 in error, got %v", err)
	}
}

func TestLinkHostValidation(t *testing.T) {
	// Link with evil host should not be followed (token exfil prevention)
	page := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		if page == 1 {
			w.Header().Set("Link", `<http://evil.com/users/u/repos?page=2>; rel="next"`)
			w.Write([]byte(`[{"name":"a","stargazers_count":1}]`))
		} else {
			t.Fatal("should not follow evil host")
		}
	}))
	defer srv.Close()
	c := NewClient("tok", srv.URL)
	repos, err := c.ListRepos(context.Background(), "u")
	if err != nil { t.Fatal(err) }
	if len(repos) != 1 {
		t.Fatalf("want 1, got %d", len(repos))
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

func TestListReposContextTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()
	c := NewClient("tok", srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, err := c.ListRepos(ctx, "u")
	if err == nil {
		t.Fatal("want timeout error")
	}
	if err != ErrNetwork {
		t.Fatalf("want ErrNetwork for timeout got %v", err)
	}
}
