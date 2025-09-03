package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/hubspot"
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

	conn := connTest.GetHubspotConnector(ctx)

	params := &common.UpsertMetadataParams{
		Fields: map[string][]common.FieldDefinition{
			"Contact": {
				{
					FieldName:   "hobby__c",
					DisplayName: "Hobby",
					Description: "Your hobby description",
					ValueType:   common.ValueTypeString,
					Unique:      true,
					UniqueProperties: common.UniqueProperties{
						HubspotGroupName: "contactinformation",
					},
				},
				{
					FieldName:   "age__c",
					DisplayName: "Age",
					Description: "How many years you lived.",
					ValueType:   common.ValueTypeInt,
					Unique:      false,
					UniqueProperties: common.UniqueProperties{
						HubspotGroupName: "contactinformation",
					},
				},
				{
					FieldName:   "interests__c",
					DisplayName: "Interests",
					Description: "Topics that are of interest.",
					ValueType:   common.ValueTypeMultiSelect,
					Unique:      false,
					StringOptions: &common.StringFieldOptions{
						Values: []string{"art", "travel", "swimming"},
					},
					UniqueProperties: common.UniqueProperties{
						HubspotGroupName: "contactinformation",
					},
				},
				{
					FieldName:   "isready__c",
					DisplayName: "IsReady",
					Description: "Indicates the readiness for next steps.",
					ValueType:   common.ValueTypeBoolean,
					Unique:      false,
					UniqueProperties: common.UniqueProperties{
						HubspotGroupName: "contactinformation",
					},
				},
			},
		},
	}

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
