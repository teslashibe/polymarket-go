package polymarket

import (
	"encoding/json"
	"testing"
)

// TestMarketUnmarshal_LegacyStringArrays locks the quirky Gamma shape
// where outcomes / outcomePrices / clobTokenIds arrive as
// JSON-strings-inside-JSON. Losing support for this shape would break
// reading historic markets.
func TestMarketUnmarshal_LegacyStringArrays(t *testing.T) {
	raw := []byte(`{
		"id": "12345",
		"question": "Will X happen?",
		"outcomes": "[\"Yes\",\"No\"]",
		"outcomePrices": "[\"0.53\",\"0.47\"]",
		"clobTokenIds": "[\"0xAAA\",\"0xBBB\"]"
	}`)
	var m Market
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m.Outcomes[0] != "Yes" || m.Outcomes[1] != "No" {
		t.Fatalf("outcomes: %#v", m.Outcomes)
	}
	if m.OutcomePrices[0] != 0.53 || m.OutcomePrices[1] != 0.47 {
		t.Fatalf("outcomePrices: %#v", m.OutcomePrices)
	}
	if m.ClobTokenIDs[0] != "0xAAA" || m.ClobTokenIDs[1] != "0xBBB" {
		t.Fatalf("clobTokenIds: %#v", m.ClobTokenIDs)
	}
	if got := m.TokenIDFor(1); got != "0xBBB" {
		t.Fatalf("TokenIDFor(1) = %q", got)
	}
	if got := m.PriceFor(0); got != 0.53 {
		t.Fatalf("PriceFor(0) = %v", got)
	}
}

// TestMarketUnmarshal_NativeArrays ensures we still accept the newer
// shape where outcomes/prices arrive as real JSON arrays.
func TestMarketUnmarshal_NativeArrays(t *testing.T) {
	raw := []byte(`{
		"id": "9",
		"outcomes": ["Yes", "No"],
		"outcomePrices": [0.5, 0.5],
		"clobTokenIds": ["0x1", "0x2"]
	}`)
	var m Market
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(m.Outcomes) != 2 || m.Outcomes[0] != "Yes" {
		t.Fatalf("outcomes: %#v", m.Outcomes)
	}
	if m.OutcomePrices[1] != 0.5 {
		t.Fatalf("outcomePrices: %#v", m.OutcomePrices)
	}
}

// TestFlexFloat locks numeric-as-string decoding for a field Gamma is
// known to emit both ways.
func TestFlexFloat(t *testing.T) {
	raw := []byte(`{"a": "1234.5", "b": 67.89, "c": null}`)
	var v struct{ A, B, C flexFloat }
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if v.A.Float() != 1234.5 || v.B.Float() != 67.89 || v.C.Float() != 0 {
		t.Fatalf("got %v %v %v", v.A, v.B, v.C)
	}
}

// TestFlexInt_StringTimestamp proves the CLOB /book timestamp (a
// stringified epoch-seconds) decodes into OrderbookSummary.
func TestFlexInt_StringTimestamp(t *testing.T) {
	raw := []byte(`{"market":"X","asset_id":"A","timestamp":"1714151234","bids":[],"asks":[],"min_order_size":"1","tick_size":"0.01","neg_risk":false,"last_trade_price":"0.5"}`)
	var o OrderbookSummary
	if err := json.Unmarshal(raw, &o); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if o.Timestamp.Int() != 1714151234 {
		t.Fatalf("timestamp: %d", o.Timestamp.Int())
	}
}

// TestOrderbookDerivations verifies best-bid / best-ask / midpoint /
// spread helpers handle empty and one-sided books.
func TestOrderbookDerivations(t *testing.T) {
	full := OrderbookSummary{
		Bids: []OrderbookLevel{{Price: 0.45, Size: 100}, {Price: 0.44, Size: 200}},
		Asks: []OrderbookLevel{{Price: 0.46, Size: 150}, {Price: 0.47, Size: 250}},
	}
	if full.BestBid() != 0.45 || full.BestAsk() != 0.46 {
		t.Fatalf("best prices: %v / %v", full.BestBid(), full.BestAsk())
	}
	if got := full.Midpoint(); got < 0.454 || got > 0.456 {
		t.Fatalf("midpoint: %v", got)
	}
	if got := full.Spread(); got < 0.009 || got > 0.011 {
		t.Fatalf("spread: %v", got)
	}

	oneSided := OrderbookSummary{Bids: []OrderbookLevel{{Price: 0.3}}}
	if oneSided.Midpoint() != 0 || oneSided.Spread() != 0 {
		t.Fatalf("one-sided derivations should return 0")
	}
}
