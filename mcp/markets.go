package polymarketmcp

import (
	"context"

	"github.com/teslashibe/mcptool"
	polymarket "github.com/teslashibe/polymarket-go"
)

// ListMarketsInput is the typed input for polymarket_list_markets.
type ListMarketsInput struct {
	Active     bool   `json:"active,omitempty" jsonschema:"description=include only active markets (default true),default=true"`
	Closed     bool   `json:"closed,omitempty" jsonschema:"description=include closed/resolved markets,default=false"`
	TagID      string `json:"tag_id,omitempty" jsonschema:"description=filter to a single Polymarket tag id"`
	Order      string `json:"order,omitempty" jsonschema:"description=sort key (volume_24hr|volume|liquidity|end_date|start_date),default=volume_24hr"`
	Ascending  bool   `json:"ascending,omitempty" jsonschema:"description=sort ascending (default false=highest first)"`
	Limit      int    `json:"limit,omitempty" jsonschema:"description=max markets to return,minimum=1,maximum=50,default=10"`
	EndDateMin string `json:"end_date_min,omitempty" jsonschema:"description=RFC3339 lower bound on resolution date"`
	EndDateMax string `json:"end_date_max,omitempty" jsonschema:"description=RFC3339 upper bound on resolution date"`
}

// MarketSummary is the token-efficient response shape for list
// endpoints. Prose-heavy fields (description) are omitted; a separate
// polymarket_get_market returns the full record.
type MarketSummary struct {
	ID              string    `json:"id"`
	Slug            string    `json:"slug,omitempty"`
	Question        string    `json:"question,omitempty"`
	ConditionID     string    `json:"condition_id,omitempty"`
	Active          bool      `json:"active"`
	Closed          bool      `json:"closed"`
	AcceptingOrders bool      `json:"accepting_orders,omitempty"`
	EnableOrderBook bool      `json:"enable_order_book,omitempty"`
	Volume24Hr      float64   `json:"volume_24hr,omitempty"`
	Liquidity       float64   `json:"liquidity,omitempty"`
	EndDate         string    `json:"end_date,omitempty"`
	Outcomes        []string  `json:"outcomes,omitempty"`
	OutcomePrices   []float64 `json:"outcome_prices,omitempty"`
	ClobTokenIDs    []string  `json:"clob_token_ids,omitempty"`
	NegRisk         bool      `json:"neg_risk,omitempty"`
}

func summarize(m polymarket.Market) MarketSummary {
	s := MarketSummary{
		ID:              m.ID,
		Slug:            m.Slug,
		Question:        mcptool.TruncateString(m.Question, 240),
		ConditionID:     m.ConditionID,
		Active:          m.Active,
		Closed:          m.Closed,
		AcceptingOrders: m.AcceptingOrders,
		EnableOrderBook: m.EnableOrderBook,
		Volume24Hr:      m.Volume24Hr.Float(),
		Liquidity:       m.Liquidity.Float(),
		Outcomes:        m.Outcomes,
		OutcomePrices:   m.OutcomePrices,
		ClobTokenIDs:    m.ClobTokenIDs,
		NegRisk:         m.NegRisk,
	}
	if !m.EndDate.IsZero() {
		s.EndDate = m.EndDate.UTC().Format("2006-01-02T15:04:05Z")
	}
	return s
}

func listMarkets(ctx context.Context, c *polymarket.Client, in ListMarketsInput) (any, error) {
	opts := polymarket.ListMarketsOpts{
		Active:     polymarket.Bool(defaultBool(in.Active, true)),
		Closed:     polymarket.Bool(in.Closed),
		TagID:      in.TagID,
		Order:      defaultString(in.Order, "volume_24hr"),
		Ascending:  in.Ascending,
		Limit:      defaultInt(in.Limit, 10),
		EndDateMin: in.EndDateMin,
		EndDateMax: in.EndDateMax,
	}
	markets, err := c.ListMarkets(ctx, opts)
	if err != nil {
		return nil, wrapErr(err, "list markets")
	}
	out := make([]MarketSummary, len(markets))
	for i, m := range markets {
		out[i] = summarize(m)
	}
	return mcptool.PageOf(out, "", opts.Limit), nil
}

