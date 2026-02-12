package phoneburner

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

func TestWrite(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	respContact := testutils.DataFromFile(t, "write/contact.json")
	respFolder := testutils.DataFromFile(t, "write/folder.json")
	respMember := testutils.DataFromFile(t, "write/member.json")
	respDialsession := testutils.DataFromFile(t, "write/dialsession.json")
	respUnauthorized := testutils.DataFromFile(t, "read/error-unauthorized.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write data must be included",
			Input:        common.WriteParams{ObjectName: "contacts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:         "Unknown objects are not supported",
			Input:        common.WriteParams{ObjectName: "tiger", RecordData: map[string]any{"x": "y"}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Create contact (form encoded)",
			Input: common.WriteParams{ObjectName: "contacts", RecordData: map[string]any{"first_name": "Johnny", "last_name": "Demo"}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/rest/1/contacts"),
					mockcond.HeaderContentURLFormEncoded(),
					// url.Values.Encode sorts keys: first_name=...&last_name=...
					mockcond.Body("first_name=Johnny&last_name=Demo"),
				},
				Then: mockserver.Response(http.StatusOK, respContact),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "30919347",
				Data: map[string]any{
					"first_name": "Johnny",
					"last_name":  "Demo",
				},
			},
		},
		{
			Name:  "Update contact (form encoded)",
			Input: common.WriteParams{ObjectName: "contacts", RecordId: "30919347", RecordData: map[string]any{"first_name": "Johnny"}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/rest/1/contacts/30919347"),
					mockcond.HeaderContentURLFormEncoded(),
					mockcond.Body("first_name=Johnny"),
				},
				Then: mockserver.Response(http.StatusOK, respContact),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "30919347",
				Data: map[string]any{
					"first_name": "Johnny",
				},
			},
		},
		{
			Name:  "Create folder (JSON)",
			Input: common.WriteParams{ObjectName: "folders", RecordData: map[string]any{"name": "Folder #1", "description": "My Description", "parent_id": 0}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/rest/1/folders"),
					mockcond.Body(`{"description":"My Description","name":"Folder #1","parent_id":0}`),
				},
				Then: mockserver.Response(http.StatusOK, respFolder),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "11888",
				Data: map[string]any{
					"folder_id":   "11888",
					"folder_name": "Folder #1",
				},
			},
		},
		{
			Name:  "Update member (form encoded)",
			Input: common.WriteParams{ObjectName: "members", RecordId: "25381104", RecordData: map[string]any{"first_name": "johnny5"}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPUT(),
					mockcond.Path("/rest/1/members/25381104"),
					mockcond.HeaderContentURLFormEncoded(),
					mockcond.Body("first_name=johnny5"),
				},
				Then: mockserver.Response(http.StatusOK, respMember),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "25381104",
				Data: map[string]any{
					"user_id":       "25381104",
					"first_name":    "johnny5",
					"username":      "gumby",
					"date_added":    "2023-09-22 14:51:38",
					"last_name":     "",
					"email_address": "gumby@example.com",
				},
			},
		},
		{
			Name:  "Create dialsession (JSON)",
			Input: common.WriteParams{ObjectName: "dialsession", RecordData: map[string]any{"preview_mode": 1}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/rest/1/dialsession"),
					mockcond.Body(`{"preview_mode":1}`),
				},
				Then: mockserver.Response(http.StatusOK, respDialsession),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success: true,
				Data: map[string]any{
					"redirect_url": "https://www.phoneburner.com/index/sso_key_login?single_sign_on_secret=abc123&redirect=%2Fphoneburner%2Fapi_begin_dialsession",
				},
			},
		},
		{
			Name:  "Provider envelope error is mapped for write (200 with http_status=401)",
			Input: common.WriteParams{ObjectName: "contacts", RecordData: map[string]any{"first_name": "Johnny"}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/rest/1/contacts"),
					mockcond.HeaderContentURLFormEncoded(),
					mockcond.Body("first_name=Johnny"),
				},
				Then: mockserver.Response(http.StatusOK, respUnauthorized),
			}.Server(),
			ExpectedErrs: []error{common.ErrAccessToken},
		},
		{
			Name:   "Unsupported write object returns not supported",
			Input:  common.WriteParams{ObjectName: "voicemails", RecordData: map[string]any{"name": "x"}},
			Server: mockserver.Dummy(),
			ExpectedErrs: []error{
				common.ErrOperationNotSupportedForObject,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
