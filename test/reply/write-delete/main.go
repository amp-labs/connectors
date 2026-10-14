// Live scenario: full create/update/delete lifecycles across objects —
// contacts (PATCH update), sequences (PATCH, create needs a steps entry),
// contact-lists and sequence-folders (PUT updates), plus a custom field
// definition PUT. Every created record is deleted at the end.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/amp-labs/connectors/common"
	replyconn "github.com/amp-labs/connectors/providers/reply"
	"github.com/amp-labs/connectors/test/reply"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := reply.GetReplyConnector(ctx)

	// Unique per-run suffix: Reply rejects duplicate names on some objects
	// (e.g. sequence folders return 409 sequenceFolder.duplicateName).
	suffix := fmt.Sprintf(" %d", time.Now().Unix())

	// Create.
	created := write(ctx, conn, "create contact (POST)", common.WriteParams{
		ObjectName: "contacts",
		RecordData: map[string]any{
			"email":     "live.writedelete@example.com",
			"firstName": "Live",
			"lastName":  "WriteDelete",
		},
	})

	// Update via PATCH.
	write(ctx, conn, "update contact (PATCH)", common.WriteParams{
		ObjectName: "contacts",
		RecordId:   created.RecordId,
		RecordData: map[string]any{"title": "Updated by live test"},
	})

	// contact-lists lifecycle (PUT update path).
	list := write(ctx, conn, "create contact list (POST)", common.WriteParams{
		ObjectName: "contact-lists",
		RecordData: map[string]any{"name": "Live WriteDelete List" + suffix},
	})
	write(ctx, conn, "update contact list (PUT)", common.WriteParams{
		ObjectName: "contact-lists",
		RecordId:   list.RecordId,
		RecordData: map[string]any{"name": "Live WriteDelete List (renamed)" + suffix},
	})
	remove(ctx, conn, "contact-lists", list.RecordId)

	// sequence-folders lifecycle (PUT update path).
	folder := write(ctx, conn, "create sequence folder (POST)", common.WriteParams{
		ObjectName: "sequence-folders",
		RecordData: map[string]any{"name": "Live WriteDelete Folder" + suffix},
	})
	write(ctx, conn, "update sequence folder (PUT)", common.WriteParams{
		ObjectName: "sequence-folders",
		RecordId:   folder.RecordId,
		RecordData: map[string]any{"name": "Live WriteDelete Folder (renamed)" + suffix},
	})
	remove(ctx, conn, "sequence-folders", folder.RecordId)

	// sequences lifecycle: create requires at least one steps entry.
	sequence := write(ctx, conn, "create sequence (POST)", common.WriteParams{
		ObjectName: "sequences",
		RecordData: map[string]any{
			"name":  "Live WriteDelete Sequence" + suffix,
			"steps": []any{map[string]any{"type": "task", "taskNote": "live write-delete test"}},
		},
	})
	write(ctx, conn, "update sequence (PATCH)", common.WriteParams{
		ObjectName: "sequences",
		RecordId:   sequence.RecordId,
		RecordData: map[string]any{"name": "Live WriteDelete Sequence (renamed)" + suffix},
	})
	remove(ctx, conn, "sequences", sequence.RecordId)

	// Update a custom field definition via PUT (verb exception path).
	write(ctx, conn, "update custom field (PUT)", common.WriteParams{
		ObjectName: "custom-fields",
		RecordId:   "150697",
		RecordData: map[string]any{"title": "Deal Stage", "fieldType": "text"},
	})

	// Delete the contact created at the start.
	remove(ctx, conn, "contacts", created.RecordId)
}

func remove(ctx context.Context, conn *replyconn.Connector, objectName, recordID string) {
	deleted, err := conn.Delete(ctx, common.DeleteParams{
		ObjectName: objectName,
		RecordId:   recordID,
	})
	if err != nil {
		utils.Fail("error deleting from Reply", "object", objectName, "error", err)
	}

	slog.Info("Delete result", "object", objectName, "recordId", recordID)
	utils.DumpJSON(deleted, os.Stdout)
}

func write(
	ctx context.Context, conn *replyconn.Connector, name string, params common.WriteParams,
) *common.WriteResult {
	res, err := conn.Write(ctx, params)
	if err != nil {
		utils.Fail("error writing to Reply", "operation", name, "error", err)
	}

	slog.Info("Write result", "operation", name, "object", params.ObjectName)
	utils.DumpJSON(res, os.Stdout)

	return res
}
