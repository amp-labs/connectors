package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/pipeliner"
	connTest "github.com/amp-labs/connectors/test/pipeliner"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
)

type NotesPayload struct {
	Note      string `json:"note"`
	OwnerId   string `json:"owner_id"`
	ContactId string `json:"contact_id"`
}

func main() {
	// Handle Ctrl-C gracefully.
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	// Set up slog logging.
	utils.SetupLogging()

	conn := connTest.GetPipelinerConnector(ctx)

	fmt.Println("> TEST Create/Update/Delete Notes")
	fmt.Println("Creating Notes")

	ownerID := getFirstObjectID(ctx, conn, "Clients")
	contactID := getFirstObjectID(ctx, conn, "Contacts")
	view := createNotes(ctx, conn, &NotesPayload{
		Note:      "important issue to resolve due 19th of July",
		OwnerId:   ownerID,
		ContactId: contactID,
	})

	fmt.Println("Updating some Notes properties")
	updateNotes(ctx, conn, view.RecordId, &NotesPayload{
		Note:      "Task due 19th of July",
		OwnerId:   ownerID,
		ContactId: contactID,
	})

	fmt.Println("View that Notes has changed accordingly")

	res := readNotes(ctx, conn)

	updatedView := searchNotes(res, "id", view.RecordId)
	for k, v := range map[string]string{
		"note":       "Task due 19th of July",
		"owner_id":   ownerID,
		"contact_id": contactID,
	} {
		if !mockutils.DoesObjectCorrespondToString(updatedView[k], v) {
			utils.Fail("error updated properties do not match", k, v, updatedView[k])
		}
	}

	fmt.Println("Removing this Notes")
	removeNotes(ctx, conn, view.RecordId)
	fmt.Println("> Successful test completion")
}

func searchNotes(res *common.ReadResult, key, value string) map[string]any {
	for _, data := range res.Data {
		if mockutils.DoesObjectCorrespondToString(data.Fields[key], value) {
			return data.Raw
		}
	}

	utils.Fail("error finding Notes")

	return nil
}

func readNotes(ctx context.Context, conn *pipeliner.Connector) *common.ReadResult {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: "Notes",
		Fields:     connectors.Fields("id", "view", "name"),
	})
	if err != nil {
		utils.Fail("error reading from Pipeliner", "error", err)
	}

	return res
}

func getFirstObjectID(ctx context.Context, conn *pipeliner.Connector, name string) string {
	res, err := conn.Read(ctx, common.ReadParams{
		ObjectName: name,
		Fields:     connectors.Fields("id"),
	})
	if err != nil {
		utils.Fail("error reading from Pipeliner", "error", err)
	}

	data := res.Data
	if len(data) == 0 {
		utils.Fail("error Pipeliner has no objects", "error", err)
	}

	ownerID, found := data[0].Fields["id"]
	if !found {
		utils.Fail("error Pipeliner object has no id")
	}

	result, ok := ownerID.(string)
	if !ok {
		utils.Fail("error Pipeliner object id is not a string")
	}

	return result
}

func createNotes(ctx context.Context, conn *pipeliner.Connector, payload *NotesPayload) *common.WriteResult {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "Notes",
		RecordId:   "",
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to Pipeliner", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to create a Notes")
	}

	return res
}

func updateNotes(ctx context.Context, conn *pipeliner.Connector, viewID string, payload *NotesPayload) *common.WriteResult {
	res, err := conn.Write(ctx, common.WriteParams{
		ObjectName: "Notes",
		RecordId:   viewID,
		RecordData: payload,
	})
	if err != nil {
		utils.Fail("error writing to Pipeliner", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to update a Notes")
	}

	return res
}

func removeNotes(ctx context.Context, conn *pipeliner.Connector, viewID string) {
	res, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: "Notes",
		RecordId:   viewID,
	})
	if err != nil {
		utils.Fail("error deleting for Pipeliner", "error", err)
	}

	if !res.Success {
		utils.Fail("failed to remove a Notes")
	}
}
