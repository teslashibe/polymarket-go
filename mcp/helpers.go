package polymarketmcp

import (
	"errors"

	"github.com/teslashibe/mcptool"
	polymarket "github.com/teslashibe/polymarket-go"
)

// wrapErr converts polymarket-package errors to structured mcptool.Error
// values the agent can reason about. Unknown errors are returned as-is
// so the MCP host treats them as `internal_error`.
func wrapErr(err error, op string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, polymarket.ErrNotFound) {
		return &mcptool.Error{Code: "not_found", Message: op + ": " + err.Error()}
	}
	if errors.Is(err, polymarket.ErrWriteNotImplemented) {
		return &mcptool.Error{Code: "write_disabled", Message: op + ": live CLOB writes are not wired; use the paper executor"}
	}
	var apiErr *polymarket.APIError
	if errors.As(err, &apiErr) {
		code := "upstream_error"
		if apiErr.Status == 429 {
			code = "rate_limited"
		} else if apiErr.Status >= 500 {
			code = "upstream_error"
		} else if apiErr.Status >= 400 {
			code = "invalid_input"
		}
		return &mcptool.Error{Code: code, Message: op + ": " + apiErr.Error()}
	}
	return err
}

func defaultInt(v, fallback int) int {
	if v <= 0 {
		return fallback
	}
	return v
}

func defaultString(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func defaultBool(v, fallback bool) bool {
	// Callers pass `in.Active` (zero = false) but want the documented
	// default when the field was omitted. This helper preserves the
	// "omitempty → default" semantic by only substituting when v is
	// the zero value; callers who want literal `false` must set the
	// field to `true` and invert in their own handler.
	if !v {
		return fallback
	}
	return v
}
