package polymarket

import "context"

// Side is "BUY" or "SELL". Exported as an untyped string alias so tool
// schemas stay simple; validated by the upstream CLOB at submission
// time.
const (
	SideBuy  = "BUY"
	SideSell = "SELL"
)

// OrderType names the matching discipline applied to a resting order.
// GTC = good-'til-cancelled (default). FOK = fill-or-kill. GTD =
// good-'til-date.
const (
	OrderTypeGTC = "GTC"
	OrderTypeFOK = "FOK"
	OrderTypeGTD = "GTD"
)

// Order is the (stubbed) canonical shape of a Polymarket CLOB order.
// The live path will need to sign an EIP-712 typed-data struct over
// this payload with the wallet's L1 key before POSTing to /order.
type Order struct {
	// ClientOrderID is a caller-generated UUID used for idempotency on
	// retries. Required.
	ClientOrderID string

	// TokenID is the CLOB asset id (one outcome of one market).
	TokenID string

	// Side is "BUY" or "SELL".
	Side string

	// Price is in USDC (0 < price < 1 for binary outcomes).
	Price float64

	// Size is the number of shares (1 share = $1 at resolution).
	Size float64

	// Type is "GTC", "FOK", or "GTD".
	Type string

	// Expiration is required when Type == "GTD"; Unix seconds.
	Expiration int64
}

// OrderResponse is the (stubbed) shape returned by POST /order.
type OrderResponse struct {
	OrderID string
	Status  string
}

// PlaceOrder submits an order to the CLOB.
//
// STUBBED in v0: always returns [ErrWriteNotImplemented]. See the
// package docstring for the rationale. Consumers should route writes
// through their own executor abstraction (paper vs live) and only
// activate this path after wiring an EIP-712 signer.
func (c *Client) PlaceOrder(ctx context.Context, o Order) (*OrderResponse, error) {
	_ = ctx
	_ = o
	return nil, ErrWriteNotImplemented
}

// CancelOrder cancels a resting order by CLOB order id.
//
// STUBBED in v0.
func (c *Client) CancelOrder(ctx context.Context, orderID string) error {
	_ = ctx
	_ = orderID
	return ErrWriteNotImplemented
}

// CancelAll cancels every open order for the authenticated caller.
//
// STUBBED in v0.
func (c *Client) CancelAll(ctx context.Context) error {
	_ = ctx
	return ErrWriteNotImplemented
}

// OpenOrder describes a resting order owned by the authenticated caller.
type OpenOrder struct {
	OrderID    string
	TokenID    string
	Side       string
	Price      float64
	Size       float64
	SizeFilled float64
	Status     string
	CreatedAt  int64
}

// GetOpenOrders lists the authenticated caller's resting orders.
//
// STUBBED in v0.
func (c *Client) GetOpenOrders(ctx context.Context) ([]OpenOrder, error) {
	_ = ctx
	return nil, ErrWriteNotImplemented
}

// Fill describes a single executed trade.
type Fill struct {
	TradeID    string
	OrderID    string
	TokenID    string
	Side       string
	Price      float64
	Size       float64
	ExecutedAt int64
	Fee        float64
}

// GetFills returns the authenticated caller's recent fills.
//
// STUBBED in v0.
func (c *Client) GetFills(ctx context.Context, sinceUnix int64) ([]Fill, error) {
	_ = ctx
	_ = sinceUnix
	return nil, ErrWriteNotImplemented
}
