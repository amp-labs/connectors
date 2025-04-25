package instantlyAI

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	apiKeysResponse := testutils.DataFromFile(t, "api-keys.json")
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
			Input: []string{"api-keys", "custom-tags", "lead-lists"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.PathSuffix("v2/api-keys"),
					Then: mockserver.Response(http.StatusOK, apiKeysResponse),
				}, {
					If:   mockcond.PathSuffix("v2/custom-tags"),
					Then: mockserver.Response(http.StatusOK, customTagsResponse),
				}, {
					If:   mockcond.PathSuffix("v2/lead-lists"),
					Then: mockserver.Response(http.StatusOK, leadListsResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"api-keys": {
						DisplayName: "Api-keys",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"scopes": {
								DisplayName:  "scopes",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"key": {
								DisplayName:  "key",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"organization_id": {
								DisplayName:  "organization_id",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"timestamp_created": {
								DisplayName:  "timestamp_created",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"timestamp_updated": {
								DisplayName:  "timestamp_updated",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
						},
						FieldsMap: map[string]string{
							"id":                "id",
							"name":              "name",
							"scopes":            "scopes",
							"key":               "key",
							"organization_id":   "organization_id",
							"timestamp_created": "timestamp_created",
							"timestamp_updated": "timestamp_updated",
						},
					},
					"custom-tags": {
						DisplayName: "Custom-tags",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"timestamp_created": {
								DisplayName:  "timestamp_created",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"timestamp_updated": {
								DisplayName:  "timestamp_updated",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"organization_id": {
								DisplayName:  "organization_id",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"label": {
								DisplayName:  "label",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"description": {
								DisplayName:  "description",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
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
								ReadOnly:     false,
								Values:       nil,
							},
							"organization_id": {
								DisplayName:  "organization_id",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"has_enrichment_task": {
								DisplayName:  "has_enrichment_task",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"owned_by": {
								DisplayName:  "owned_by",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"timestamp_created": {
								DisplayName:  "timestamp_created",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
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
	connector, err := NewConnector(common.Parameters{
		Module:              common.ModuleRoot,
		AuthenticatedClient: http.DefaultClient,
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(serverURL)

	return connector, nil
}
