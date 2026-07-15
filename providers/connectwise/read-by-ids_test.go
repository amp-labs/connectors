package connectwise

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestGetRecordsByIds(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseContacts := testutils.DataFromFile(t, "read/contacts-batch-by-ids.json")

	tests := []testconn.TestCaseGetRecordsByIds{
		{
			Name:         "Empty record identifiers",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Read contacts by identifiers",
			Input: testconn.ReadByIdsParams{
				ObjectName: "contacts",
				RecordIds:  []string{"57920", "57921", "57922"},
				Fields:     []string{"firstName", "customField53", "Job Level"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v4_6_release/apis/3.0/company/contacts"),
					mockcond.Permute(
						queryParam("conditions", "id in (%v)"),
						"57920", "57921", "57922",
					),
				},
				Then: mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Comparator: testconn.ComparatorSortedSubsetReadByIds,
			Expected: []common.ReadResultRow{{
				Id: "57920",
				Fields: map[string]any{
					"firstname":     "Maxime Schaefer [1]",
					"customfield53": "Software Developer",
					"job level":     "Software Developer",
				},
				Raw: map[string]any{"lastName": "Wayne Blanda"},
			}, {
				Id: "57921",
				Fields: map[string]any{
					"firstname":     "Roderick Rippin [2]",
					"customfield53": "Sales Representative",
					"job level":     "Sales Representative",
				},
				Raw: map[string]any{"lastName": "Lemuel Hackett"},
			}, {
				Id: "57922",
				Fields: map[string]any{
					"firstname":     "Randi Haag [3]",
					"customfield53": "Manager",
					"job level":     "Manager",
				},
				Raw: map[string]any{"lastName": "Ferne Bradtke"},
			}},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableBatchReader, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}

func queryParam(key, valueTemplate string) func(fields []string) mockcond.Condition {
	return func(identifiers []string) mockcond.Condition {
		param := fmt.Sprintf(valueTemplate, strings.Join(identifiers, ","))

		return mockcond.QueryParam(key, param)
	}
}
