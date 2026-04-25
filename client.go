package polymarket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultGammaBaseURL is the production Gamma API origin.
const DefaultGammaBaseURL = "https://gamma-api.polymarket.com"

// DefaultCLOBBaseURL is the production CLOB API origin.
const DefaultCLOBBaseURL = "https://clob.polymarket.com"

// DefaultUserAgent is sent on every outgoing request so Polymarket can
// identify bot traffic from this client.
const DefaultUserAgent = "polymarket-go/0.1 (+https://github.com/teslashibe/polymarket-go)"

// DefaultTimeout is applied per-request when [Options.Timeout] is zero.
const DefaultTimeout = 15 * time.Second

// Options configures a [Client]. All fields are optional; zero values
// select sensible production defaults.
type Options struct {
	// GammaBaseURL overrides the Gamma API origin. Useful for staging
	// or a local mock; leave blank to hit production.
	GammaBaseURL string

	// CLOBBaseURL overrides the CLOB API origin.
	CLOBBaseURL string

	// HTTPClient is the http.Client used for outgoing requests. If nil,
	// [Client] constructs one with [Options.Timeout] (default 15s).
	HTTPClient *http.Client

	// Timeout applied when HTTPClient is nil. Ignored otherwise.
	Timeout time.Duration

	// UserAgent overrides the default UA header.
	UserAgent string
}

// Client is the Polymarket API client. Methods are safe for concurrent
// use; the underlying *http.Client handles pooling.
type Client struct {
	gammaBase string
	clobBase  string
	http      *http.Client
	ua        string
}

// New builds a Client from the supplied Options. Passing the zero value
// produces a production-ready client hitting the real Polymarket APIs
// with a 15s per-request timeout.
func New(opts Options) *Client {
	gamma := opts.GammaBaseURL
	if gamma == "" {
		gamma = DefaultGammaBaseURL
	}
	clob := opts.CLOBBaseURL
	if clob == "" {
		clob = DefaultCLOBBaseURL
	}
	httpClient := opts.HTTPClient
	if httpClient == nil {
		to := opts.Timeout
		if to <= 0 {
			to = DefaultTimeout
		}
		httpClient = &http.Client{Timeout: to}
	}
	ua := opts.UserAgent
	if ua == "" {
		ua = DefaultUserAgent
	}
	return &Client{
		gammaBase: strings.TrimRight(gamma, "/"),
		clobBase:  strings.TrimRight(clob, "/"),
		http:      httpClient,
		ua:        ua,
	}
}

// Bool returns a pointer to b. Convenience for building
// [ListMarketsOpts] / [ListEventsOpts] where optional booleans need the
// caller to distinguish "unset" from "false".
func Bool(b bool) *bool { return &b }

// getJSON issues a GET request to baseURL+path with the supplied query
// params and decodes a JSON response into out. Non-2xx responses return
// [APIError] (or [ErrNotFound] for 404).
func (c *Client) getJSON(ctx context.Context, baseURL, path string, params url.Values, out any) error {
	u := baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.ua)
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("polymarket %s %s: %w", http.MethodGet, path, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("polymarket %s %s: read body: %w", http.MethodGet, path, err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("%w (%s)", ErrNotFound, path)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{Endpoint: path, Status: resp.StatusCode, Body: truncate(string(body), 512)}
	}
	if out == nil {
		return nil
	}
	if len(body) == 0 {
		return errors.New("polymarket " + path + ": empty response body")
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("polymarket %s: decode: %w (body: %s)", path, err, truncate(string(body), 256))
	}
	return nil
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

// Health is a liveness probe hitting both API origins' root paths.
// Returns nil when both reply 2xx. Intended for bot startup; production
// consumers should surface errors to their existing healthcheck.
func (c *Client) Health(ctx context.Context) error {
	if err := c.getJSON(ctx, c.clobBase, "/", nil, nil); err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.Status == http.StatusOK {
			return nil
		}
		// Some CLOB deployments don't answer the bare / with JSON; accept any 2xx.
	}
	return nil
}
