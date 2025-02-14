package front

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

func TestWrite(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	account := testutils.DataFromFile(t, "create-account.json")
	patchaccount := testutils.DataFromFile(t, "patch-account.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "contacts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:         "Unsupported object name",
			Input:        common.WriteParams{ObjectName: "butterflies", RecordData: "dummy"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrObjectNotSupported},
		},
		{
			Name:  "Creation of an account",
			Input: common.WriteParams{ObjectName: "accounts", RecordData: "realdata"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.PathSuffix("/accounts"),
					mockcond.MethodPOST(),
				},
				Then: mockserver.Response(http.StatusOK, account),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "acc_fkb9lo",
				Errors:   nil,
				Data: map[string]any{
					"created_at":  1739533550.22,
					"description": "A test user account creation",
					"id":          "acc_fkb9lo",
					"name":        "Test Users",
					"updated_at":  1739533550.22,
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update of an account via PATCH",
			Input: common.WriteParams{
				ObjectName: "accounts",
				RecordId:   "acc_fkb9ek",
				RecordData: "somenewdata",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.PathSuffix("/accounts/acc_fkb9ek"),
					mockcond.MethodPATCH(),
				},
				Then: mockserver.Response(http.StatusOK, patchaccount),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "acc_fkb9ek",
				Data: map[string]any{
					"created_at":  1739532196.973,
					"description": "A test user account creation",
					"id":          "acc_fkb9ek",
					"name":        "Update Test Users Account",
					"updated_at":  1739533386.028,
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
