package devrev

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

func TestWriteResponseKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		objectName string
		want       string
	}{
		{"accounts", "account"},
		{"rev-users", "rev_user"},
		{"dev-orgs.auth-connections", "auth_connection"},
		{"groups.members", "member"},
		{"schemas.subtypes", "subtype"},
	}
	for _, tt := range tests {
		t.Run(tt.objectName, func(t *testing.T) {
			got := writeResponseKey(tt.objectName)
			if got != tt.want {
				t.Errorf("writeResponseKey(%q) = %q, want %q", tt.objectName, got, tt.want)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	t.Parallel()

	responseArticlesCreate := testutils.DataFromFile(t, "write-articles-create-response.json")
	responseArticlesUpdate := testutils.DataFromFile(t, "write-articles-update-response.json")
	responseRevUserCreate := testutils.DataFromFile(t, "write-rev-user-response.json")
	responseAuthTokensCreate := testutils.DataFromFile(t, "write-auth-tokens-create-response.json")

	tests := []testroutines.Write{
		{
			Name: "Create article successfully",
			Input: common.WriteParams{
				ObjectName: "articles",
				RecordData: map[string]any{
					"title": "api test",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Path("/articles.create"),
				},
				Then: mockserver.Response(http.StatusCreated, responseArticlesCreate),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "don:core:devrev:article/1",
				Data: map[string]any{
					"id":    "don:core:devrev:article/1",
					"title": "api test",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update article successfully",
			Input: common.WriteParams{
				ObjectName: "articles",
				RecordId:   "don:core:devrev:article/1",
				RecordData: map[string]any{
					"title": "updated title",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Path("/articles.update"),
					mockcond.Body(`{"id":"don:core:devrev:article/1","title":"updated title"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseArticlesUpdate),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "don:core:devrev:article/1",
				Data: map[string]any{
					"id":    "don:core:devrev:article/1",
					"title": "updated title",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Create rev-user successfully (hyphenated object)",
			Input: common.WriteParams{
				ObjectName: "rev-users",
				RecordData: map[string]any{
					"display_name": "alex.customer",
					"email":        "alex.customer@example.com",
					"full_name":    "Alex Johnson",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Path("/rev-users.create"),
				},
				Then: mockserver.Response(http.StatusCreated, responseRevUserCreate),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "1b5d9e8e-6e12-4a0a-bf67-2a8e34c8e2aa",
				Data: map[string]any{
					"id":           "1b5d9e8e-6e12-4a0a-bf67-2a8e34c8e2aa",
					"display_id":   "REVU-4481",
					"display_name": "alex.customer",
					"email":        "alex.customer@example.com",
					"full_name":    "Alex Johnson",
					"state":        "active",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Create auth-token successfully (flat response, no wrapper)",
			Input: common.WriteParams{
				ObjectName: "auth-tokens",
				RecordData: map[string]any{
					"grant_type":           "urn:devrev:params:oauth:grant-type:token-issue",
					"requested_token_type": "urn:devrev:params:oauth:token-type:aat",
					"client_id":            "crawler-service",
					"aud":                  []any{"https://api.devrev.ai"},
					"scope":                "webcrawler.read webcrawler.write",
					"expires_in":           30,
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Path("/auth-tokens.create"),
				},
				Then: mockserver.Response(http.StatusCreated, responseAuthTokensCreate),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "",
				Data: map[string]any{
					"access_token":  "dvt_aat_abc123xyz",
					"client_id":     "crawler-service",
					"expires_in":    float64(3600),
					"refresh_token": "dvt_rt_xyz789",
					"scope":         "webcrawler.read webcrawler.write",
					"token_type":    "bearer",
				},
			},
			ExpectedErrs: nil,
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
