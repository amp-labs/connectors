package slack

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestGetRecordsByIds(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseConversation1 := testutils.DataFromFile(t, "read/conversation-1.json")
	responseConversation2 := testutils.DataFromFile(t, "read/conversation-2.json")

	tests := []testconn.TestCaseGetRecordsByIds{
		{
			Name:         "Empty record identifiers",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			// Messages have no singular lookup endpoint; the record lives in the webhook
			// payload, so the caller is told to read it from there instead of fetching.
			Name: "Messages report records as inline only",
			Input: testconn.ReadByIdsParams{
				ObjectName: "messages",
				RecordIds:  []string{"C0B9N5F9ULE:1781123456.000200"},
				Fields:     []string{"text"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrRecordsOnlyInline},
		},
		{
			// Objects that are readable but have no ".info" endpoint keep the plain
			// not-supported sentinel: their records are nowhere to be found.
			Name: "Object without a singular endpoint is not fetchable",
			Input: testconn.ReadByIdsParams{
				ObjectName: "bookmarks",
				RecordIds:  []string{"Bk0B9V3RLZ4M"},
				Fields:     []string{"title"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrGetRecordNotSupportedForObject},
		},
		{
			Name: "Read conversations by identifiers",
			Input: testconn.ReadByIdsParams{
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
			Comparator: testconn.ComparatorSortedSubsetReadByIds,
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

			tt.Run(t, func() (testconn.TestableBatchReader, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}