// GetMarketInput is the typed input for polymarket_get_market.
type GetMarketInput struct {
	ID   string `json:"id,omitempty" jsonschema:"description=numeric Polymarket market id (prefer slug when possible)"`
	Slug string `json:"slug,omitempty" jsonschema:"description=URL slug from polymarket.com/event/<slug> (one of id or slug is required)"`
}

func getMarket(ctx context.Context, c *polymarket.Client, in GetMarketInput) (any, error) {
	switch {
	case in.Slug != "":
		m, err := c.GetMarketBySlug(ctx, in.Slug)
		if err != nil {
			return nil, wrapErr(err, "get market by slug")
		}
		return summarize(*m), nil
	case in.ID != "":
		m, err := c.GetMarket(ctx, in.ID)
		if err != nil {
			return nil, wrapErr(err, "get market by id")
		}
		return summarize(*m), nil
	default:
		return nil, &mcptool.Error{Code: "invalid_input", Message: "one of 'id' or 'slug' is required"}
	}
}

// ListEventsInput is the typed input for polymarket_list_events.
type ListEventsInput struct {
	Active     bool   `json:"active,omitempty" jsonschema:"description=include only active events,default=true"`
	Closed     bool   `json:"closed,omitempty" jsonschema:"description=include closed events,default=false"`
	TagID      string `json:"tag_id,omitempty" jsonschema:"description=filter to a single tag id"`
	Order      string `json:"order,omitempty" jsonschema:"description=sort key (volume_24hr|volume|liquidity|end_date),default=volume_24hr"`
	Ascending  bool   `json:"ascending,omitempty" jsonschema:"description=sort ascending"`
	Limit      int    `json:"limit,omitempty" jsonschema:"description=max events to return,minimum=1,maximum=50,default=10"`
	EndDateMin string `json:"end_date_min,omitempty" jsonschema:"description=RFC3339 lower bound on event resolution date"`
	EndDateMax string `json:"end_date_max,omitempty" jsonschema:"description=RFC3339 upper bound on event resolution date"`
}

// EventSummary trims the Event to the fields useful for discovery.
type EventSummary struct {
	ID         string          `json:"id"`
	Slug       string          `json:"slug,omitempty"`
	Title      string          `json:"title,omitempty"`
	Category   string          `json:"category,omitempty"`
	Active     bool            `json:"active"`
	Closed     bool            `json:"closed"`
	Volume24Hr float64         `json:"volume_24hr,omitempty"`
	Liquidity  float64         `json:"liquidity,omitempty"`
	EndDate    string          `json:"end_date,omitempty"`
	Markets    []MarketSummary `json:"markets,omitempty"`
}

func summarizeEvent(e polymarket.Event) EventSummary {
	s := EventSummary{
		ID:         e.ID,
		Slug:       e.Slug,
		Title:      mcptool.TruncateString(e.Title, 240),
		Category:   e.Category,
		Active:     e.Active,
		Closed:     e.Closed,
		Volume24Hr: e.Volume24Hr.Float(),
		Liquidity:  e.Liquidity.Float(),
	}
	if !e.EndDate.IsZero() {
		s.EndDate = e.EndDate.UTC().Format("2006-01-02T15:04:05Z")
	}
	if len(e.Markets) > 0 {
		s.Markets = make([]MarketSummary, len(e.Markets))
		for i, m := range e.Markets {
			s.Markets[i] = summarize(m)
		}
	}
	return s
}

func listEvents(ctx context.Context, c *polymarket.Client, in ListEventsInput) (any, error) {
	opts := polymarket.ListEventsOpts{
		Active:     polymarket.Bool(defaultBool(in.Active, true)),
		Closed:     polymarket.Bool(in.Closed),
		TagID:      in.TagID,
		Order:      defaultString(in.Order, "volume_24hr"),
		Ascending:  in.Ascending,
		Limit:      defaultInt(in.Limit, 10),
		EndDateMin: in.EndDateMin,
		EndDateMax: in.EndDateMax,
	}
	events, err := c.ListEvents(ctx, opts)
	if err != nil {
		return nil, wrapErr(err, "list events")
	}
	out := make([]EventSummary, len(events))
	for i, e := range events {
		out[i] = summarizeEvent(e)
	}
	return mcptool.PageOf(out, "", opts.Limit), nil
}

