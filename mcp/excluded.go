package polymarketmcp

// Excluded lists *polymarket.Client methods that are intentionally not
// exposed as MCP tools. The value is a one-line human reason the
// coverage test in mcp_test.go reports when a new method is added
// without either a tool wrapping it or an entry here.
//
// Keep this list tight: when an exclusion is no longer justified, add a
// tool and delete the entry.
var Excluded = map[string]string{
	// polymarket_get_market dispatches by id OR slug inside a single
	// tool, so WrapsMethod can only list GetMarket. Same pattern for
	// events.
	"GetMarketBySlug": "covered by polymarket_get_market (dispatches on id or slug)",
	"GetEventBySlug":  "covered by polymarket_get_event (dispatches on id or slug)",
	"ListAllMarkets":  "helper for callers that page through polymarket_list_markets",

	// Liveness probe — host's /mcp/v1/health route handles readiness.
	"Health": "liveness probe owned by the host application",

	// Write path is stubbed in v0; exposing it through MCP would let
	// the agent call it and receive a confusing "not implemented"
	// every time. Ship the tools alongside the live path in a future
	// release.
	"PlaceOrder":    "write path stubbed; route writes through the trader's paper/live executor",
	"CancelOrder":   "write path stubbed",
	"CancelAll":     "write path stubbed",
	"GetOpenOrders": "write path stubbed — caller's order book comes from the trader's local DB",
	"GetFills":      "write path stubbed — fills are recorded by the trader's executor",
}
