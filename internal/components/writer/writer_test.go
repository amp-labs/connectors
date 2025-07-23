package writer

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

func TestWrite(t *testing.T) {
	t.Parallel()

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "orders"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:     "Unknown object name is not supported",
			Input:    common.WriteParams{ObjectName: "someUnknownObject", RecordData: "dummy"},
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

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

// Connector used to test HTTPWriter.
type mockedConnector struct {
	mocked.Connector
	*HTTPWriter
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
		HTTPWriter: NewHTTPWriter(
			connector.HTTPClient().Client,
			registry,
			common.ModuleRoot,
			operations.WriteHandlers{},
		),
	}, nil
}
