package surveymonkey

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { //nolint:funlen
	t.Parallel()

	createContact := testutils.DataFromFile(t, "write/contact-create.json")
	updateContact := testutils.DataFromFile(t, "write/contact-update.json")
	createContactList := testutils.DataFromFile(t, "write/contact-list-create.json")
	updateContactList := testutils.DataFromFile(t, "write/contact-list-update.json")
	createSurvey := testutils.DataFromFile(t, "write/survey-create.json")
	updateSurvey := testutils.DataFromFile(t, "write/survey-update.json")
	createSurveyFolder := testutils.DataFromFile(t, "write/survey-folder-create.json")
	updateContactField := testutils.DataFromFile(t, "write/contact-field-update.json")

	tests := []testconn.TestCaseWrite{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: objectContacts},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:         "Unknown object is not supported",
			Input:        common.WriteParams{ObjectName: objectGroups, RecordData: map[string]any{"name": "x"}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Create contact",
			Input: common.WriteParams{
				ObjectName: objectContacts,
				RecordData: map[string]any{
					"first_name": "Jane",
					"last_name":  "Doe",
					"email":      "jane.new@example.com",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v3/contacts"),
					mockcond.Body(`{"email":"jane.new@example.com","first_name":"Jane","last_name":"Doe"}`),
				},
				Then: mockserver.Response(http.StatusOK, createContact),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "5678",
				Data: map[string]any{
					"id":         "5678",
					"first_name": "Jane",
					"last_name":  "Doe",
					"email":      "jane.new@example.com",
				},
			},
		},
		{
			Name: "Update contact",
			Input: common.WriteParams{
				ObjectName: objectContacts,
				RecordId:   "1234",
				RecordData: map[string]any{
					"first_name": "Jane",
					"last_name":  "Updated",
					"email":      "jane.updated@example.com",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v3/contacts/1234"),
					mockcond.Body(`{"email":"jane.updated@example.com","first_name":"Jane","last_name":"Updated"}`),
				},
				Then: mockserver.Response(http.StatusOK, updateContact),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "1234",
				Data: map[string]any{
					"id":         "1234",
					"first_name": "Jane",
					"last_name":  "Updated",
					"email":      "jane.updated@example.com",
				},
			},
		},
		{
			Name: "Create contact list",
			Input: common.WriteParams{
				ObjectName: objectContactLists,
				RecordData: map[string]any{"name": "Amp Integration List"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v3/contact_lists"),
					mockcond.Body(`{"name":"Amp Integration List"}`),
				},
				Then: mockserver.Response(http.StatusOK, createContactList),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "9012",
				Data: map[string]any{
					"id":   "9012",
					"name": "Amp Integration List",
				},
			},
		},
		{
			Name: "Update contact list",
			Input: common.WriteParams{
				ObjectName: objectContactLists,
				RecordId:   "5678",
				RecordData: map[string]any{"name": "Amp Integration List (Updated)"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v3/contact_lists/5678"),
					mockcond.Body(`{"name":"Amp Integration List (Updated)"}`),
				},
				Then: mockserver.Response(http.StatusOK, updateContactList),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "5678",
				Data: map[string]any{
					"id":   "5678",
					"name": "Amp Integration List (Updated)",
				},
			},
		},
		{
			Name: "Create survey",
			Input: common.WriteParams{
				ObjectName: objectSurveys,
				RecordData: map[string]any{"title": "My Survey"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v3/surveys"),
					mockcond.Body(`{"title":"My Survey"}`),
				},
				Then: mockserver.Response(http.StatusOK, createSurvey),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "1234",
				Data: map[string]any{
					"id":    "1234",
					"title": "My Survey",
				},
			},
		},
		{
			Name: "Update survey",
			Input: common.WriteParams{
				ObjectName: objectSurveys,
				RecordId:   "1234",
				RecordData: map[string]any{"title": "My Survey (Updated)"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v3/surveys/1234"),
					mockcond.Body(`{"title":"My Survey (Updated)"}`),
				},
				Then: mockserver.Response(http.StatusOK, updateSurvey),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "1234",
				Data: map[string]any{
					"id":    "1234",
					"title": "My Survey (Updated)",
				},
			},
		},
		{
			Name: "Create survey folder",
			Input: common.WriteParams{
				ObjectName: objectSurveyFolders,
				RecordData: map[string]any{"title": "Amp Integration Folder"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v3/survey_folders"),
					mockcond.Body(`{"title":"Amp Integration Folder"}`),
				},
				Then: mockserver.Response(http.StatusOK, createSurveyFolder),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "5678",
				Data: map[string]any{
					"id":    "5678",
					"title": "Amp Integration Folder",
				},
			},
		},
		{
			Name: "Update survey folder is not supported",
			Input: common.WriteParams{
				ObjectName: objectSurveyFolders,
				RecordId:   "5678",
				RecordData: map[string]any{"title": "Updated"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Update contact field",
			Input: common.WriteParams{
				ObjectName: objectContactFields,
				RecordId:   "1",
				RecordData: map[string]any{"label": "Custom 1 (Updated)"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/v3/contact_fields/1"),
					mockcond.Body(`{"label":"Custom 1 (Updated)"}`),
				},
				Then: mockserver.Response(http.StatusOK, updateContactField),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "1",
				Data: map[string]any{
					"id":    "1",
					"label": "Custom 1 (Updated)",
				},
			},
		},
		{
			Name: "Create contact field is not supported",
			Input: common.WriteParams{
				ObjectName: objectContactFields,
				RecordData: map[string]any{"label": "New Field"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableWriter, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
