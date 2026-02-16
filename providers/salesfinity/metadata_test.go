package salesfinity

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
	callLogsResponse := testutils.DataFromFile(t, "call-log.json")
	analyticsListPerformanceResponse := testutils.DataFromFile(t, "analytics-list-performance.json")
	analyticsSdrPerformanceResponse := testutils.DataFromFile(t, "analytics-sdr-performance.json")
	contactListsCsvResponse := testutils.DataFromFile(t, "contact-lists-csv.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe call-logs object by sampling first record from data array",
			Input: []string{"call-log"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/call-log"),
					mockcond.QueryParam("limit", "1"),
				},
				Then: mockserver.Response(http.StatusOK, callLogsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"call-log": {
						DisplayName: "Call-Log",
						Fields: map[string]common.FieldMetadata{
							"_id": {
								DisplayName:  "_id",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"to": {
								DisplayName:  "to",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"disposition": {
								DisplayName:  "disposition",
								ValueType:    common.ValueTypeOther,
								ProviderType: "",
								Values:       nil,
							},
							"contact": {
								DisplayName:  "contact",
								ValueType:    common.ValueTypeOther,
								ProviderType: "",
								Values:       nil,
							},
							"contact_list": {
								DisplayName:  "contact_list",
								ValueType:    common.ValueTypeOther,
								ProviderType: "",
								Values:       nil,
							},
							"user": {
								DisplayName:  "user",
								ValueType:    common.ValueTypeOther,
								ProviderType: "",
								Values:       nil,
							},
							"createdAt": {
								DisplayName:  "createdAt",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"updatedAt": {
								DisplayName:  "updatedAt",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"recording_url": {
								DisplayName:  "recording_url",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"transcription": {
								DisplayName:  "transcription",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"from": {
								DisplayName:  "from",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
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
		{
			Name:  "Successfully describe analytics/list-performance object by sampling first record from data array",
			Input: []string{"analytics/list-performance"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/analytics/list-performance"),
					mockcond.QueryParam("limit", "1"),
				},
				Then: mockserver.Response(http.StatusOK, analyticsListPerformanceResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"analytics/list-performance": {
						DisplayName: "Analytics/List-Performance",
						Fields: map[string]common.FieldMetadata{
							"_id": {
								DisplayName:  "_id",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"name": {
								DisplayName:  "name",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"source": {
								DisplayName:  "source",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"total_calls": {
								DisplayName:  "total_calls",
								ValueType:    common.ValueTypeFloat,
								ProviderType: "",
								Values:       nil,
							},
							"connected_calls": {
								DisplayName:  "connected_calls",
								ValueType:    common.ValueTypeFloat,
								ProviderType: "",
								Values:       nil,
							},
							"conversations": {
								DisplayName:  "conversations",
								ValueType:    common.ValueTypeFloat,
								ProviderType: "",
								Values:       nil,
							},
							"meetings_set": {
								DisplayName:  "meetings_set",
								ValueType:    common.ValueTypeFloat,
								ProviderType: "",
								Values:       nil,
							},
							"good_quality_contacts": {
								DisplayName:  "good_quality_contacts",
								ValueType:    common.ValueTypeFloat,
								ProviderType: "",
								Values:       nil,
							},
							"data_quality": {
								DisplayName:  "data_quality",
								ValueType:    common.ValueTypeFloat,
								ProviderType: "",
								Values:       nil,
							},
							"owner": {
								DisplayName:  "owner",
								ValueType:    common.ValueTypeOther,
								ProviderType: "",
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
		{
			Name:  "Successfully describe analytics/sdr-performance object by sampling first record from data array",
			Input: []string{"analytics/sdr-performance"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/analytics/sdr-performance"),
					mockcond.QueryParam("limit", "1"),
				},
				Then: mockserver.Response(http.StatusOK, analyticsSdrPerformanceResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"analytics/sdr-performance": {
						DisplayName: "Analytics/Sdr-Performance",
						Fields: map[string]common.FieldMetadata{
							"_id": {
								DisplayName:  "_id",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"name": {
								DisplayName:  "name",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"source": {
								DisplayName:  "source",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"total_calls": {
								DisplayName:  "total_calls",
								ValueType:    common.ValueTypeFloat,
								ProviderType: "",
								Values:       nil,
							},
							"connected_calls": {
								DisplayName:  "connected_calls",
								ValueType:    common.ValueTypeFloat,
								ProviderType: "",
								Values:       nil,
							},
							"conversations": {
								DisplayName:  "conversations",
								ValueType:    common.ValueTypeFloat,
								ProviderType: "",
								Values:       nil,
							},
							"meetings_set": {
								DisplayName:  "meetings_set",
								ValueType:    common.ValueTypeFloat,
								ProviderType: "",
								Values:       nil,
							},
							"good_quality_contacts": {
								DisplayName:  "good_quality_contacts",
								ValueType:    common.ValueTypeFloat,
								ProviderType: "",
								Values:       nil,
							},
							"data_quality": {
								DisplayName:  "data_quality",
								ValueType:    common.ValueTypeFloat,
								ProviderType: "",
								Values:       nil,
							},
							"owner": {
								DisplayName:  "owner",
								ValueType:    common.ValueTypeOther,
								ProviderType: "",
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
		{
			Name:  "Successfully describe contact-lists/csv object by sampling first record from data array",
			Input: []string{"contact-lists/csv"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/contact-lists/csv"),
					mockcond.QueryParam("limit", "1"),
				},
				Then: mockserver.Response(http.StatusOK, contactListsCsvResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contact-lists/csv": {
						DisplayName: "Contact-Lists/Csv",
						Fields: map[string]common.FieldMetadata{
							"_id": {
								DisplayName:  "_id",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"name": {
								DisplayName:  "name",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
								Values:       nil,
							},
							"user": {
								DisplayName:  "user",
								ValueType:    common.ValueTypeString,
								ProviderType: "",
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
		{
			Name:  "Returns error when server responds with 500",
			Input: []string{"call-log"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/call-log"),
					mockcond.QueryParam("limit", "1"),
				},
				Then: mockserver.Response(http.StatusInternalServerError),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"call-log": mockutils.ExpectedSubsetErrors{
						common.ErrServer,
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Returns error when response is empty",
			Input: []string{"call-log"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/call-log"),
					mockcond.QueryParam("limit", "1"),
				},
				Then: mockserver.Response(http.StatusOK, []byte("{}")),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"call-log": mockutils.ExpectedSubsetErrors{
						common.ErrMissingExpectedValues,
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Returns error when data array is empty",
			Input: []string{"call-log"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/call-log"),
					mockcond.QueryParam("limit", "1"),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{"data": []}`)),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"call-log": mockutils.ExpectedSubsetErrors{
						common.ErrMissingExpectedValues,
					},
				},
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

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
