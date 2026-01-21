package dropboxsign

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

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	createTeamResponse := testutils.DataFromFile(t, "create-team.json")
	updateTeamResponse := testutils.DataFromFile(t, "update-team.json")

	tests := []testroutines.Write{
		{
			Name:         "Object Name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: "team"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name: "Successfully creation of a team",
			Input: common.WriteParams{ObjectName: "team", RecordData: map[string]any{
				"name": "New Team Name",
			}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, createTeamResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "New Team Name",
				Data: map[string]any{
					"name": "New Team Name",
					"accounts": []any{
						map[string]any{
							"account_id":    "5008b25c7f67153e57d5a357b1687968068fb465",
							"email_address": "me@dropboxsign.com",
							"is_locked":     false,
							"is_paid_hs":    true,
							"is_paid_hf":    false,
							"quotas": map[string]any{
								"templates_left":              nil,
								"documents_left":              nil,
								"api_signature_requests_left": float64(1250),
							},
							"role_code": "a",
						},
					},
					"invited_accounts": []any{},
					"invited_emails":   []any{},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully update of a team",
			Input: common.WriteParams{
				ObjectName: "team",
				RecordId:   "New Team Name",
				RecordData: map[string]any{
					"name": "New Team Name",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPUT(),
				Then:  mockserver.Response(http.StatusOK, updateTeamResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "New Team Name",
				Data: map[string]any{
					"name": "New Team Name",
					"accounts": []any{
						map[string]any{
							"account_id":    "5008b25c7f67153e57d5a357b1687968068fb465",
							"email_address": "me@dropboxsign.com",
							"is_locked":     false,
							"is_paid_hs":    true,
							"is_paid_hf":    false,
							"quotas": map[string]any{
								"templates_left":              nil,
								"documents_left":              nil,
								"api_signature_requests_left": float64(1250),
							},
							"role_code": "a",
						},
					},
					"invited_accounts": []any{
						map[string]any{
							"account_id":    "8e239b5a50eac117fdd9a0e2359620aa57cb2463",
							"email_address": "george@hellofax.com",
							"is_locked":     false,
							"is_paid_hs":    false,
							"is_paid_hf":    false,
							"quotas": map[string]any{
								"templates_left":              float64(0),
								"documents_left":              float64(3),
								"api_signature_requests_left": float64(0),
							},
						},
					},
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
