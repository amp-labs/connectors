package gotoconn

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

const testAccountKey = "8967235839898"

func TestListObjectMetadata(t *testing.T) { //nolint:funlen
	t.Parallel()

	webinarsResponse := testutils.DataFromFile(t, "webinars.json")
	sessionsResponse := testutils.DataFromFile(t, "sessions.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Webinars metadata is sampled from organizer-scoped endpoint",
			Input: []string{"webinars"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/G2W/rest/v2/organizers/" + testAccountKey + "/webinars"),
					mockcond.QueryParam("size", "1"),
				},
				Then: mockserver.Response(200, webinarsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"webinars": {
						DisplayName: "Webinars",
						Fields: map[string]common.FieldMetadata{
							"webinarKey":          {DisplayName: "webinarKey", ValueType: common.ValueTypeString, ProviderType: "string"},
							"subject":             {DisplayName: "subject", ValueType: common.ValueTypeString, ProviderType: "string"},
							"numberOfRegistrants": {DisplayName: "numberOfRegistrants", ValueType: common.ValueTypeFloat, ProviderType: "float"},
							"inSession":           {DisplayName: "inSession", ValueType: common.ValueTypeBoolean, ProviderType: "boolean"},
							"times":               {DisplayName: "times", ValueType: common.ValueTypeOther, ProviderType: "other"},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Sessions metadata is sampled from GoToAssist extended sessions endpoint",
			Input: []string{"sessions"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/G2A/rest/v1/extendedsessions"),
					mockcond.QueryParam("size", "1"),
				},
				Then: mockserver.Response(200, sessionsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"sessions": {
						DisplayName: "Sessions",
						Fields: map[string]common.FieldMetadata{
							"sessionId":   {DisplayName: "sessionId", ValueType: common.ValueTypeString, ProviderType: "string"},
							"sessionType": {DisplayName: "sessionType", ValueType: common.ValueTypeString, ProviderType: "string"},
							"status":      {DisplayName: "status", ValueType: common.ValueTypeString, ProviderType: "string"},
							"expertName":  {DisplayName: "expertName", ValueType: common.ValueTypeString, ProviderType: "string"},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		//nolint:varnamelen
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
		Metadata: map[string]string{
			"accountKey": testAccountKey,
		},
		Module: providers.ModuleGoTo,
	})
	if err != nil {
		return nil, err
	}

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