// GetEventInput is the typed input for polymarket_get_event.
type GetEventInput struct {
	ID   string `json:"id,omitempty" jsonschema:"description=numeric event id"`
	Slug string `json:"slug,omitempty" jsonschema:"description=event slug (one of id or slug is required)"`
}

func getEvent(ctx context.Context, c *polymarket.Client, in GetEventInput) (any, error) {
	switch {
	case in.Slug != "":
		e, err := c.GetEventBySlug(ctx, in.Slug)
		if err != nil {
			return nil, wrapErr(err, "get event by slug")
		}
		return summarizeEvent(*e), nil
	case in.ID != "":
		e, err := c.GetEvent(ctx, in.ID)
		if err != nil {
			return nil, wrapErr(err, "get event by id")
		}
		return summarizeEvent(*e), nil
	default:
		return nil, &mcptool.Error{Code: "invalid_input", Message: "one of 'id' or 'slug' is required"}
	}
}

// SearchInput is the typed input for polymarket_search.
type SearchInput struct {
	Query string `json:"query" jsonschema:"description=free-text search across events+markets,required"`
	Limit int    `json:"limit,omitempty" jsonschema:"description=max results per type,minimum=1,maximum=50,default=10"`
}

// SearchResponse is the token-efficient projection of the Gamma /public-search response.
type SearchResponse struct {
	Events  []EventSummary  `json:"events,omitempty"`
	Markets []MarketSummary `json:"markets,omitempty"`
}

func search(ctx context.Context, c *polymarket.Client, in SearchInput) (any, error) {
	res, err := c.Search(ctx, in.Query, defaultInt(in.Limit, 10))
	if err != nil {
		return nil, wrapErr(err, "search")
	}
	out := SearchResponse{
		Events:  make([]EventSummary, len(res.Events)),
		Markets: make([]MarketSummary, len(res.Markets)),
	}
	for i, e := range res.Events {
		out.Events[i] = summarizeEvent(e)
	}
	for i, m := range res.Markets {
		out.Markets[i] = summarize(m)
	}
	return out, nil
}

// ListTagsInput is the typed input for polymarket_list_tags.
type ListTagsInput struct{}

func listTags(ctx context.Context, c *polymarket.Client, _ ListTagsInput) (any, error) {
	tags, err := c.ListTags(ctx)
	if err != nil {
		return nil, wrapErr(err, "list tags")
	}
	return mcptool.PageOf(tags, "", 50), nil
}

var marketTools = []mcptool.Tool{
	mcptool.Define[*polymarket.Client, ListMarketsInput](
		"polymarket_list_markets",
		"List Polymarket markets filtered by status/tag/date and ordered by volume or liquidity.",
		"ListMarkets",
		listMarkets,
	),
	mcptool.Define[*polymarket.Client, GetMarketInput](
		"polymarket_get_market",
		"Get a single Polymarket market by numeric id or URL slug.",
		"GetMarket",
		getMarket,
	),
	mcptool.Define[*polymarket.Client, ListEventsInput](
		"polymarket_list_events",
		"List Polymarket events (bundles of related markets) with their child markets inlined.",
		"ListEvents",
		listEvents,
	),
	mcptool.Define[*polymarket.Client, GetEventInput](
		"polymarket_get_event",
		"Get a single Polymarket event (with all child markets) by numeric id or slug.",
		"GetEvent",
		getEvent,
	),
	mcptool.Define[*polymarket.Client, SearchInput](
		"polymarket_search",
		"Unified free-text search across Polymarket events and markets.",
		"Search",
		search,
	),
	mcptool.Define[*polymarket.Client, ListTagsInput](
		"polymarket_list_tags",
		"List Polymarket tags (categories/topics) usable as tag_id filters on list_markets / list_events.",
		"ListTags",
		listTags,
	),
}
