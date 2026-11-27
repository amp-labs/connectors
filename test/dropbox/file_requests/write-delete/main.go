package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/internal/datautils"
	connTest "github.com/amp-labs/connectors/test/dropbox"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
	"github.com/brianvoe/gofakeit/v6"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := connTest.GetDropboxConnector(ctx)

	destination := os.Getenv("DROPBOX_FILE_REQUEST_DESTINATION")
	if destination == "" {
		utils.Fail("set DROPBOX_FILE_REQUEST_DESTINATION to a folder path in the connected account (e.g. /FileRequests)")
	}

	title := fmt.Sprintf("ampersand-fr-%s", gofakeit.LetterN(8))
	updatedTitle := title + "-updated"

	// Update must set open:false before delete (Dropbox only removes closed file requests).
	testscenario.ValidateCreateUpdateDelete(ctx, conn, "file_requests",
		map[string]any{
			"title":       title,
			"destination": destination,
			"open":        true,
		},
		map[string]any{
			"title": updatedTitle,
			"open":  false,
		},
		testscenario.CRUDTestSuite{
			ReadFields:          datautils.NewSet("id", "title", "is_open", "url"),
			RecordIdentifierKey: "id",
			SearchBy: testscenario.Property{
				Key:   "title",
				Value: title,
			},
			UpdatedFields: map[string]string{
				"title": updatedTitle,
			},
			ValidateUpdatedFields: func(record map[string]any) {
				if !mockutils.DoesObjectCorrespondToString(record["title"], updatedTitle) {
					utils.Fail("title did not update", "title", updatedTitle, record["title"])
				}
				open, ok := record["is_open"].(bool)
				if !ok || open {
					utils.Fail("file request must be closed before delete", "is_open", record["is_open"])
				}
			},
		},
	)
}
