package reader

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/mocked"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestRead(t *testing.T) {
	t.Parallel()

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "orders"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:     "Unknown object name is not supported",
			Input:    common.ReadParams{ObjectName: "someUnknownObject", Fields: connectors.Fields("id")},
			Server:   mockserver.Dummy(),
			Expected: nil,
			ExpectedErrs: []error{
				common.ErrOperationNotSupportedForObject,
			},
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

// Connector used to test HTTPReader.
type mockedConnector struct {
	mocked.Connector
	*HTTPReader
}

func constructTestConnector(serverURL string) (*mockedConnector, error) {
	connector := mocked.Connector{
		BaseURL: serverURL,
	}

	registry, err := components.NewEndpointRegistry(nil)
	if err != nil {
		return nil, err
	}

	return &mockedConnector{
		Connector: connector,
		HTTPReader: NewHTTPReader(
			connector.HTTPClient().Client,
			registry,
			common.ModuleRoot,
			operations.ReadHandlers{},
		),
	}, nil
}
