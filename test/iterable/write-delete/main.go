package main

import (
	"context"
	"log/slog"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/iterable"
	connTest "github.com/amp-labs/connectors/test/iterable"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/brianvoe/gofakeit/v6"
)

type ListsPayload struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

var objectName = "lists" // nolint: gochecknoglobals

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetIterableConnector(ctx)

	slog.Info("> TEST Create/Delete list")
	slog.Info("Creating list")

	name := gofakeit.Name()
	createLists(ctx, conn, &ListsPayload{
		Name:        name,
		Description: gofakeit.Name(), // this should be a relatively short field
	})

	slog.Info("Reading lists")

	res := readLists(ctx, conn)

	slog.Info("Finding recently created list")

	list := searchList(res, "name", name)
	listID := listIdentifierAsString(list)

	slog.Info("Removing this list")
	removeLists(ctx, conn, listID)
	slog.Info("> Successful test completion")
}

func listIdentifierAsString(list map[string]any) string {
	id, ok := list["id"].(float64)
	if !ok {
		return ""
	}

	listIdentifier := int64(id)

	return strconv.FormatInt(listIdentifier, 10)
}

func searchList(res *common.ReadResult, key, value string) map[string]any {
	for _, data := range res.Data {
		if mockutils.DoesObjectCorrespondToString(data.Fields[key], value) {
			return data.Fields
		}
	}

	utils.Fail("error finding a list")

	return nil
}

func readLists(ctx context.Context, conn *iterable.Connector) *common.ReadResult {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields("id", "name", "description"),
	})
	if err != nil {
		utils.Fail("error reading from Iterable", "error", err)
	}

	return res
}

func createLists(ctx context.Context, conn *iterable.Connector, payload *ListsPayload) *common.WriteResult {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: objectName,
		RecordId:   "",
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to Iterable", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to create a list")
	}

	return res
}

func removeLists(ctx context.Context, conn *iterable.Connector, listID string) {
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: objectName,
		RecordId:   listID,
	})
	if err != nil {
		utils.Fail("error deleting for Iterable", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to remove a list")
	}
}
