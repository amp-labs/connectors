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

func TestWrite(t *testing.T) {
	t.Parallel()

	responseArticlesCreate := testutils.DataFromFile(t, "write-articles-create-response.json")
	responseArticlesUpdate := testutils.DataFromFile(t, "write-articles-update-response.json")

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
