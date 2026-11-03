package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	grain "github.com/amp-labs/connectors/test/grain"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/testscenario"
)

// Grain has no delete endpoint for recordings or uploads. The recording title is restored after the
// update; the staged upload is left behind, and Grain periodically deletes uploads that never reach success.
// https://developers.grain.com/uploads.html#expiration-cleanup
func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	utils.SetupLogging()

	conn := grain.GetGrainConnector(ctx)

	testscenario.ValidateUpdate(ctx, conn, "recordings",
		map[string]any{"title": "Ampersand connector test"},
		testscenario.UpdateTestSuite{
			ReadFields:            datautils.NewSet("id", "title"),
			RecordIdentifierKey:   "id",
			UpdatedFields:         map[string]string{"title": "Ampersand connector test"},
			RestoreOriginalFields: true,
		},
	)

	slog.Info("> TEST Create upload")

	created, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "uploads",
		RecordData: map[string]any{
			"media_type":   "transcript",
			"auto_promote": false,
			"attributes": map[string]any{
				"title": "Ampersand connector test",
				"transcript": map[string]any{
					"language": "en",
					"segments": []any{
						map[string]any{"speaker_key": "s1", "start": 0, "end": 1200, "text": "Connector test."},
					},
				},
			},
		},
	})
	if err != nil {
		utils.Fail("error creating upload", "error", err)
	}

	utils.DumpJSON(created, os.Stdout)

	slog.Info("> TEST Update upload", "id", created.RecordId)

	updated, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "uploads",
		RecordId:   created.RecordId,
		RecordData: map[string]any{
			"attributes": map[string]any{"title": "Ampersand connector test (updated)"},
		},
	})
	if err != nil {
		utils.Fail("error updating upload", "error", err)
	}

	utils.DumpJSON(updated, os.Stdout)

	if updated.Data["title"] != "Ampersand connector test (updated)" {
		utils.Fail("error updated upload title does not match", "title", updated.Data["title"])
	}
}
