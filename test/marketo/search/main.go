package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	mk "github.com/amp-labs/connectors/test/marketo"
)

func main() {
	ctx := context.Background()

	err := testSearchEmailTemplates(ctx)
	if err != nil {
		slog.Error(err.Error())
	}

	err = testSearchFolders(ctx)
	if err != nil {
		slog.Error(err.Error())
	}

	err = testSearchLeads(ctx)
	if err != nil {
		slog.Error(err.Error())
	}
}

func testSearchEmailTemplates(ctx context.Context) error {
	conn := mk.GetMarketoConnector(ctx)

	params := common.SearchParams{
		ObjectName: "emailTemplates",
		Fields:     connectors.Fields("description", "id", "status"),
		Filter: common.SearchFilter{
			FieldFilters: []common.FieldFilter{
				{
					FieldName: "status",
					Operator:  common.FilterOperatorEQ,
					Value:     "draft",
				},
			},
		},
	}

	res, err := conn.Search(ctx, &params)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testSearchFolders(ctx context.Context) error {
	conn := mk.GetMarketoConnector(ctx)

	params := common.SearchParams{
		ObjectName: "folders",
		Fields:     connectors.Fields("isArchive", "id", "name"),
		Filter: common.SearchFilter{
			FieldFilters: []common.FieldFilter{
				{
					FieldName: "workSpace",
					Operator:  common.FilterOperatorEQ,
					Value:     "LiftAi",
				},
			},
		},
		NextPage: "https://388-inb-483.mktorest.com/rest/asset/v1/folders.json?maxReturn=2\u0026offset=2\u0026workSpace=LiftAi",
	}

	res, err := conn.Search(ctx, &params)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testSearchCampaigns(ctx context.Context) error {
	conn := mk.GetMarketoConnectorLeads(ctx)

	params := common.ReadParams{
		ObjectName: "campaigns",
		Fields:     connectors.Fields("createdAt", "id", "name"),
		Since:      time.Now().Add(-1800 * time.Hour),
	}

	res, err := conn.Read(ctx, params)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}

func testSearchLeads(ctx context.Context) error {
	conn := mk.GetMarketoConnectorLeads(ctx)

	params := common.SearchParams{
		ObjectName: "leads",
		Fields:     connectors.Fields("id", "email", "createdAt", "firstName"),
		Filter: common.SearchFilter{
			FieldFilters: []common.FieldFilter{
				{
					FieldName: "filterType",
					Operator:  common.FilterOperatorEQ,
					Value:     "email",
				},
				{
					FieldName: "filterValues",
					Operator:  common.FilterOperatorEQ,
					Value:     "test321@gmail.com",
				},
			},
		},
	}

	res, err := conn.Search(ctx, &params)
	if err != nil {
		return err
	}

	// Print the results
	jsonStr, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %w", err)
	}

	_, _ = os.Stdout.Write(jsonStr)
	_, _ = os.Stdout.WriteString("\n")

	return nil
}
