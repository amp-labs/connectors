package salesforce

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/providers"
)

func TestSubscribeConnector(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	conn, err := constructTestConnector("example.com", providers.ModuleSalesforceStandard)
	if err != nil {
		t.Fatalf("failed to construct test connector: %v", err)
	}

	_, ok := any(conn).(connectors.SubscribeConnector)
	if !ok {
		t.Fatalf("expected SubscribeConnector, got %T", conn)
	}
}
