// Package github reads pull requests from the GitHub REST API. It implements
// port.PullRequestSource and knows nothing about risk.
package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pfrancor17/devlens/server/internal/domain"
)

const (
	defaultBaseURL = "https://api.github.com"
	filesPerPage   = 100
	// GitHub returns at most 3,000 files for a pull request.
	maxFilePages = 30
)

// Client is a minimal GitHub API client.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

// Option customizes a Client.
type Option func(*Client)

// WithBaseURL points the client at another API root, such as GitHub
// Enterprise or a test server.
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

// NewClient builds a client. An empty token works for public repositories
// with GitHub's lower anonymous rate limit.
func NewClient(token string, opts ...Option) *Client {
	c := &Client{
		baseURL: defaultBaseURL,
		token:   token,
		http:    &http.Client{Timeout: 15 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type apiPullRequest struct {
	Title   string `json:"title"`
	HTMLURL string `json:"html_url"`
	User    struct {
		Login string `json:"login"`
	} `json:"user"`
}

type apiFile struct {
	Filename  string `json:"filename"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}

// PullRequest fetches a pull request and every file it changes.
func (c *Client) PullRequest(ctx context.Context, repo domain.Repository, number int) (domain.PullRequest, error) {
	var meta apiPullRequest
	if err := c.get(ctx, fmt.Sprintf("/repos/%s/pulls/%d", repo, number), &meta); err != nil {
		return domain.PullRequest{}, fmt.Errorf("read pull request %s#%d: %w", repo, number, err)
	}

	files, err := c.files(ctx, repo, number)
	if err != nil {
		return domain.PullRequest{}, fmt.Errorf("read files of %s#%d: %w", repo, number, err)
	}

	return domain.PullRequest{
		Repository: repo,
		Number:     number,
		Title:      meta.Title,
		Author:     meta.User.Login,
		URL:        meta.HTMLURL,
		Files:      files,
	}, nil
}

func (c *Client) files(ctx context.Context, repo domain.Repository, number int) ([]domain.ChangedFile, error) {
	var files []domain.ChangedFile
	for page := 1; page <= maxFilePages; page++ {
		var batch []apiFile
		endpoint := fmt.Sprintf("/repos/%s/pulls/%d/files?per_page=%d&page=%d", repo, number, filesPerPage, page)
		if err := c.get(ctx, endpoint, &batch); err != nil {
			return nil, err
		}
		for _, f := range batch {
			files = append(files, domain.ChangedFile{Path: f.Filename, Additions: f.Additions, Deletions: f.Deletions})
		}
		if len(batch) < filesPerPage {
			break
		}
	}
	return files, nil
}

func (c *Client) get(ctx context.Context, endpoint string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+endpoint, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("call GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub answered %s for %s", resp.Status, endpoint)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode GitHub response: %w", err)
	}
	return nil
}
