// Package polymarket is a Go client for Polymarket's public data APIs.
//
// It wraps two surfaces:
//
//   - Gamma API (https://gamma-api.polymarket.com) — public read API for
//     markets, events, tags, and search. No authentication required.
//   - CLOB API (https://clob.polymarket.com) — order book, prices, and
//     (eventually) order placement. Read endpoints are public; write
//     endpoints require L2 API-key auth derived from an L1 wallet
//     signature on Polygon.
//
// The write path is stubbed in v0. Every order-placement / cancellation
// method returns [ErrWriteNotImplemented]. Fill them in only after the
// consuming application has wired up a signing key store and replayable
// idempotency guards — Polymarket trades are real money on-chain.
//
// The mcp subpackage exposes every public client method as an
// mcptool.Tool so bindings into agent-setup-style harnesses are a
// one-line registration.
package polymarket
