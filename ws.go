package polymarket

import (
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const DefaultMarketWSURL = "wss://ws-subscriptions-clob.polymarket.com/ws/market"

type MarketStreamOptions struct {
	URL            string
	AssetIDs       []string
	UserAgent      string
	ReconnectDelay time.Duration
	ReadTimeout    time.Duration
	OnStatus       func(StreamStatus)
}

type MarketStreamer struct {
	opts MarketStreamOptions
}

type StreamStatus struct {
	State string
	Err   error
	At    time.Time
}

type WSEvent struct {
	EventType string           `json:"event_type"`
	AssetID   string           `json:"asset_id"`
	Market    string           `json:"market,omitempty"`
	Side      string           `json:"side,omitempty"`
	Price     float64          `json:"price,omitempty"`
	Bids      []OrderbookLevel `json:"bids,omitempty"`
	Asks      []OrderbookLevel `json:"asks,omitempty"`
	SourceTS  *time.Time       `json:"source_ts,omitempty"`
	Raw       json.RawMessage  `json:"raw"`
}

func NewMarketStreamer(opts MarketStreamOptions) *MarketStreamer {
	if opts.URL == "" {
		opts.URL = DefaultMarketWSURL
	}
	if opts.UserAgent == "" {
		opts.UserAgent = DefaultUserAgent
	}
	if opts.ReconnectDelay == 0 {
		opts.ReconnectDelay = 3 * time.Second
	}
	if opts.ReadTimeout == 0 {
		opts.ReadTimeout = 60 * time.Second
	}
	return &MarketStreamer{opts: opts}
}

func (s *MarketStreamer) Run(ctx context.Context, onEvent func(context.Context, WSEvent)) error {
	for ctx.Err() == nil {
		if err := s.read(ctx, onEvent); err != nil && ctx.Err() == nil {
			s.status("disconnected", err)
			slog.Warn("polymarket market ws reconnecting", "error", err)
			sleep(ctx, s.opts.ReconnectDelay)
		}
	}
	return ctx.Err()
}

func (s *MarketStreamer) read(ctx context.Context, onEvent func(context.Context, WSEvent)) error {
	dialer := websocket.Dialer{HandshakeTimeout: 15 * time.Second}
	conn, _, err := dialer.DialContext(ctx, s.opts.URL, http.Header{"User-Agent": {s.opts.UserAgent}})
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := conn.WriteJSON(map[string]any{"type": "market", "assets_ids": s.opts.AssetIDs}); err != nil {
		return err
	}
	s.status("connected", nil)
	slog.Info("polymarket market ws subscribed", "assets", len(s.opts.AssetIDs), "url", s.opts.URL)

	for ctx.Err() == nil {
		_ = conn.SetReadDeadline(time.Now().Add(s.opts.ReadTimeout))
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		for _, raw := range splitRawMessages(msg) {
			e := ParseWSEvent(raw)
			if e.AssetID != "" {
				slog.Debug("polymarket market ws event", "asset_id", e.AssetID, "event_type", e.EventType, "price", e.Price, "raw_bytes", len(e.Raw))
				onEvent(ctx, e)
			}
		}
	}
	return ctx.Err()
}

func (s *MarketStreamer) status(state string, err error) {
	if s.opts.OnStatus != nil {
		s.opts.OnStatus(StreamStatus{State: state, Err: err, At: time.Now().UTC()})
	}
}

func ParseWSEvent(raw []byte) WSEvent {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return WSEvent{}
	}
	e := WSEvent{
		EventType: firstString(m, "event_type", "type"),
		AssetID:   firstString(m, "asset_id", "assetId", "token_id", "tokenId"),
		Market:    firstString(m, "market", "market_id", "condition_id"),
		Side:      firstString(m, "side"),
		Price:     eventPrice(m),
		SourceTS:  eventTime(m),
		Raw:       append([]byte(nil), raw...),
	}
	e.Bids = levels(m["bids"])
	e.Asks = levels(m["asks"])
	return e
}

func splitRawMessages(msg []byte) []json.RawMessage {
	var arr []json.RawMessage
	if len(msg) > 0 && msg[0] == '[' && json.Unmarshal(msg, &arr) == nil {
		return arr
	}
	return []json.RawMessage{msg}
}

func eventPrice(m map[string]any) float64 {
	if v := firstFloat(m, "price", "last_trade_price", "last_price", "best_ask", "best_bid"); v > 0 {
		return v
	}
	bid, ask := bookEdge(m["bids"], true), bookEdge(m["asks"], false)
	if bid > 0 && ask > 0 {
		return (bid + ask) / 2
	}
	return math.Max(bid, ask)
}

func levels(v any) []OrderbookLevel {
	raw, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]OrderbookLevel, 0, len(raw))
	for _, item := range raw {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		out = append(out, OrderbookLevel{Price: flexFloat(firstFloat(obj, "price")), Size: flexFloat(firstFloat(obj, "size"))})
	}
	return out
}

func bookEdge(v any, bid bool) float64 {
	best := 0.0
	for _, level := range levels(v) {
		p := level.Price.Float()
		if p == 0 {
			continue
		}
		if best == 0 || (bid && p > best) || (!bid && p < best) {
			best = p
		}
	}
	return best
}

func eventTime(m map[string]any) *time.Time {
	for _, k := range []string{"timestamp", "ts", "time"} {
		if v, ok := m[k]; ok {
			if t, ok := parseWSTime(v); ok {
				return &t
			}
		}
	}
	return nil
}

func parseWSTime(v any) (time.Time, bool) {
	switch x := v.(type) {
	case float64:
		if x > 1e12 {
			return time.UnixMilli(int64(x)), true
		}
		return time.Unix(int64(x), 0), true
	case string:
		if n, err := strconv.ParseInt(x, 10, 64); err == nil {
			if n > 1e12 {
				return time.UnixMilli(n), true
			}
			return time.Unix(n, 0), true
		}
		if t, err := time.Parse(time.RFC3339, x); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func firstString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k].(string); ok {
			return v
		}
	}
	return ""
}

func firstFloat(m map[string]any, keys ...string) float64 {
	for _, k := range keys {
		switch v := m[k].(type) {
		case float64:
			return v
		case string:
			f, _ := strconv.ParseFloat(strings.TrimSpace(v), 64)
			return f
		}
	}
	return 0
}

func sleep(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}
