package slack

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestGetRecordsByIds(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseConversation1 := testutils.DataFromFile(t, "read/conversation-1.json")
	responseConversation2 := testutils.DataFromFile(t, "read/conversation-2.json")

	tests := []testroutines.TestCaseGetRecordsByIds{
		{
			Name:         "Empty record identifiers",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Read conversations by identifiers",
			Input: testroutines.ReadByIdsParams{
				ObjectName: "conversations",
				RecordIds:  []string{"C0BA24516MS", "C0B9V3RLZ4M"},
				Fields:     []string{"name"},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{{
					If: mockcond.And{
						mockcond.MethodGET(),
						mockcond.Path("/api/conversations.info"),
						mockcond.QueryParam("channel", "C0BA24516MS"),
					},
					Then: mockserver.Response(http.StatusOK, responseConversation1),
				}, {
					If: mockcond.And{
						mockcond.MethodGET(),
						mockcond.Path("/api/conversations.info"),
						mockcond.QueryParam("channel", "C0B9V3RLZ4M"),
					},
					Then: mockserver.Response(http.StatusOK, responseConversation2),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSortedSubsetReadByIds,
			Expected: []common.ReadResultRow{{
				Id:     "C0B9V3RLZ4M",
				Fields: map[string]any{"name": "holidays"},
				Raw:    map[string]any{"updated": float64(1781194917474)},
			}, {
				Id:     "C0BA24516MS",
				Fields: map[string]any{"name": "comedy"},
				Raw:    map[string]any{"updated": float64(1781194918084)},
			}},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testroutines.TestableBatchReader, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}

func TestSingleRecordMappings(t *testing.T) {
	result := testutils.NewCompareResult()

	first := datautils.FromMap(readSingleRecordResourceNameToQueryParam).KeySet()
	second := datautils.FromMap(readSingleRecordResourceNameToResponseField).KeySet()

	if !first.Equals(second) {
		result.AddDiff("key sets are different")
	}

	result.Validate(t, "Keys are shared between "+
		"readSingleRecordResourceNameToQueryParam and readSingleRecordResourceNameToResponseField")
}
