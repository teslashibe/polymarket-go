package polymarket

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const DefaultRTDSURL = "wss://ws-live-data.polymarket.com"

type RTDSOptions struct {
	URL            string
	UserAgent      string
	Symbols        []string
	ReconnectDelay time.Duration
	ReadTimeout    time.Duration
	PingInterval   time.Duration
	OnStatus       func(StreamStatus)
}

type RTDSStreamer struct {
	opts RTDSOptions
}

type RTDSPriceTick struct {
	Source string          `json:"source"`
	Symbol string          `json:"symbol"`
	Asset  string          `json:"asset"`
	TS     time.Time       `json:"ts"`
	Price  float64         `json:"price"`
	Raw    json.RawMessage `json:"raw"`
}

func NewRTDSStreamer(opts RTDSOptions) *RTDSStreamer {
	if opts.URL == "" {
		opts.URL = DefaultRTDSURL
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
	if opts.PingInterval == 0 {
		opts.PingInterval = 5 * time.Second
	}
	return &RTDSStreamer{opts: opts}
}

func (s *RTDSStreamer) Run(ctx context.Context, onTick func(context.Context, RTDSPriceTick)) error {
	for ctx.Err() == nil {
		if err := s.read(ctx, onTick); err != nil && ctx.Err() == nil {
			s.status("disconnected", err)
			slog.Warn("polymarket rtds reconnecting", "error", err)
			sleep(ctx, s.opts.ReconnectDelay)
		}
	}
	return ctx.Err()
}

func (s *RTDSStreamer) read(ctx context.Context, onTick func(context.Context, RTDSPriceTick)) error {
	dialer := websocket.Dialer{HandshakeTimeout: 15 * time.Second}
	conn, _, err := dialer.DialContext(ctx, s.opts.URL, http.Header{"User-Agent": {s.opts.UserAgent}})
	if err != nil {
		return err
	}
	defer conn.Close()

	symbols := s.chainlinkSymbols()
	subs := make([]map[string]any, 0, len(symbols))
	for _, symbol := range symbols {
		subs = append(subs, map[string]any{
			"topic":   "crypto_prices_chainlink",
			"type":    "*",
			"filters": `{"symbol":"` + symbol + `"}`,
		})
	}
	if err := conn.WriteJSON(map[string]any{
		"action":        "subscribe",
		"subscriptions": subs,
	}); err != nil {
		return err
	}
	s.status("connected", nil)
	slog.Info("polymarket rtds subscribed", "symbols", strings.Join(symbols, ","), "url", s.opts.URL)

	pingErr := s.startPing(ctx, conn)
	for ctx.Err() == nil {
		select {
		case err := <-pingErr:
			return err
		default:
		}
		_ = conn.SetReadDeadline(time.Now().Add(s.opts.ReadTimeout))
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		if string(msg) == "PONG" {
			continue
		}
		for _, tick := range ParseRTDSTicks(msg) {
			onTick(ctx, tick)
		}
	}
	return ctx.Err()
}

func (s *RTDSStreamer) chainlinkSymbols() []string {
	if len(s.opts.Symbols) == 0 {
		return []string{"btc/usd", "eth/usd", "sol/usd", "xrp/usd"}
	}
	out := make([]string, 0, len(s.opts.Symbols))
	seen := map[string]bool{}
	for _, symbol := range s.opts.Symbols {
		symbol = strings.ToLower(strings.TrimSpace(symbol))
		if symbol == "" || seen[symbol] {
			continue
		}
		seen[symbol] = true
		out = append(out, symbol)
	}
	return out
}

func (s *RTDSStreamer) startPing(ctx context.Context, conn *websocket.Conn) <-chan error {
	errCh := make(chan error, 1)
	ticker := time.NewTicker(s.opts.PingInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				if err := conn.WriteMessage(websocket.TextMessage, []byte("PING")); err != nil {
					select {
					case errCh <- err:
					default:
					}
					_ = conn.Close()
					return
				}
			}
		}
	}()
	return errCh
}

func (s *RTDSStreamer) status(state string, err error) {
	if s.opts.OnStatus != nil {
		s.opts.OnStatus(StreamStatus{State: state, Err: err, At: time.Now().UTC()})
	}
}

func ParseRTDSTicks(raw []byte) []RTDSPriceTick {
	var msg rtdsMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		return nil
	}
	if msg.Topic != "crypto_prices_chainlink" {
		return nil
	}
	if msg.Payload.Symbol != "" {
		tick, ok := rtdsPayloadTick(msg.Topic, msg.Payload, raw)
		if ok {
			return []RTDSPriceTick{tick}
		}
		return nil
	}
	out := make([]RTDSPriceTick, 0, len(msg.Payload.Data))
	for _, item := range msg.Payload.Data {
		if item.Symbol == "" {
			item.Symbol = msg.Payload.Symbol
		}
		tick, ok := rtdsPayloadTick(msg.Topic, item, raw)
		if ok {
			out = append(out, tick)
		}
	}
	return out
}

type rtdsMessage struct {
	Topic     string      `json:"topic"`
	Type      string      `json:"type"`
	Timestamp int64       `json:"timestamp"`
	Payload   rtdsPayload `json:"payload"`
}

type rtdsPayload struct {
	Symbol    string        `json:"symbol"`
	Timestamp int64         `json:"timestamp"`
	Value     float64       `json:"value"`
	Data      []rtdsPayload `json:"data"`
}

func rtdsPayloadTick(topic string, payload rtdsPayload, raw []byte) (RTDSPriceTick, bool) {
	if payload.Symbol == "" || payload.Value <= 0 {
		return RTDSPriceTick{}, false
	}
	ts := time.Now().UTC()
	if payload.Timestamp > 0 {
		ts = time.UnixMilli(payload.Timestamp).UTC()
	}
	return RTDSPriceTick{
		Source: "chainlink",
		Symbol: strings.ToLower(payload.Symbol),
		Asset:  assetFromChainlinkSymbol(payload.Symbol),
		TS:     ts,
		Price:  payload.Value,
		Raw:    append([]byte(nil), raw...),
	}, true
}

func assetFromChainlinkSymbol(symbol string) string {
	base := strings.Split(strings.ToUpper(strings.TrimSpace(symbol)), "/")
	if len(base) == 0 {
		return ""
	}
	return base[0]
}
