package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/providers/salesforce"
	connTest "github.com/amp-labs/connectors/test/salesforce"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	conn := connTest.GetSalesforceConnector(ctx)

	ctx = common.WithAuthToken(ctx, connTest.GetSalesforceAccessToken())

	fmt.Print("====== Creating Field ======")

	performUpsert(conn, ctx, &common.UpsertMetadataParams{
		Fields: map[string][]common.FieldDefinition{
			"TestObject15__c": {{
				FieldName:   "Priority__c",
				DisplayName: "Priority",
				Description: "Description of the field priority",
				ValueType:   common.FieldTypeSingleSelect,
				Required:    true,
				Unique:      false,
				Indexed:     false,
				StringOptions: &common.StringFieldOptions{
					Values:           []string{"high", "low"},
					ValuesRestricted: true,
					DefaultValue:     goutils.Pointer("low"),
				},
			}},
		},
	})

	fmt.Print("====== Updating Field ======")

	performUpsert(conn, ctx, &common.UpsertMetadataParams{
		Fields: map[string][]common.FieldDefinition{
			"TestObject15__c": {{
				FieldName:   "Priority__c",
				DisplayName: "Priority",
				Description: "Very new description of the field priority",
				ValueType:   common.FieldTypeSingleSelect,
				Required:    true,
				Unique:      false,
				Indexed:     false,
				StringOptions: &common.StringFieldOptions{
					Values:           []string{"high", "medium", "low"},
					ValuesRestricted: true,
					DefaultValue:     goutils.Pointer("low"),
				},
			}},
		},
	})
}

func performUpsert(conn *salesforce.Connector, ctx context.Context, params *common.UpsertMetadataParams) {
	res, err := conn.UpsertMetadata(ctx, params)
	if err != nil {
		utils.Fail("upsert FAILED", "error", err)
	}

	for objectName, fields := range res.Fields {
		for fieldName, field := range fields {
			slog.Info("SUCCESS", "object", objectName, "field", fieldName, "action", field.Action)
		}
	}
}
