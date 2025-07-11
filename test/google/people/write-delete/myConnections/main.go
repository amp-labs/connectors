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

var objectName = "myConnections"

type myConnectionsPayload struct {
	MyConnections myConnectionsObject `json:"myConnections"`
}

type myConnectionsObject struct {
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

	slog.Info("> TEST Create/Update/Delete myConnections")
	slog.Info("Creating myConnections")

	name := gofakeit.Name()
	createMyConnections(ctx, conn, &myConnectionsPayload{
		MyConnections: myConnectionsObject{
			Name: name,
		},
	})

	slog.Info("Reading myConnections")

	res := readMyConnections(ctx, conn)

	slog.Info("Finding recently created myConnections")

	myConnections := searchMyConnections(res, "name", name)
	myConnectionsID := fmt.Sprintf("%v", myConnections["id"])
	myConnectionsEtag := fmt.Sprintf("%v", myConnections["etag"])

	slog.Info("Updating myConnections name")

	updatedName := gofakeit.Name()
	updateMyConnections(ctx, conn, myConnectionsID, &myConnectionsPayload{
		MyConnections: myConnectionsObject{
			Name: updatedName,
			Etag: myConnectionsEtag,
		},
	})

	slog.Info("View that myConnections has changed accordingly")

	res = readMyConnections(ctx, conn)

	myConnections = searchMyConnections(res, "id", myConnectionsID)

	myConnectionsName, ok := myConnections["name"].(string)
	if !ok || myConnectionsName != updatedName {
		utils.Fail("error updated backgroundColor doesn't match")
	}

	slog.Info("Removing this myConnections")
	removeMyConnections(ctx, conn, myConnectionsID)
	slog.Info("> Successful test completion")
}

func searchMyConnections(res *common.ReadResult, key, value string) map[string]any {
	for _, data := range res.Data {
		if mockutils.DoesObjectCorrespondToString(data.Fields[key], value) {
			return data.Fields
		}
	}

	utils.Fail("error finding myConnections")

	return nil
}

func readMyConnections(ctx context.Context, conn *google.Connector) *common.ReadResult {
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

func createMyConnections(ctx context.Context, conn *google.Connector, payload *myConnectionsPayload) {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   "",
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to google", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to create a myConnections")
	}
}

func updateMyConnections(ctx context.Context, conn *google.Connector, myConnectionsID string, payload *myConnectionsPayload) {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   myConnectionsID,
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to google", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to update a myConnections")
	}
}

func removeMyConnections(ctx context.Context, conn *google.Connector, myConnectionsID string) {
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: objectName,
		RecordId:   myConnectionsID,
	})
	if err != nil {
		utils.Fail("error deleting for google", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to remove a myConnections")
	}
}
