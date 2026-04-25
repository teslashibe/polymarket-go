package polymarketmcp_test

import (
	"reflect"
	"testing"

	"github.com/teslashibe/mcptool"
	polymarket "github.com/teslashibe/polymarket-go"
	polymarketmcp "github.com/teslashibe/polymarket-go/mcp"
)

// TestEveryClientMethodIsWrappedOrExcluded is the canonical drift-detector
// enforced by the mcp-tool-conventions rule: every exported method on
// *polymarket.Client must either be wrapped by a Tool or listed in
// Excluded with a reason. Adding a new method without either breaks CI.
func TestEveryClientMethodIsWrappedOrExcluded(t *testing.T) {
	rep := mcptool.Coverage(
		reflect.TypeOf(&polymarket.Client{}),
		polymarketmcp.Provider{}.Tools(),
		polymarketmcp.Excluded,
	)
	if len(rep.Missing) > 0 {
		t.Fatalf("client methods missing MCP exposure (add a tool or list in excluded.go): %v", rep.Missing)
	}
	if len(rep.UnknownExclusions) > 0 {
		t.Fatalf("excluded.go references methods that don't exist on *Client: %v", rep.UnknownExclusions)
	}
}

// TestToolsValidate enforces naming + schema + description rules on every
// tool. Runs the same ValidateTools helper the host harness runs.
func TestToolsValidate(t *testing.T) {
	if err := mcptool.ValidateTools(polymarketmcp.Provider{}.Tools()); err != nil {
		t.Fatal(err)
	}
}

// TestProviderPlatform locks the platform identifier. Changing it would
// break every host that stores credentials keyed by platform id.
func TestProviderPlatform(t *testing.T) {
	if got := (polymarketmcp.Provider{}).Platform(); got != "polymarket" {
		t.Fatalf("Platform() = %q, want %q", got, "polymarket")
	}
}
