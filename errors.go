package polymarket

import (
	"errors"
	"fmt"
)

// ErrWriteNotImplemented is returned by every write method in v0. The
// consuming application must opt into live trading deliberately; see the
// package doc for the reasoning.
var ErrWriteNotImplemented = errors.New("polymarket: write path not implemented (use a paper executor or wire the CLOB signer)")

// ErrNotFound is returned when the upstream API responds 404 for a market,
// event, or token lookup. Callers can use errors.Is to branch.
var ErrNotFound = errors.New("polymarket: not found")

// APIError is the structured form of a non-2xx response from Polymarket.
// Status is the HTTP status code; Body is the raw error text the API
// returned (best-effort UTF-8). Endpoint names the path that was called
// so errors bubble up with context.
type APIError struct {
	Endpoint string
	Status   int
	Body     string
}

func (e *APIError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("polymarket %s: http %d", e.Endpoint, e.Status)
	}
	return fmt.Sprintf("polymarket %s: http %d: %s", e.Endpoint, e.Status, e.Body)
}
