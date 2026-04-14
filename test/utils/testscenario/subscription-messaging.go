package testscenario

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/utils"
)

type ConnectorWebhookSubscriber interface {
	components.SubscriptionCreator
	components.WebhookMessageVerifier
	connectors.WriteConnector
}

type SubscribeParamBuilder func(webhookURL string) *common.SubscribeParams

type SubscribeEventsSuite struct {
	SubscribeParamBuilder SubscribeParamBuilder
	WebhookRouter         WebhookRouter
	Triggers              []SubscriptionTrigger
	VerificationParams    *common.VerificationParams
}

type WebhookRouter struct {
	Routes []Route
}

type Route datautils.Pair[RoutingCondition, http.HandlerFunc]

type RoutingCondition func(request *http.Request) bool

type SubscriptionTrigger struct {
	ObjectName string
	Payload    any
}

// ValidateSubscribeReceive is a comprehensive test scenario utilizing subscription connector operations.
//
// Flow:
// 1. Starts local server
// 2. Asks user for public URL (ngrok)
// 3. Creates subscription
// 4. Optionally triggers events (Write)
// 5. Waits for webhook(s)
// 6. Exits cleanly
func ValidateSubscribeReceive(ctx context.Context,
	conn ConnectorWebhookSubscriber,
	suite SubscribeEventsSuite,
) {
	fmt.Println("> TEST Subscribe/Write/Recieve")

	fmt.Println("Start webhook handler")
	messageChannel := make(chan webhookMessageResult)
	webhookURL, shutdown := startWebhookHandler(ctx, conn, suite.WebhookRouter, suite.VerificationParams, messageChannel)
	defer shutdown()

	fmt.Printf("Local webhook server started at: %s\n", webhookURL)
	fmt.Println("Please provide the public URL (e.g., from ngrok) that tunnels to this local server.")
	fmt.Print("Public Webhook URL: ")

	var publicURL string
	if _, err := fmt.Scanln(&publicURL); err != nil {
		failOnError(fmt.Errorf("failed to read public URL: %w", err))
	}

	fmt.Println("============== Invoking connector.Subscribe() ==================")
	params := *suite.SubscribeParamBuilder(publicURL)
	subscriptionResult, err := conn.Subscribe(ctx, params)
	failOnError(err)
	utils.DumpJSON(subscriptionResult.Result, os.Stdout)

	fmt.Println("============== Invoking connector.Write() ==================")
	for _, trigger := range suite.Triggers {
		fmt.Printf("Creating object %v\n", trigger.ObjectName)
		createResult, err := createObject[any](ctx, conn, trigger.ObjectName, &trigger.Payload)
		failOnError(err)
		utils.DumpJSON(createResult, os.Stdout)
	}

	// Waiting for the events to arrive. Then report on them and exit.
	expectedNumEvents := len(suite.Triggers)
	receivedNumEvents := 0
	fmt.Printf("============== Waiting for %d webhook messages ==================\n", expectedNumEvents)

	for receivedNumEvents < expectedNumEvents {
		select {
		case message := <-messageChannel:
			receivedNumEvents++
			fmt.Printf("[%d/%d] Received webhook message:\n", receivedNumEvents, expectedNumEvents)
			if message.Error == "" {
				utils.DumpJSON(message.Body, os.Stdout)
			} else {
				utils.DumpJSON(message.Error, os.Stdout)
			}

		case <-ctx.Done():
			fmt.Println("Context cancelled, stopping...")
			return
		}
	}

	fmt.Println("============== Done ==================")
}

func startWebhookHandler(
	ctx context.Context, conn ConnectorWebhookSubscriber,
	router WebhookRouter,
	verificationParams *common.VerificationParams,
	messageChannel chan webhookMessageResult,
) (string, func()) {
	// Main server loop.
	webhookHandler := createWebhookHandler(ctx, conn, router, verificationParams, messageChannel)

	mux := http.NewServeMux()
	mux.HandleFunc("/", webhookHandler)

	// Construct and start server.
	const port = "4550"
	fmt.Printf("Starting webhook server on :%v\n", port)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
		}
	}()

	shutdown := func() {
		_ = server.Shutdown(context.Background())
	}

	return "http://localhost:" + port, shutdown
}
