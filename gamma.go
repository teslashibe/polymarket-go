package polymarket

import (
	"context"
	"net/url"
	"strconv"
)

// ListMarketsOpts filters / sorts a Gamma /markets listing.
//
// Active / Closed are pointers so callers can distinguish "unset" from
// "false" (Polymarket defaults to closed=false, so leaving Closed nil
// means "live markets only" — use Bool(true) to fetch resolved markets).
type ListMarketsOpts struct {
	Active    *bool
	Closed    *bool
	Archived  *bool
	TagID     string
	Slug      string
	Order     string
	Ascending bool
	Limit     int
	Offset    int
	// EndDateMin / EndDateMax filter by market resolution date. RFC3339.
	EndDateMin string
	EndDateMax string
}

// ListEventsOpts mirrors [ListMarketsOpts] for the /events endpoint.
type ListEventsOpts struct {
	Active       *bool
	Closed       *bool
	Archived     *bool
	Featured     *bool
	TagID        string
	Slug         string
	Order        string
	Ascending    bool
	Limit        int
	Offset       int
	RelatedTags  bool
	ExcludeTagID string
	EndDateMin   string
	EndDateMax   string
}

// ListMarkets returns markets matching opts. The default ordering is
// volume_24hr descending — the single most useful ordering for a
// trading bot.
func (c *Client) ListMarkets(ctx context.Context, opts ListMarketsOpts) ([]Market, error) {
	q := url.Values{}
	if opts.Active != nil {
		q.Set("active", strconv.FormatBool(*opts.Active))
	}
	if opts.Closed != nil {
		q.Set("closed", strconv.FormatBool(*opts.Closed))
	}
	if opts.Archived != nil {
		q.Set("archived", strconv.FormatBool(*opts.Archived))
	}
	if opts.TagID != "" {
		q.Set("tag_id", opts.TagID)
	}
	if opts.Slug != "" {
		q.Set("slug", opts.Slug)
	}
	if opts.Order != "" {
		q.Set("order", opts.Order)
	}
	q.Set("ascending", strconv.FormatBool(opts.Ascending))
	if opts.Limit > 0 {
		q.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		q.Set("offset", strconv.Itoa(opts.Offset))
	}
	if opts.EndDateMin != "" {
		q.Set("end_date_min", opts.EndDateMin)
	}
	if opts.EndDateMax != "" {
		q.Set("end_date_max", opts.EndDateMax)
	}
	var out []Market
	if err := c.getJSON(ctx, c.gammaBase, "/markets", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ListAllMarkets(ctx context.Context, opts ListMarketsOpts) ([]Market, error) {
	pageSize := opts.Limit
	if pageSize <= 0 {
		pageSize = 500
	}
	var all []Market
	for offset := 0; ; offset += pageSize {
		opts.Limit = pageSize
		opts.Offset = offset
		page, err := c.ListMarkets(ctx, opts)
		if err != nil {
			return nil, err
		}
		all = append(all, page...)
		if len(page) < pageSize {
			return all, nil
		}
	}
}

// GetMarket returns a single market by numeric id.
func (c *Client) GetMarket(ctx context.Context, id string) (*Market, error) {
	var out Market
	if err := c.getJSON(ctx, c.gammaBase, "/markets/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetMarketBySlug returns a single market by slug (the path segment on
// polymarket.com URLs, e.g. "fed-decision-in-october").
func (c *Client) GetMarketBySlug(ctx context.Context, slug string) (*Market, error) {
	var out Market
	if err := c.getJSON(ctx, c.gammaBase, "/markets/slug/"+url.PathEscape(slug), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListEvents returns events matching opts. Events bundle related
// markets; this is the recommended entry point for broad market
// discovery because a single request returns multiple markets with full
// metadata.
func (c *Client) ListEvents(ctx context.Context, opts ListEventsOpts) ([]Event, error) {
	q := url.Values{}
	if opts.Active != nil {
		q.Set("active", strconv.FormatBool(*opts.Active))
	}
	if opts.Closed != nil {
		q.Set("closed", strconv.FormatBool(*opts.Closed))
	}
	if opts.Archived != nil {
		q.Set("archived", strconv.FormatBool(*opts.Archived))
	}
	if opts.Featured != nil {
		q.Set("featured", strconv.FormatBool(*opts.Featured))
	}
	if opts.TagID != "" {
		q.Set("tag_id", opts.TagID)
	}
	if opts.Slug != "" {
		q.Set("slug", opts.Slug)
	}
	if opts.Order != "" {
		q.Set("order", opts.Order)
	}
	q.Set("ascending", strconv.FormatBool(opts.Ascending))
	if opts.Limit > 0 {
		q.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		q.Set("offset", strconv.Itoa(opts.Offset))
	}
	if opts.RelatedTags {
		q.Set("related_tags", "true")
	}
	if opts.ExcludeTagID != "" {
		q.Set("exclude_tag_id", opts.ExcludeTagID)
	}
	if opts.EndDateMin != "" {
		q.Set("end_date_min", opts.EndDateMin)
	}
	if opts.EndDateMax != "" {
		q.Set("end_date_max", opts.EndDateMax)
	}
	var out []Event
	if err := c.getJSON(ctx, c.gammaBase, "/events", q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetEvent returns one event by numeric id.
func (c *Client) GetEvent(ctx context.Context, id string) (*Event, error) {
	var out Event
	if err := c.getJSON(ctx, c.gammaBase, "/events/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetEventBySlug returns one event by slug.
func (c *Client) GetEventBySlug(ctx context.Context, slug string) (*Event, error) {
	var out Event
	if err := c.getJSON(ctx, c.gammaBase, "/events/slug/"+url.PathEscape(slug), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SearchResult is the /public-search response shape.
type SearchResult struct {
	Events   []Event  `json:"events,omitempty"`
	Markets  []Market `json:"markets,omitempty"`
	Profiles []any    `json:"profiles,omitempty"`
}

// Search runs Polymarket's unified search across events, markets, and
// profiles. q is the free-text query. limit caps each result type.
func (c *Client) Search(ctx context.Context, q string, limit int) (*SearchResult, error) {
	params := url.Values{}
	params.Set("q", q)
	if limit > 0 {
		params.Set("limit_per_type", strconv.Itoa(limit))
	}
	var out SearchResult
	if err := c.getJSON(ctx, c.gammaBase, "/public-search", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListTags returns the set of tags callers can pass to [ListMarketsOpts.TagID] /
// [ListEventsOpts.TagID].
func (c *Client) ListTags(ctx context.Context) ([]Tag, error) {
	var out []Tag
	if err := c.getJSON(ctx, c.gammaBase, "/tags", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
