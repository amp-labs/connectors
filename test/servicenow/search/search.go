package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/servicenow"
)

// Mid-call lookup demo: find an Incident by ticket number, return a small set
// of fields, validate the caller against the record on the client side.
func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	conn := servicenow.GetServiceNowConnector(ctx)

	res, err := conn.Search(ctx, &common.SearchParams{
		ObjectName: "case",
		Fields:     datautils.NewStringSet("number", "state", "short_description", "priority"),
		Filter: common.SearchFilter{
			FieldFilters: []common.FieldFilter{
				{
					FieldName: "number",
					Operator:  common.FilterOperatorEQ,
					Value:     "CS0001001",
				},
			},
		},
		Limit: 1,
	})
	if err != nil {
		return err
	}

	out, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}

	_, _ = os.Stdout.Write(out)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
