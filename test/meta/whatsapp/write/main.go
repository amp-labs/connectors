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

type textMessagePayload struct {
	MessagingProduct string         `json:"messaging_product"`
	RecipientType    string         `json:"recipient_type"`
	To               string         `json:"to"`
	Type             string         `json:"type"`
	Text             map[string]any `json:"text"`
}

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()
	conn := connTest.GetWhatsAppConnector(ctx)
	slog.Info("Running message template create")
	runMessageTemplateCreate(ctx, conn)
	slog.Info("Running text message create")
	runTextMessageCreate(ctx, conn)
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

func runTextMessageCreate(ctx context.Context, conn connectors.WriteConnector) {
	to := os.Getenv("WHATSAPP_TO")
	if to == "" {
		slog.Info("Skipping text message create (set WHATSAPP_TO to run)")
		return
	}

	createRes, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "messages",
		RecordData: textMessagePayload{
			MessagingProduct: "whatsapp",
			RecipientType:    "individual",
			To:               to,
			Type:             "text",
			Text: map[string]any{
				"preview_url": true,
				"body": "As requested, here's the link to our latest product: " +
					"https://www.meta.com/quest/quest-3/",
			},
		},
	})
	if err != nil {
		utils.Fail("error creating text message", "error", err)
	}
	if !createRes.Success {
		utils.Fail("failed to create text message", "response", createRes)
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
