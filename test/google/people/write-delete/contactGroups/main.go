package main

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/google"
	connTest "github.com/amp-labs/connectors/test/google"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/brianvoe/gofakeit/v6"
)

var objectName = "contactGroups"

type contactGroupPayload struct {
	ContactGroup contactGroupObject `json:"contactGroup"`
}

type contactGroupObject struct {
	Name string `json:"name"`
	Etag string `json:"etag,omitempty"` // required for update
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetGoogleConnector(ctx, google.ModulePeople)

	slog.Info("> TEST Create/Update/Delete contactGroup")
	slog.Info("Creating contactGroup")

	name := gofakeit.Name()
	createContactGroup(ctx, conn, &contactGroupPayload{
		ContactGroup: contactGroupObject{
			Name: name,
		},
	})

	slog.Info("Reading contactGroup")

	res := readContactGroup(ctx, conn)

	slog.Info("Finding recently created contactGroup")

	contactGroup := searchContactGroup(res, "name", name)
	contactGroupID := fmt.Sprintf("%v", contactGroup["id"])
	contactGroupEtag := fmt.Sprintf("%v", contactGroup["etag"])

	slog.Info("Updating contactGroup name")

	updatedName := gofakeit.Name()
	updateContactGroup(ctx, conn, contactGroupID, &contactGroupPayload{
		ContactGroup: contactGroupObject{
			Name: updatedName,
			Etag: contactGroupEtag,
		},
	})

	slog.Info("View that contactGroup has changed accordingly")

	res = readContactGroup(ctx, conn)

	contactGroup = searchContactGroup(res, "id", contactGroupID)

	contactGroupName, ok := contactGroup["name"].(string)
	if !ok || contactGroupName != updatedName {
		utils.Fail("error updated backgroundColor doesn't match")
	}

	slog.Info("Removing this contactGroup")
	removeContactGroup(ctx, conn, contactGroupID)
	slog.Info("> Successful test completion")
}

func searchContactGroup(res *common.ReadResult, key, value string) map[string]any {
	for _, data := range res.Data {
		if mockutils.DoesObjectCorrespondToString(data.Fields[key], value) {
			return data.Fields
		}
	}

	utils.Fail("error finding contactGroup")

	return nil
}

func readContactGroup(ctx context.Context, conn *google.Connector) *common.ReadResult {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields: connectors.Fields(
			"id", "etag", "name",
		),
	})
	if err != nil {
		utils.Fail("error reading from google", "error", err)
	}

	return res
}

func createContactGroup(ctx context.Context, conn *google.Connector, payload *contactGroupPayload) {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   "",
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to google", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to create a contactGroup")
	}
}

func updateContactGroup(ctx context.Context, conn *google.Connector, contactGroupID string, payload *contactGroupPayload) {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   contactGroupID,
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to google", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to update a contactGroup")
	}
}

func removeContactGroup(ctx context.Context, conn *google.Connector, contactGroupID string) {
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: objectName,
		RecordId:   contactGroupID,
	})
	if err != nil {
		utils.Fail("error deleting for google", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to remove a contactGroup")
	}
}
