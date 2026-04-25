package polymarket

import (
	"context"
	"net/url"
)

// GetOrderbook returns the current orderbook summary for a CLOB token
// (i.e. one outcome of one market). tokenID is the CLOB asset id —
// available on [Market.ClobTokenIDs] or via [Market.TokenIDFor].
func (c *Client) GetOrderbook(ctx context.Context, tokenID string) (*OrderbookSummary, error) {
	q := url.Values{}
	q.Set("token_id", tokenID)
	var out OrderbookSummary
	if err := c.getJSON(ctx, c.clobBase, "/book", q, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetPrice fetches the current best-price for one side of a token.
// side must be "BUY" or "SELL" (case-sensitive upstream).
func (c *Client) GetPrice(ctx context.Context, tokenID, side string) (float64, error) {
	q := url.Values{}
	q.Set("token_id", tokenID)
	q.Set("side", side)
	var out PriceResponse
	if err := c.getJSON(ctx, c.clobBase, "/price", q, &out); err != nil {
		return 0, err
	}
	return out.Price.Float(), nil
}

// GetMidpoint fetches the current midpoint price for a token (average of
// best bid and best ask). Convenient as a single-number "what's this
// market priced at?" without the full orderbook.
func (c *Client) GetMidpoint(ctx context.Context, tokenID string) (float64, error) {
	q := url.Values{}
	q.Set("token_id", tokenID)
	var out MidpointResponse
	if err := c.getJSON(ctx, c.clobBase, "/midpoint", q, &out); err != nil {
		return 0, err
	}
	return out.Mid.Float(), nil
}
