package instantlyai

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

	backgroundJobsResponse := testutils.DataFromFile(t, "background-jobs.json")
	customTagsResponse := testutils.DataFromFile(t, "custom-tags.json")
	leadListsResponse := testutils.DataFromFile(t, "lead-lists.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe multiple objects with metadata",
			Input: []string{"background-jobs", "custom-tags", "lead-lists"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("v2/background-jobs"),
					Then: mockserver.Response(http.StatusOK, backgroundJobsResponse),
				}, {
					If:   mockcond.Path("v2/custom-tags"),
					Then: mockserver.Response(http.StatusOK, customTagsResponse),
				}, {
					If:   mockcond.Path("v2/lead-lists"),
					Then: mockserver.Response(http.StatusOK, leadListsResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"background-jobs": {
						DisplayName: "Background-jobs",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"workspace_id": {
								DisplayName:  "workspace_id",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"type": {
								DisplayName:  "type",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"entity_id": {
								DisplayName:  "entity_id",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"entity_type": {
								DisplayName:  "entity_type",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"data": {
								DisplayName:  "data",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"progress": {
								DisplayName:  "progress",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"status": {
								DisplayName:  "status",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"created_at": {
								DisplayName:  "created_at",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"updated_at": {
								DisplayName:  "updated_at",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
						},
						FieldsMap: map[string]string{
							"id":           "id",
							"workspace_id": "workspace_id",
							"type":         "type",
							"entity_id":    "entity_id",
							"entity_type":  "entity_type",
							"data":         "data",
							"progress":     "progress",
							"status":       "status",
							"created_at":   "created_at",
							"updated_at":   "updated_at",
						},
					},
					"custom-tags": {
						DisplayName: "Custom-tags",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"timestamp_created": {
								DisplayName:  "timestamp_created",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"timestamp_updated": {
								DisplayName:  "timestamp_updated",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"organization_id": {
								DisplayName:  "organization_id",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"label": {
								DisplayName:  "label",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"description": {
								DisplayName:  "description",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
						},
						FieldsMap: map[string]string{
							"id":                "id",
							"timestamp_created": "timestamp_created",
							"timestamp_updated": "timestamp_updated",
							"organization_id":   "organization_id",
							"label":             "label",
							"description":       "description",
						},
					},
					"lead-lists": {
						DisplayName: "Lead-lists",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"organization_id": {
								DisplayName:  "organization_id",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"has_enrichment_task": {
								DisplayName:  "has_enrichment_task",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"owned_by": {
								DisplayName:  "owned_by",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"timestamp_created": {
								DisplayName:  "timestamp_created",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
						},
						FieldsMap: map[string]string{
							"id":                  "id",
							"organization_id":     "organization_id",
							"has_enrichment_task": "has_enrichment_task",
							"owned_by":            "owned_by",
							"name":                "name",
							"timestamp_created":   "timestamp_created",
						},
					},
				},
				Errors: nil,
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
		Module:              common.ModuleRoot,
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
