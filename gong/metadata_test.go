package gong

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/tools/scrapper"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Unknown object requested",
			Input:        []string{"butterflies"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{scrapper.ErrObjectNotFound},
		},
		{
			Name:   "Successfully describe one object with metadata",
			Input:  []string{"flows"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"flows": {
						DisplayName: "List Gong Engage flows (/v2/flows)",
						FieldsMap: map[string]string{
							"id":           "id",
							"name":         "name",
							"folderId":     "folderId",
							"folderName":   "folderName",
							"visibility":   "visibility",
							"creationDate": "creationDate",
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:   "Successfully describe multiple objects with metadata",
			Input:  []string{"workspaces", "logs"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"workspaces": {
						DisplayName: "List all company workspaces (/v2/workspaces)",
						FieldsMap: map[string]string{
							"id":          "id",
							"name":        "name",
							"description": "description",
						},
					},
					"logs": {
						DisplayName: "Retrieve logs data by type and time range (/v2/logs)",
						FieldsMap: map[string]string{
							"userId":                   "userId",
							"userEmailAddress":         "userEmailAddress",
							"userFullName":             "userFullName",
							"impersonatorUserId":       "impersonatorUserId",
							"impersonatorEmailAddress": "impersonatorEmailAddress",
							"impersonatorFullName":     "impersonatorFullName",
							"impersonatorCompanyId":    "impersonatorCompanyId",
							"eventTime":                "eventTime",
							"logRecord":                "logRecord",
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
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
