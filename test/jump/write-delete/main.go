package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/jump"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

// Creates, updates and deletes one amp-test contact as the user in the creds file "metadata.userId".
// Any error, including a 429, stops the run: Jump rate limits per IP and escalates repeated 429s to long lockouts.
func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetJumpConnector(ctx)

	suffix := time.Now().UTC().Format("20060102150405")
	email := fmt.Sprintf("amp-test-write-%s@example.com", suffix)
	updatedEmail := fmt.Sprintf("amp-test-write-%s-updated@example.com", suffix)

	testscenario.ValidateCreateUpdateDelete(ctx, conn,
		"contacts",
		map[string]any{
			"firstName": "amp-test",
			"lastName":  "Write " + suffix,
			"email":     email,
		},
		map[string]any{
			"lastName": "Write " + suffix + " Updated",
			"email":    updatedEmail,
		},
		testscenario.CRUDTestSuite{
			ReadFields: datautils.NewSet("id", "name", "email"),
			SearchBy: testscenario.Property{
				Key:   "email",
				Value: email,
				Since: time.Now().Add(-1 * time.Minute),
			},
			RecordIdentifierKey: "id",
			UpdatedFields: map[string]string{
				"name":  "amp-test Write " + suffix + " Updated",
				"email": updatedEmail,
			},
		},
	)
}
