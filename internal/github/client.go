// Package github provides a minimal GitHub REST API v3 client for the
// operations needed by the agent pipeline. Uses only stdlib — no third-party
// dependency on go-github or similar.
package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const baseURL = "https://api.github.com"

// Client is a thin GitHub REST v3 client authenticated with a personal access
// token (or fine-grained PAT / GitHub App installation token).
type Client struct {
	token      string
	httpClient *http.Client
}

// New returns a Client authenticated with the given token.
// An empty token is accepted but most operations will return 401.
func New(token string) *Client {
	return &Client{token: token, httpClient: &http.Client{}}
}

// ── Pull Requests ──────────────────────────────────────────────────────────

// CreatePRInput holds the fields for creating a pull request.
type CreatePRInput struct {
	Owner string
	Repo  string
	Title string
	Body  string
	Head  string // branch with changes
	Base  string // target branch (typically "main")
	Draft bool
}

// PR represents the relevant fields of a GitHub pull request.
type PR struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	HTMLURL string `json:"html_url"`
	State   string `json:"state"`
	// Head and Base contain branch information.
	Head struct{ Ref string } `json:"head"`
	Base struct{ Ref string } `json:"base"`
}

// CreatePR opens a new pull request and returns it.
func (c *Client) CreatePR(ctx context.Context, in CreatePRInput) (*PR, error) {
	body := map[string]any{
		"title": in.Title,
		"body":  in.Body,
		"head":  in.Head,
		"base":  in.Base,
		"draft": in.Draft,
	}

	var pr PR
	path := fmt.Sprintf("/repos/%s/%s/pulls", in.Owner, in.Repo)
	if err := c.do(ctx, http.MethodPost, path, body, &pr); err != nil {
		return nil, fmt.Errorf("create pr: %w", err)
	}
	return &pr, nil
}

// GetPR fetches a single pull request by number.
func (c *Client) GetPR(ctx context.Context, owner, repo string, number int) (*PR, error) {
	var pr PR
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, number)
	if err := c.do(ctx, http.MethodGet, path, nil, &pr); err != nil {
		return nil, fmt.Errorf("get pr: %w", err)
	}
	return &pr, nil
}

// ── helpers ────────────────────────────────────────────────────────────────

type apiError struct {
	Message string `json:"message"`
	Status  int
}

func (e *apiError) Error() string {
	return fmt.Sprintf("github API %d: %s", e.Status, e.Message)
}

func (c *Client) do(ctx context.Context, method, path string, reqBody any, out any) error {
	var bodyReader io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		var apiErr apiError
		_ = json.Unmarshal(rawBody, &apiErr)
		apiErr.Status = resp.StatusCode
		if apiErr.Message == "" {
			apiErr.Message = strings.TrimSpace(string(rawBody))
		}
		return &apiErr
	}

	if out != nil {
		return json.Unmarshal(rawBody, out)
	}
	return nil
}
