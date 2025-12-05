package dropboxsign

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	bulkJobsResponse := testutils.DataFromFile(t, "read-bulk_send_job.json")
	faxResponse := testutils.DataFromFile(t, "read-fax.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe multiple objects with metadata",
			Input: []string{"bulk_send_job", "fax"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/v3/bulk_send_job/list"),
					Then: mockserver.Response(http.StatusOK, bulkJobsResponse),
				}, {
					If:   mockcond.Path("/v3/fax/list"),
					Then: mockserver.Response(http.StatusOK, faxResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"bulk_send_job": {
						DisplayName: "Bulk_send_job",
						Fields: map[string]common.FieldMetadata{
							"bulk_send_job_id": {
								DisplayName:  "bulk_send_job_id",
								ValueType:    "string",
								ProviderType: "string",
								Values:       nil,
							},
							"total": {
								DisplayName:  "total",
								ValueType:    "float",
								ProviderType: "float",
								Values:       nil,
							},
							"is_creator": {
								DisplayName:  "is_creator",
								ValueType:    "boolean",
								ProviderType: "boolean",
								Values:       nil,
							},
							"created_at": {
								DisplayName:  "created_at",
								ValueType:    "float",
								ProviderType: "float",
								Values:       nil,
							},
						},
						FieldsMap: nil,
					},
					"fax": {
						DisplayName: "Fax",
						Fields: map[string]common.FieldMetadata{
							"fax_id": {
								DisplayName:  "fax_id",
								ValueType:    "string",
								ProviderType: "string",
								Values:       nil,
							},
							"title": {
								DisplayName:  "title",
								ValueType:    "string",
								ProviderType: "string",
								Values:       nil,
							},
							"original_title": {
								DisplayName:  "original_title",
								ValueType:    "string",
								ProviderType: "string",
								Values:       nil,
							},
							"subject": {
								DisplayName:  "subject",
								ValueType:    "string",
								ProviderType: "string",
								Values:       nil,
							},
							"message": {
								DisplayName:  "message",
								ValueType:    "string",
								ProviderType: "string",
								Values:       nil,
							},
							"metadata": {
								DisplayName:  "metadata",
								ValueType:    "other",
								ProviderType: "other",
								Values:       nil,
							},
							"created_at": {
								DisplayName:  "created_at",
								ValueType:    "float",
								ProviderType: "float",
								Values:       nil,
							},
							"sender": {
								DisplayName:  "sender",
								ValueType:    "string",
								ProviderType: "string",
								Values:       nil,
							},
							"transmissions": {
								DisplayName:  "transmissions",
								ValueType:    "other",
								ProviderType: "other",
								Values:       nil,
							},
							"files_url": {
								DisplayName:  "files_url",
								ValueType:    "string",
								ProviderType: "string",
								Values:       nil,
							},
						},
						FieldsMap: nil,
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
