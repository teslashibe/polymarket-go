// Package polymarketmcp exposes github.com/teslashibe/polymarket-go as a
// set of mcptool.Tool values backing a single mcptool.Provider. The
// tools cover Polymarket's public Gamma API (markets / events / search)
// and CLOB read endpoints (orderbook / price / midpoint).
//
// Registration into an agent-setup style harness is a single line in the
// harness's mcp/platforms wiring file.
package polymarketmcp

import "github.com/teslashibe/mcptool"

// Provider implements mcptool.Provider for Polymarket. Zero value is
// ready to use.
type Provider struct{}

// Platform returns "polymarket". Tool names are prefixed accordingly
// (polymarket_list_markets, polymarket_get_orderbook, …).
func (Provider) Platform() string { return "polymarket" }

// Tools returns every polymarket_* tool exposed by this provider. Order
// is cosmetic; the host registry sorts by name.
func (Provider) Tools() []mcptool.Tool {
	out := make([]mcptool.Tool, 0, 10)
	out = append(out, marketTools...)
	out = append(out, orderbookTools...)
	return out
}
