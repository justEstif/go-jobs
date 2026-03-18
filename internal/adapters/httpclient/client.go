// Package httpclient implements driving-port adapters that call the go-jobs
// /api/v1/ JSON API over HTTP.
//
// Use this package when the CLI is targeting a remote (or local) go-jobs server
// instead of running in-process against a direct database connection.
//
// Construct the three adapters via their New* constructors, all of which accept
// a *Client that holds the base URL, auth token, and underlying http.Client.
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Client is the shared HTTP transport used by all httpclient adapters.
// It holds the base URL of the go-jobs server and the current Bearer token.
//
// The token may be empty for operations that do not require authentication
// (register, login, unauthenticated search). All protected operations will
// return an error if the token is empty.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient constructs a Client targeting baseURL.
// token is the opaque session token stored by `go-jobs login`; pass an empty
// string when constructing a client for unauthenticated operations only.
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL:    baseURL,
		token:      token,
		httpClient: &http.Client{},
	}
}

// get performs a GET request to path (relative to /api/v1/) and decodes the
// JSON response body into out. bearer controls whether the Authorization header
// is included.
func (c *Client) get(ctx context.Context, path string, query url.Values, bearer bool, out any) error {
	u := c.baseURL + "/api/v1/" + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if bearer {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return c.do(req, out)
}

// post performs a POST request to path with a JSON body and decodes the
// response into out (may be nil if no response body is expected).
func (c *Client) post(ctx context.Context, path string, bearer bool, body, out any) error {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		r = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/"+path, r)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if bearer {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return c.do(req, out)
}

// do executes req, checks for HTTP errors, and decodes a JSON body into out.
// out may be nil; in that case only the status code is checked.
func (c *Client) do(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return decodeAPIError(resp)
	}

	if out != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

// apiError is the shape returned by the server on errors: {"error": "message"}.
type apiError struct {
	Message string `json:"error"`
}

func decodeAPIError(resp *http.Response) error {
	var e apiError
	_ = json.NewDecoder(resp.Body).Decode(&e)
	if e.Message != "" {
		return fmt.Errorf("server error %d: %s", resp.StatusCode, e.Message)
	}
	return fmt.Errorf("server error %d", resp.StatusCode)
}
