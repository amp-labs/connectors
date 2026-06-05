package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	connTest "github.com/amp-labs/connectors/test/meta"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/brianvoe/gofakeit/v6"
)

type messageTemplatePayload struct {
	Name       string           `json:"name"`
	Language   string           `json:"language"`
	Category   string           `json:"category"`
	Components []map[string]any `json:"components"`
}

type sendMessagePayload struct {
	MessagingProduct string         `json:"messaging_product"`
	To               string         `json:"to"`
	Type             string         `json:"type"`
	Template         map[string]any `json:"template,omitempty"`
}

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()
	conn := connTest.GetWhatsAppConnector(ctx)
	slog.Info("Running message template create")
	runMessageTemplateCreate(ctx, conn)
	slog.Info("Running template message send")
	runTextMessageCreate(ctx, conn, connTest.GetWhatsAppTo())
}

func runMessageTemplateCreate(ctx context.Context, conn connectors.WriteConnector) {
	templateName := "seasonal_promotion_" + strings.ToLower(gofakeit.LetterN(8))

	createRes, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "message_templates",
		RecordData: messageTemplatePayload{
			Name:       templateName,
			Language:   "en_US",
			Category:   "MARKETING",
			Components: seasonalPromotionComponents(),
		},
	})
	if err != nil {
		utils.Fail("error creating message template", "error", err)
	}
	if !createRes.Success {
		utils.Fail("failed to create message template", "response", createRes)
	}
	utils.DumpJSON(createRes, os.Stdout)
}

func runTextMessageCreate(ctx context.Context, conn connectors.WriteConnector, to string) {
	if to == "" {
		slog.Info("Skipping template message send (set metadata.whatsappTo in meta-creds.json or META_WHATSAPP_TO to run)")
		return
	}

	createRes, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "messages",
		RecordData: sendMessagePayload{
			MessagingProduct: "whatsapp",
			To:               to,
			Type:             "template",
			Template: map[string]any{
				"name": "jaspers_market_order_confirmation_v1",
				"language": map[string]any{
					"code": "en_US",
				},
				"components": []map[string]any{
					{
						"type": "body",
						"parameters": []map[string]any{
							{"type": "text", "text": "John Doe"},
							{"type": "text", "text": "123456"},
							{"type": "text", "text": "Jun 4, 2026"},
						},
					},
				},
			},
		},
	})
	if err != nil {
		utils.Fail("error sending template message", "error", err)
	}
	if !createRes.Success {
		utils.Fail("failed to send template message", "response", createRes)
	}
	utils.DumpJSON(createRes, os.Stdout)
}

func seasonalPromotionComponents() []map[string]any {
	return []map[string]any{
		{
			"type":   "HEADER",
			"format": "TEXT",
			"text":   "Our {{1}} is on!",
			"example": map[string]any{
				"header_text": []string{"Summer Sale"},
			},
		},
		{
			"type": "BODY",
			"text": "Shop now through {{1}} and use code {{2}} to get {{3}} off.",
			"example": map[string]any{
				"body_text": [][]string{
					{"25OFF", "25%", "50%"},
				},
			},
		},
		{
			"type": "FOOTER",
			"text": "Use the buttons below to manage your subscriptions",
		},
		{
			"type": "BUTTONS",
			"buttons": []map[string]any{
				{
					"type": "QUICK_REPLY",
					"text": "Unsubscribe from Promos",
				},
				{
					"type": "QUICK_REPLY",
					"text": "Unsubscribe from All",
				},
			},
		},
	}
}
