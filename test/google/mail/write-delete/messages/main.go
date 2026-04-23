package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/google"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

type Payload struct {
	Message payloadMessage `json:"message"`
}

type payloadMessage struct {
	Raw string `json:"raw"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetGoogleMailConnector(ctx)

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"drafts",
		Payload{
			Message: payloadMessage{
				Raw: "Q29udGVudC1UeXBlOiB0ZXh0L3BsYWluOyBjaGFyc2V0PSJ1dGYtOCIKQ29udGVudC1UcmFuc2Zlci1FbmNvZGluZzogN2JpdApNSU1FLVZlcnNpb246IDEuMApUbzogZ2R1c2VyMUB3b3Jrc3BhY2VzYW1wbGVzLmRldgpGcm9tOiBnZHVzZXIyQHdvcmtzcGFjZXNhbXBsZXMuZGV2ClN1YmplY3Q6IEF1dG9tYXRlZCBkcmFmdAoKVGhpcyBpcyBhdXRvbWF0ZWQgZHJhZnQgbWFpbAo=", // nolint:lll
			},
		},
		Payload{
			Message: payloadMessage{
				Raw: "Q29udGVudC1UeXBlOiB0ZXh0L3BsYWluOyBjaGFyc2V0PSJ1dGYtOCIKQ29udGVudC1UcmFuc2Zlci1FbmNvZGluZzogN2JpdApNSU1FLVZlcnNpb246IDEuMApUbzogZ2R1c2VyMUB3b3Jrc3BhY2VzYW1wbGVzLmRldgpGcm9tOiBnZHVzZXIyQHdvcmtzcGFjZXNhbXBsZXMuZGV2ClN1YmplY3Q6IEJyYW5kIG5ldyB1cGRhdGVkIHN1YmplY3QKCk1lc3NhZ2Ugd2FzIHVwZGF0ZWQgZnJvbSBwb3N0bWFuLg==", // nolint:lll
			},
		},
		testscenario.CRUDTestSuite{
			ReadFields:          datautils.NewSet("id"),
			RecordIdentifierKey: "id",
		},
	)
}
