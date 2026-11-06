package reader

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testutils"
	"github.com/google/go-cmp/cmp"
)

func TestReadSparseLineItemFields(t *testing.T) {
	t.Parallel()

	response := testutils.DataFromFile(t, "checkout-sessions/sparse-line-item-metadata.json")
	server := mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If: mockcond.And{
			mockcond.Path("/v1/checkout/sessions"),
			mockcond.QueryParam("expand[]", "data.line_items"),
		},
		Then: mockserver.Response(http.StatusOK, response),
	}.Server()
	t.Cleanup(server.Close)

	strategy, err := constructTestStrategy(server)
	if err != nil {
		t.Fatal(err)
	}

	result, err := strategy.Read(context.Background(), common.ReadParams{
		ObjectName: "checkout/sessions",
		Fields: connectors.Fields(
			"$['line_items']['data'][*]['id']",
			"$['line_items']['data'][*]['currency']",
			"$['line_items']['data'][*]['price']['metadata']['department']",
		),
	})
	if err != nil {
		t.Fatal(err)
	}

	var original struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(response, &original); err != nil {
		t.Fatal(err)
	}

	expected := &common.ReadResult{
		Rows: 1,
		Done: true,
		Data: []common.ReadResultRow{{
			Id:  "cs_sparse_metadata",
			Raw: original.Data[0],
			Fields: map[string]any{
				"line_items": map[string]any{
					"data": []any{
						map[string]any{"id": "li_first", "currency": "usd"},
						map[string]any{
							"id": "li_second", "currency": "eur",
							"price": map[string]any{"metadata": map[string]any{"department": "research"}},
						},
						map[string]any{"id": "li_third", "currency": "cad"},
					},
				},
			},
		}},
	}
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("read result mismatch (-want +got):\n%s", diff)
	}
}
