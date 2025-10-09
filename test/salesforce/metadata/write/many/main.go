package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
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

	params := &common.UpsertMetadataParams{
		Fields: map[string][]common.FieldDefinition{
			"TestObject15__c": {
				{
					FieldName:   "Birthday__c",
					DisplayName: "Birthday",
					Description: "Story describing birthday",
					ValueType:   common.ValueTypeString,
					Required:    false,
					Unique:      false,
					Indexed:     false,
					StringOptions: &common.StringFieldOptions{
						Length: goutils.Pointer(30),
					},
				},
				{
					FieldName:   "Hobby__c",
					DisplayName: "Hobby",
					Description: "Your hobby description",
					ValueType:   common.ValueTypeString,
					Required:    false,
					Unique:      false,
					Indexed:     false,
					StringOptions: &common.StringFieldOptions{
						Length:          goutils.Pointer(444),
						NumDisplayLines: goutils.Pointer(39),
					},
				},
				{
					FieldName:   "Age__c",
					DisplayName: "Age",
					Description: "How many years you lived.",
					ValueType:   common.ValueTypeInt,
					Required:    true,
					Unique:      false,
					Indexed:     false,
					NumericOptions: &common.NumericFieldOptions{
						DefaultValue: goutils.Pointer(18.0),
						Precision:    goutils.Pointer(3),
						Scale:        goutils.Pointer(2),
					},
				},
				{
					FieldName:   "Interests__c",
					DisplayName: "Interests",
					Description: "Topics that are of interest.",
					ValueType:   common.ValueTypeMultiSelect,
					Required:    true,
					Unique:      false,
					Indexed:     false,
					StringOptions: &common.StringFieldOptions{
						Values:           []string{"art", "travel", "swimming"},
						ValuesRestricted: true,
						DefaultValue:     goutils.Pointer("art"),
					},
				},
				{
					FieldName:   "IsReady__c",
					DisplayName: "IsReady",
					Description: "Indicates the readiness for next steps.",
					ValueType:   common.ValueTypeBoolean,
					Required:    false,
					Unique:      false,
					Indexed:     false,
					StringOptions: &common.StringFieldOptions{
						DefaultValue: goutils.Pointer("false"),
					},
				},
				{
					FieldName:   "Connection__c",
					DisplayName: "Connection",
					Description: "Connection to other objects.",
					ValueType:   common.ValueTypeOther,
					Required:    false,
					Unique:      false,
					Indexed:     false,
					Association: &common.AssociationDefinition{
						AssociationType: "associatedAccount",
						TargetObject:    "Account",
						// TargetField: "Identifier",  makes an IndirectLookup field
						// (Salesforce account must have that in the first place)
						OnDelete:               "SetNull",
						ReverseLookupFieldName: "MyAccount",
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
