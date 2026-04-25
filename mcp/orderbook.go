package polymarketmcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	polymarket "github.com/teslashibe/polymarket-go"
)

// GetOrderbookInput is the typed input for polymarket_get_orderbook.
type GetOrderbookInput struct {
	TokenID string `json:"token_id" jsonschema:"description=CLOB asset id for one market outcome (find via polymarket_get_market / list_markets),required"`
}

// OrderbookView is the token-efficient orderbook projection returned to
// the agent. Full depth is often 50+ levels per side — we cap at top-5
// so a scan across 20 markets stays under the MCP response byte cap.
type OrderbookView struct {
	Market         string          `json:"market"`
	AssetID        string          `json:"asset_id"`
	Timestamp      int64           `json:"timestamp"`
	BestBid        float64         `json:"best_bid,omitempty"`
	BestAsk        float64         `json:"best_ask,omitempty"`
	Midpoint       float64         `json:"midpoint,omitempty"`
	Spread         float64         `json:"spread,omitempty"`
	LastTradePrice float64         `json:"last_trade_price,omitempty"`
	MinOrderSize   float64         `json:"min_order_size,omitempty"`
	TickSize       float64         `json:"tick_size,omitempty"`
	NegRisk        bool            `json:"neg_risk,omitempty"`
	Bids           []priceSize     `json:"bids,omitempty"`
	Asks           []priceSize     `json:"asks,omitempty"`
}

type priceSize struct {
	Price float64 `json:"price"`
	Size  float64 `json:"size"`
}

func summarizeBook(o polymarket.OrderbookSummary) OrderbookView {
	view := OrderbookView{
		Market:         o.Market,
		AssetID:        o.AssetID,
		Timestamp:      o.Timestamp.Int(),
		BestBid:        o.BestBid(),
		BestAsk:        o.BestAsk(),
		Midpoint:       o.Midpoint(),
		Spread:         o.Spread(),
		LastTradePrice: o.LastTradePrice.Float(),
		MinOrderSize:   o.MinOrderSize.Float(),
		TickSize:       o.TickSize.Float(),
		NegRisk:        o.NegRisk,
	}
	const topN = 5
	for i, b := range o.Bids {
		if i >= topN {
			break
		}
		view.Bids = append(view.Bids, priceSize{Price: b.Price.Float(), Size: b.Size.Float()})
	}
	for i, a := range o.Asks {
		if i >= topN {
			break
		}
		view.Asks = append(view.Asks, priceSize{Price: a.Price.Float(), Size: a.Size.Float()})
	}
	return view
}

func getOrderbook(ctx context.Context, c *polymarket.Client, in GetOrderbookInput) (any, error) {
	if in.TokenID == "" {
		return nil, &mcptool.Error{Code: "invalid_input", Message: "token_id is required"}
	}
	ob, err := c.GetOrderbook(ctx, in.TokenID)
	if err != nil {
		return nil, wrapErr(err, "get orderbook")
	}
	return summarizeBook(*ob), nil
}

// GetPriceInput is the typed input for polymarket_get_price.
type GetPriceInput struct {
	TokenID string `json:"token_id" jsonschema:"description=CLOB asset id,required"`
	Side    string `json:"side" jsonschema:"description=which side's top-of-book price (BUY or SELL),enum=BUY,enum=SELL,required"`
}

func getPrice(ctx context.Context, c *polymarket.Client, in GetPriceInput) (any, error) {
	if in.TokenID == "" || in.Side == "" {
		return nil, &mcptool.Error{Code: "invalid_input", Message: "token_id and side are required"}
	}
	p, err := c.GetPrice(ctx, in.TokenID, in.Side)
	if err != nil {
		return nil, wrapErr(err, "get price")
	}
	return map[string]any{"price": p, "side": in.Side, "token_id": in.TokenID}, nil
}

// GetMidpointInput is the typed input for polymarket_get_midpoint.
type GetMidpointInput struct {
	TokenID string `json:"token_id" jsonschema:"description=CLOB asset id,required"`
}

func getMidpoint(ctx context.Context, c *polymarket.Client, in GetMidpointInput) (any, error) {
	if in.TokenID == "" {
		return nil, &mcptool.Error{Code: "invalid_input", Message: "token_id is required"}
	}
	m, err := c.GetMidpoint(ctx, in.TokenID)
	if err != nil {
		return nil, wrapErr(err, "get midpoint")
	}
	return map[string]any{"midpoint": m, "token_id": in.TokenID}, nil
}

var orderbookTools = []mcptool.Tool{
	mcptool.Define[*polymarket.Client, GetOrderbookInput](
		"polymarket_get_orderbook",
		"Get the top-of-book (top 5 bids / asks) for a Polymarket token with best_bid / best_ask / midpoint / spread derived.",
		"GetOrderbook",
		getOrderbook,
	),
	mcptool.Define[*polymarket.Client, GetPriceInput](
		"polymarket_get_price",
		"Get the current BUY or SELL top-of-book price for a Polymarket token.",
		"GetPrice",
		getPrice,
	),
	mcptool.Define[*polymarket.Client, GetMidpointInput](
		"polymarket_get_midpoint",
		"Get the current midpoint price (avg of best bid / ask) for a Polymarket token.",
		"GetMidpoint",
		getMidpoint,
	),
}
