# polymarket-go

Go client for the [Polymarket](https://polymarket.com) prediction market, plus an `mcp/`
subpackage that exposes it as a set of [mcptool](https://github.com/teslashibe/mcptool)
tools for use with [agent-setup](https://github.com/teslashibe/agent-setup) derived
harnesses (e.g. [polybot](https://github.com/teslashibe/polybot)).

The client wraps two Polymarket surfaces:

- **Gamma API** (`https://gamma-api.polymarket.com`) — public read API for markets,
  events, tags, search. No authentication.
- **CLOB API** (`https://clob.polymarket.com`) — orderbook, prices, and (eventually)
  order placement. Read endpoints are public; write endpoints require L2 auth with
  API keys derived from an L1 wallet signature.

## Status

- Read path (Gamma + CLOB reads): **complete**.
- Write path (order placement, cancel, allowances): **stubbed**. Calling any
  write method returns `ErrWriteNotImplemented`. Filling these in requires
  EIP-712 typed-data signing against a Polygon wallet and is deliberately left
  out of v0 so the consuming bot can't accidentally move money.

## Quick start

```go
import polymarket "github.com/teslashibe/polymarket-go"

c := polymarket.New(polymarket.Options{})

markets, err := c.ListMarkets(ctx, polymarket.ListMarketsOpts{
    Active:    polymarket.Bool(true),
    Closed:    polymarket.Bool(false),
    Order:     "volume_24hr",
    Ascending: false,
    Limit:     50,
})
```

## MCP

```go
import polymarketmcp "github.com/teslashibe/polymarket-go/mcp"

provider := polymarketmcp.Provider{}
_ = provider.Tools() // []mcptool.Tool — feed into your MCP registry.
```

The Cursor rule `.cursor/rules/mcp-tool-conventions.mdc` in `agent-setup` and
[teslashibe/mcptool](https://github.com/teslashibe/mcptool) describe the full
convention set.

## License

MIT.
