package atlassian

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseIssueSchema := testutils.DataFromFile(t, "issue-metadata.json")

	tests := []testroutines.Metadata{
		{
			Name:  "Server response must include array",
			Input: []string{"butterflies"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, ""),
			}.Server(),
			ExpectedErrs: []error{common.ErrEmptyJSONHTTPResponse},
		},
		{
			Name:  "Server response must have at least one field",
			Input: []string{"butterflies"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `[]`),
			}.Server(),
			ExpectedErrs: []error{
				ErrMissingMetadata,
				ErrParsingMetadata,
			},
		},
		{
			Name:  "Field response must have identifier",
			Input: []string{"butterflies"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `[{}]`),
			}.Server(),
			ExpectedErrs: []error{ErrParsingMetadata},
		},
		{
			Name:  "Field response must have display name",
			Input: []string{"butterflies"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `[{"id": "issuerestriction"}]`),
			}.Server(),
			ExpectedErrs: []error{ErrParsingMetadata},
		},
		{
			Name:  "Successfully describe Issue metadata",
			Input: []string{},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseIssueSchema),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"issue": {
						DisplayName: "Issue",
						FieldsMap: map[string]string{
							// Manually attached fields:
							"id": "Id",
							// Fields coming from server response:
							"issuekey":                      "Key",
							"priority":                      "Priority",
							"creator":                       "Creator",
							"worklog":                       "Log Work",
							"issuetype":                     "Issue Type",
							"issuelinks":                    "Linked Issues",
							"fixVersions":                   "Fix versions",
							"issuerestriction":              "Restrict to",
							"statuscategorychangedate":      "Status Category Changed",
							"aggregatetimeoriginalestimate": "Σ Original Estimate",
							"customfield_10028":             "Submitted forms",
							"customfield_10035":             "Project overview key",
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func TestListObjectMetadataWithoutMetadata(t *testing.T) {
	t.Parallel()

	connector, err := NewConnector(
		WithAuthenticatedClient(http.DefaultClient),
		WithWorkspace("test-workspace"),
		WithModule(ModuleJira),
	)
	if err != nil {
		t.Fatal("failed to create connector")
	}

	_, err = connector.ListObjectMetadata(context.Background(), nil)
	if !errors.Is(err, ErrMissingCloudId) {
		t.Fatalf("expected ListObjectMetadata method to complain about missing cloud id")
	}
}
