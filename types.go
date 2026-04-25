package polymarket

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Market is one side of a binary prediction (or one option of a
// multi-outcome event).
//
// Field names mirror Gamma's responses. Numeric values that Gamma
// encodes as JSON strings (e.g. volume24hr may arrive as "1234.56") are
// decoded via [flexFloat]. The three "JSON-inside-JSON" array fields
// (outcomes, outcomePrices, clobTokenIds) are normalized by
// [Market.UnmarshalJSON] below.
type Market struct {
	ID                    string    `json:"id"`
	Question              string    `json:"question,omitempty"`
	Slug                  string    `json:"slug,omitempty"`
	ConditionID           string    `json:"conditionId,omitempty"`
	QuestionID            string    `json:"questionID,omitempty"`
	Description           string    `json:"description,omitempty"`
	Active                bool      `json:"active"`
	Closed                bool      `json:"closed"`
	Archived              bool      `json:"archived,omitempty"`
	AcceptingOrders       bool      `json:"acceptingOrders,omitempty"`
	EnableOrderBook       bool      `json:"enableOrderBook,omitempty"`
	MinimumOrderSize      flexFloat `json:"minimumOrderSize,omitempty"`
	MinimumTickSize       flexFloat `json:"minimumTickSize,omitempty"`
	OrderPriceMinTickSize flexFloat `json:"orderPriceMinTickSize,omitempty"`
	OrderMinSize          flexFloat `json:"orderMinSize,omitempty"`
	Volume                flexFloat `json:"volume,omitempty"`
	Volume24Hr            flexFloat `json:"volume24hr,omitempty"`
	Volume1Wk             flexFloat `json:"volume1wk,omitempty"`
	Volume1Mo             flexFloat `json:"volume1mo,omitempty"`
	Liquidity             flexFloat `json:"liquidity,omitempty"`
	LiquidityNum          flexFloat `json:"liquidityNum,omitempty"`
	StartDate             time.Time `json:"startDate,omitempty"`
	EndDate               time.Time `json:"endDate,omitempty"`
	CreatedAt             time.Time `json:"createdAt,omitempty"`
	UpdatedAt             time.Time `json:"updatedAt,omitempty"`
	ResolvedBy            string    `json:"resolvedBy,omitempty"`
	Resolution            string    `json:"resolution,omitempty"`
	Outcomes              []string  `json:"-"`
	OutcomePrices         []float64 `json:"-"`
	ClobTokenIDs          []string  `json:"-"`
	NegRisk               bool      `json:"negRisk,omitempty"`
	UMAResolutionStatus   string    `json:"umaResolutionStatus,omitempty"`
	ResolutionSource      string    `json:"resolutionSource,omitempty"`
	EventSlug             string    `json:"eventSlug,omitempty"`
	EventID               string    `json:"eventId,omitempty"`
	Icon                  string    `json:"icon,omitempty"`
	Image                 string    `json:"image,omitempty"`
}

// UnmarshalJSON decodes a Gamma market. Polymarket historically stored
// outcomes/outcomePrices/clobTokenIds as JSON strings inside the JSON
// document (e.g. `"outcomes": "[\"Yes\",\"No\"]"`). Newer responses
// sometimes emit real arrays; we accept either shape.
func (m *Market) UnmarshalJSON(data []byte) error {
	type alias Market
	aux := struct {
		*alias
		RawOutcomes      json.RawMessage `json:"outcomes,omitempty"`
		RawOutcomePrices json.RawMessage `json:"outcomePrices,omitempty"`
		RawClobTokenIDs  json.RawMessage `json:"clobTokenIds,omitempty"`
	}{alias: (*alias)(m)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if outs, err := decodeStringArray(aux.RawOutcomes); err != nil {
		return fmt.Errorf("outcomes: %w", err)
	} else {
		m.Outcomes = outs
	}
	if prices, err := decodeFloatArray(aux.RawOutcomePrices); err != nil {
		return fmt.Errorf("outcomePrices: %w", err)
	} else {
		m.OutcomePrices = prices
	}
	if ids, err := decodeStringArray(aux.RawClobTokenIDs); err != nil {
		return fmt.Errorf("clobTokenIds: %w", err)
	} else {
		m.ClobTokenIDs = ids
	}
	return nil
}

// TokenID returns the CLOB asset id for the Yes outcome (index 0), or
// empty when the market hasn't exposed tokens yet.
func (m Market) TokenID() string {
	if len(m.ClobTokenIDs) == 0 {
		return ""
	}
	return m.ClobTokenIDs[0]
}

// TokenIDFor returns the CLOB asset id for outcome i (0 = Yes, 1 = No for
// binaries). Empty string when missing.
func (m Market) TokenIDFor(outcome int) string {
	if outcome < 0 || outcome >= len(m.ClobTokenIDs) {
		return ""
	}
	return m.ClobTokenIDs[outcome]
}

// PriceFor returns the outcome price at index i, or 0 when missing.
func (m Market) PriceFor(outcome int) float64 {
	if outcome < 0 || outcome >= len(m.OutcomePrices) {
		return 0
	}
	return m.OutcomePrices[outcome]
}

// Event groups one or more related Markets (e.g. "Who will win the 2028
// election?" with a market per candidate). Market[] is populated when
// the caller requested it (default on /events endpoints).
type Event struct {
	ID           string    `json:"id"`
	Slug         string    `json:"slug,omitempty"`
	Title        string    `json:"title,omitempty"`
	Description  string    `json:"description,omitempty"`
	Category     string    `json:"category,omitempty"`
	Active       bool      `json:"active"`
	Closed       bool      `json:"closed"`
	Archived     bool      `json:"archived,omitempty"`
	Featured     bool      `json:"featured,omitempty"`
	Volume       flexFloat `json:"volume,omitempty"`
	Volume24Hr   flexFloat `json:"volume24hr,omitempty"`
	Volume1Wk    flexFloat `json:"volume1wk,omitempty"`
	Liquidity    flexFloat `json:"liquidity,omitempty"`
	StartDate    time.Time `json:"startDate,omitempty"`
	EndDate      time.Time `json:"endDate,omitempty"`
	CreatedAt    time.Time `json:"createdAt,omitempty"`
	UpdatedAt    time.Time `json:"updatedAt,omitempty"`
	Icon         string    `json:"icon,omitempty"`
	Image        string    `json:"image,omitempty"`
	Markets      []Market  `json:"markets,omitempty"`
	Tags         []Tag     `json:"tags,omitempty"`
	CommentCount int       `json:"commentCount,omitempty"`
}

// Tag is a category / sport / topic label used to filter markets and
// events. IDs are small integers encoded as strings by Gamma; Label is
// human-readable.
type Tag struct {
	ID        string `json:"id"`
	Label     string `json:"label,omitempty"`
	Slug      string `json:"slug,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// OrderbookLevel is a single bid or ask rung.
type OrderbookLevel struct {
	Price flexFloat `json:"price"`
	Size  flexFloat `json:"size"`
}

// OrderbookSummary is the /book response from the CLOB API.
//
// Bids are sorted price-descending, asks price-ascending. Timestamp is
// the server-reported time (the CLOB API returns it as a numeric string,
// decoded here as seconds since epoch).
type OrderbookSummary struct {
	Market         string           `json:"market"`
	AssetID        string           `json:"asset_id"`
	Timestamp      flexInt          `json:"timestamp"`
	Hash           string           `json:"hash,omitempty"`
	Bids           []OrderbookLevel `json:"bids"`
	Asks           []OrderbookLevel `json:"asks"`
	MinOrderSize   flexFloat        `json:"min_order_size"`
	TickSize       flexFloat        `json:"tick_size"`
	NegRisk        bool             `json:"neg_risk"`
	LastTradePrice flexFloat        `json:"last_trade_price"`
}

// BestBid returns the highest bid price, or 0 if there are none.
func (o OrderbookSummary) BestBid() float64 {
	if len(o.Bids) == 0 {
		return 0
	}
	return o.Bids[0].Price.Float()
}

// BestAsk returns the lowest ask price, or 0 if there are none.
func (o OrderbookSummary) BestAsk() float64 {
	if len(o.Asks) == 0 {
		return 0
	}
	return o.Asks[0].Price.Float()
}

// Midpoint is the average of the best bid and ask. Returns 0 when the
// book is one-sided or empty.
func (o OrderbookSummary) Midpoint() float64 {
	bb, ba := o.BestBid(), o.BestAsk()
	if bb == 0 || ba == 0 {
		return 0
	}
	return (bb + ba) / 2
}

// Spread returns ask-bid; 0 if either side is empty.
func (o OrderbookSummary) Spread() float64 {
	bb, ba := o.BestBid(), o.BestAsk()
	if bb == 0 || ba == 0 {
		return 0
	}
	return ba - bb
}

// PriceResponse is the /price response. Side is "BUY" or "SELL".
type PriceResponse struct {
	Price flexFloat `json:"price"`
}

// MidpointResponse is the /midpoint response.
type MidpointResponse struct {
	Mid flexFloat `json:"mid"`
}

// flexInt is an int64 that accepts either a JSON number or a JSON
// string containing a number (the CLOB API returns timestamps as
// stringified numbers). MarshalJSON emits a plain number.
type flexInt int64

// Int returns the underlying int64.
func (f flexInt) Int() int64 { return int64(f) }

func (f *flexInt) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	if data[0] == '"' {
		s, err := strconv.Unquote(string(data))
		if err != nil {
			return err
		}
		if s == "" {
			return nil
		}
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		*f = flexInt(v)
		return nil
	}
	var v int64
	if err := json.Unmarshal(data, &v); err != nil {
		var vf float64
		if err2 := json.Unmarshal(data, &vf); err2 == nil {
			*f = flexInt(int64(vf))
			return nil
		}
		return err
	}
	*f = flexInt(v)
	return nil
}

func (f flexInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(f))
}

// flexFloat is a float64 that accepts either a JSON number or a JSON
// string containing a number. Polymarket's various APIs mix the two
// inconsistently even for the same field (e.g. "0.52" vs 0.52).
// MarshalJSON always emits a plain number so downstream consumers can
// treat it as a normal float.
type flexFloat float64

// Float returns the underlying float64.
func (f flexFloat) Float() float64 { return float64(f) }

func (f *flexFloat) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}
	if data[0] == '"' {
		s, err := strconv.Unquote(string(data))
		if err != nil {
			return err
		}
		if s == "" {
			return nil
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		*f = flexFloat(v)
		return nil
	}
	var v float64
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*f = flexFloat(v)
	return nil
}

func (f flexFloat) MarshalJSON() ([]byte, error) {
	return json.Marshal(float64(f))
}

// decodeStringArray accepts three Gamma shapes: (1) a real JSON array of
// strings; (2) a JSON string containing a JSON array of strings
// (historic); (3) null / empty. Unknown shapes error.
func decodeStringArray(raw json.RawMessage) ([]string, error) {
	raw = stripJSONWS(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	if raw[0] == '[' {
		var out []string
		if err := json.Unmarshal(raw, &out); err != nil {
			return nil, err
		}
		return out, nil
	}
	if raw[0] == '"' {
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, err
		}
		if s == "" {
			return nil, nil
		}
		var out []string
		if err := json.Unmarshal([]byte(s), &out); err != nil {
			return nil, err
		}
		return out, nil
	}
	return nil, fmt.Errorf("unsupported JSON shape for string array: %s", string(raw))
}

// decodeFloatArray handles "[0.5, 0.5]", "[\"0.5\", \"0.5\"]", and the
// string-wrapped forms of both.
func decodeFloatArray(raw json.RawMessage) ([]float64, error) {
	raw = stripJSONWS(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var inner json.RawMessage = raw
	if raw[0] == '"' {
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, err
		}
		if s == "" {
			return nil, nil
		}
		inner = json.RawMessage(s)
	}
	var strs []string
	if err := json.Unmarshal(inner, &strs); err == nil {
		out := make([]float64, 0, len(strs))
		for _, s := range strs {
			if s == "" {
				out = append(out, 0)
				continue
			}
			v, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return out, nil
	}
	var floats []float64
	if err := json.Unmarshal(inner, &floats); err != nil {
		return nil, err
	}
	return floats, nil
}

func stripJSONWS(raw json.RawMessage) json.RawMessage {
	for len(raw) > 0 {
		switch raw[0] {
		case ' ', '\n', '\r', '\t':
			raw = raw[1:]
		default:
			return raw
		}
	}
	return raw
}
