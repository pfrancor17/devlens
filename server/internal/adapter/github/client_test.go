package github_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pfrancor17/devlens/server/internal/adapter/github"
	"github.com/pfrancor17/devlens/server/internal/domain"
)

func TestPullRequestReadsMetadataAndPaginatesFiles(t *testing.T) {
	var authHeader string
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/acme/shop/pulls/7", func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		fmt.Fprint(w, `{"title":"Add refunds","html_url":"https://github.com/acme/shop/pull/7","user":{"login":"pato"}}`)
	})
	mux.HandleFunc("/repos/acme/shop/pulls/7/files", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "1" {
			fmt.Fprint(w, "["+strings.TrimSuffix(strings.Repeat(`{"filename":"a.go","additions":1,"deletions":1},`, 100), ",")+"]")
			return
		}
		fmt.Fprint(w, `[{"filename":"refund_test.go","additions":5,"deletions":0}]`)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := github.NewClient("secret", github.WithBaseURL(server.URL))
	pr, err := client.PullRequest(context.Background(), domain.Repository{Owner: "acme", Name: "shop"}, 7)
	if err != nil {
		t.Fatalf("PullRequest: %v", err)
	}

	if pr.Title != "Add refunds" || pr.Author != "pato" || pr.URL == "" {
		t.Errorf("metadata = %+v", pr)
	}
	if len(pr.Files) != 101 {
		t.Errorf("files = %d, want 101 across two pages", len(pr.Files))
	}
	if authHeader != "Bearer secret" {
		t.Errorf("Authorization = %q, want bearer token", authHeader)
	}
}

func TestPullRequestReportsGitHubErrors(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	client := github.NewClient("", github.WithBaseURL(server.URL))
	_, err := client.PullRequest(context.Background(), domain.Repository{Owner: "acme", Name: "shop"}, 404)
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("err = %v, want a 404 error", err)
	}
}
