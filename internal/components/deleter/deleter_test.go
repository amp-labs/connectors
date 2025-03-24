package deleter

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/mocked"
	"github.com/amp-labs/connectors/internal/components/operations"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestDelete(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	tests := []testroutines.Delete{
		{
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write object and its ID must be included",
			Input:        common.DeleteParams{ObjectName: "orders"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:     "Unknown object name is not supported",
			Input:    common.DeleteParams{ObjectName: "someUnknownObject", RecordId: "123"},
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

			tt.Run(t, func() (connectors.DeleteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

// Connector used to test HTTPDeleter.
type mockedConnector struct {
	mocked.Connector
	*HTTPDeleter
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
		HTTPDeleter: NewHTTPDeleter(
			connector.HTTPClient().Client,
			registry,
			staticschema.RootModuleID,
			operations.DeleteHandlers{},
		),
	}, nil
}
