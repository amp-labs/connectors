package phoneburner

import (
	"net/http"
	"sort"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseContacts := testutils.DataFromFile(t, "read/contacts.json")
	responseDialSessions := testutils.DataFromFile(t, "read/dialsession.json")
	responseFolders := testutils.DataFromFile(t, "read/folders.json")
	responseMembers := testutils.DataFromFile(t, "read/members.json")
	responseTags := testutils.DataFromFile(t, "read/tags.json")
	responseVoicemails := testutils.DataFromFile(t, "read/voicemails.json")
	responseUnauthorized := testutils.DataFromFile(t, "read/error-unauthorized.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "contacts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unknown objects are not supported",
			Input:        common.ReadParams{ObjectName: "tiger", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Read contacts",
			Input: common.ReadParams{ObjectName: "contacts", Fields: connectors.Fields("contact_user_id", "first_name", "last_name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/1/contacts"),
					mockcond.QueryParam("page_size", "100"),
					mockcond.QueryParam("page", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"contact_user_id": "30919237",
						"first_name":      "John",
						"last_name":       "Demo",
					},
					Raw: map[string]any{
						"contact_user_id": "30919237",
						"owner_id":        "13514766",
						"first_name":      "John",
						"last_name":       "Demo",
						"date_added":      "2023-10-15 10:38:07",
						"raw_phone":       "6025551234",
						"primary_email": map[string]any{
							"email": "john.demo@example.com",
						},
					},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Provider envelope error is mapped (200 with http_status=401)",
			Input: common.ReadParams{ObjectName: "contacts", Fields: connectors.Fields("contact_user_id")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/1/contacts"),
					mockcond.QueryParam("page_size", "100"),
					mockcond.QueryParam("page", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseUnauthorized),
			}.Server(),
			ExpectedErrs: []error{common.ErrAccessToken},
		},
		{
			Name:  "Read dial sessions (has next page)",
			Input: common.ReadParams{ObjectName: "dialsession", Fields: connectors.Fields("dialsession_id", "start_when")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/1/dialsession"),
					mockcond.QueryParam("page_size", "100"),
					mockcond.QueryParam("page", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseDialSessions),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"dialsession_id": "49463",
						"start_when":     "2013-11-04 08:16:08",
					},
					Raw: map[string]any{
						"dialsession_id": "49463",
						"callerid":       nil,
						"start_when":     "2013-11-04 08:16:08",
						"call_count":     float64(0),
					},
				}},
				NextPage: testroutines.URLTestServer + "/rest/1/dialsession?page=2&page_size=100",
				Done:     false,
			},
		},
		{
			Name:  "Read members",
			Input: common.ReadParams{ObjectName: "members", Fields: connectors.Fields("user_id", "email_address")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/1/members"),
					mockcond.QueryParam("page_size", "100"),
					mockcond.QueryParam("page", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseMembers),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"user_id":       "1234567",
						"email_address": "saul@example.com",
					},
					Raw: map[string]any{
						"user_id":       "1234567",
						"username":      "demo_user",
						"first_name":    "Saul",
						"last_name":     "Goodman",
						"email_address": "saul@example.com",
					},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Read tags (has next page)",
			Input: common.ReadParams{ObjectName: "tags", Fields: connectors.Fields("id", "title")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/1/tags"),
					mockcond.QueryParam("page_size", "100"),
					mockcond.QueryParam("page", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseTags),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":    float64(1),
						"title": "Tag #1",
					},
					Raw: map[string]any{
						"id":    float64(1),
						"title": "Tag #1",
					},
				}, {
					Fields: map[string]any{
						"id":    float64(2),
						"title": "Tag #2",
					},
					Raw: map[string]any{
						"id":    float64(2),
						"title": "Tag #2",
					},
				}},
				NextPage: testroutines.URLTestServer + "/rest/1/tags?page=2&page_size=100",
				Done:     false,
			},
		},
		{
			Name: "Incremental read members returns empty for future Since",
			Input: common.ReadParams{
				ObjectName: "members",
				Fields:     connectors.Fields("user_id", "email_address"),
				Since:      time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/1/members"),
					mockcond.QueryParam("page_size", "100"),
					mockcond.QueryParam("page", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseMembers),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     0,
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Read voicemails",
			Input: common.ReadParams{ObjectName: "voicemails", Fields: connectors.Fields("recording_id", "name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/1/voicemails"),
					mockcond.QueryParam("page_size", "100"),
					mockcond.QueryParam("page", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseVoicemails),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"recording_id": "170999",
						"name":         "Basic Voicemail",
					},
					Raw: map[string]any{
						"recording_id": "170999",
						"name":         "Basic Voicemail",
						"playback_url": "http://sampledomain.com/pbx/dsrecording/AH-AEE-AAADGFF/170999/x.mp3",
						"created_when": "2013-11-07 10:16:56",
					},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Incremental read voicemails returns empty for future Since",
			Input: common.ReadParams{
				ObjectName: "voicemails",
				Fields:     connectors.Fields("recording_id", "name"),
				Since:      time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/1/voicemails"),
					mockcond.QueryParam("page_size", "100"),
					mockcond.QueryParam("page", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseVoicemails),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     0,
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Read folders (map response)",
			Input: common.ReadParams{ObjectName: "folders", Fields: connectors.Fields("folder_id", "folder_name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/rest/1/folders"),
				Then:  mockserver.Response(http.StatusOK, responseFolders),
			}.Server(),
			Comparator: comparatorSubsetReadOrderByFolderID,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"folder_id":   "11888",
						"folder_name": "Folder #1",
					},
					Raw: map[string]any{
						"folder_id":   "11888",
						"folder_name": "Folder #1",
					},
				}, {
					Fields: map[string]any{
						"folder_id":   "11999",
						"folder_name": "Folder #2",
					},
					Raw: map[string]any{
						"folder_id":   "11999",
						"folder_name": "Folder #2",
					},
				}},
				NextPage: "",
				Done:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func comparatorSubsetReadOrderByFolderID(serverURL string, actual, expected *common.ReadResult) bool {
	sort.Slice(actual.Data, func(i, j int) bool {
		ai, _ := actual.Data[i].Fields["folder_id"].(string)
		aj, _ := actual.Data[j].Fields["folder_id"].(string)
		return ai < aj
	})
	sort.Slice(expected.Data, func(i, j int) bool {
		ai, _ := expected.Data[i].Fields["folder_id"].(string)
		aj, _ := expected.Data[j].Fields["folder_id"].(string)
		return ai < aj
	})

	return testroutines.ComparatorSubsetRead(serverURL, actual, expected)
}

func constructTestConnector(serverURL string) (*Connector, error) {
	conn, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	conn.SetBaseURL(mockutils.ReplaceURLOrigin(conn.HTTPClient().Base, serverURL))

	return conn, nil
}
