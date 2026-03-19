package main

import (
	"context"
	"os/signal"
	"reflect"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/microsoft"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

type payload struct {
	Subject      string      `json:"subject,omitempty"`
	Body         body        `json:"body,omitempty"`
	From         *recipient  `json:"from,omitempty"`
	ToRecipients []recipient `json:"toRecipients,omitempty"`
}

type body struct {
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
}

type recipient struct {
	EmailAddress address `json:"emailAddress"`
}

type address struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

const TextContentType = "text"

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetMicrosoftGraphConnector(ctx)

	subject := gofakeit.Name()
	bodyData := gofakeit.Name()
	from := gofakeit.Username()
	to := gofakeit.Username()
	updatedSubject := gofakeit.Name()
	updatedBodyData := gofakeit.Name()

	// https://learn.microsoft.com/en-us/graph/api/resources/message?view=graph-rest-1.0
	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"me/messages",
		payload{
			Subject: subject,
			Body: body{
				Content:     bodyData,
				ContentType: TextContentType,
			},
			From: &recipient{
				EmailAddress: address{
					Address: from + "@test.com",
					Name:    from,
				},
			},
			ToRecipients: []recipient{{
				EmailAddress: address{
					Address: to + "@test.com",
					Name:    to,
				},
			}},
		},
		payload{
			Subject: updatedSubject,
			Body: body{
				Content:     updatedBodyData,
				ContentType: TextContentType,
			},
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("id", "subject", "body", "from", "toRecipients"),
			SearchBy: testscenario.Property{
				Key:   "subject",
				Value: subject,
			},
			RecordIdentifierKey: "id",
			ValidateUpdatedFields: func(record map[string]any) {
				if !reflect.DeepEqual(record["subject"], updatedSubject) {
					utils.Fail("subject mismatch")
				}
				if !reflect.DeepEqual(record["body"], map[string]any{
					"content":     updatedBodyData,
					"contentType": TextContentType,
				}) {
					utils.Fail("body mismatch")
				}
				if !reflect.DeepEqual(record["from"], map[string]any{
					"emailAddress": map[string]any{
						"address": from + "@test.com",
						"name":    from,
					},
				}) {
					utils.Fail("from mismatch")
				}
				if !reflect.DeepEqual(record["torecipients"], []any{
					map[string]any{
						"emailAddress": map[string]any{
							"address": to + "@test.com",
							"name":    to,
						},
					},
				}) {
					utils.Fail("torecipients mismatch")
				}
			},
		},
	)
}
