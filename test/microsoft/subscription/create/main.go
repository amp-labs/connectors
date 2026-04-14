package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/microsoft"
	connTest "github.com/amp-labs/connectors/test/microsoft"
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

	conn := connTest.GetMicrosoftGraphConnector(ctx)

	subject := gofakeit.Name()
	bodyData := gofakeit.Name()
	from := gofakeit.Username()
	to := gofakeit.Username()

	// https://7565-46-150-81-5.ngrok-free.app

	testscenario.ValidateSubscribeReceive(ctx, conn,
		testscenario.SubscribeEventsSuite{
			SubscribeParamBuilder: func(webhookURL string) *common.SubscribeParams {
				return &common.SubscribeParams{
					Request: microsoft.SubscribeRequest{
						WebhookURL: webhookURL,
					},
					RegistrationResult: nil,
					SubscriptionEvents: map[common.ObjectName]common.ObjectEvents{
						"me/messages": {
							Events: []common.SubscriptionEventType{
								common.SubscriptionEventTypeCreate,
							},
							WatchFields: []string{
								"subject", "from", "toRecipients",
							},
							WatchFieldsAll:    true,
							PassThroughEvents: nil,
						},
					},
				}
			},
			WebhookRouter: testscenario.WebhookRouter{
				Routes: []testscenario.Route{
					{
						// This route is executed when Microsoft is verifying that webhook is rechable.
						Left: func(request *http.Request) bool {
							url, err := urlbuilder.FromRawURL(request.URL)
							if err != nil {
								return false
							}

							return url.HasQueryParam("validationToken")
						},
						// https://learn.microsoft.com/en-us/graph/change-notifications-delivery-webhooks?tabs=http#notificationurl-validation
						Right: func(writer http.ResponseWriter, request *http.Request) {
							url, err := urlbuilder.FromRawURL(request.URL)
							if err != nil {
								writer.WriteHeader(http.StatusInternalServerError)
								return
							}

							validationToken, ok := url.GetFirstQueryParam("validationToken")
							if !ok {
								writer.WriteHeader(http.StatusInternalServerError)
								return
							}

							writer.WriteHeader(http.StatusOK)
							writer.Header().Set("Content-Type", "text/plain")
							_, _ = writer.Write([]byte(validationToken))
						},
					},
					// Default handling.
					// https://learn.microsoft.com/en-us/graph/change-notifications-with-resource-data?tabs=csharp#decrypting-resource-data-from-change-notifications
				},
			},
			Triggers: []testscenario.SubscriptionTrigger{{
				ObjectName: "me/messages",
				Payload: payload{
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
				}},
			},
			VerificationParams: nil,
		},
	)
}
