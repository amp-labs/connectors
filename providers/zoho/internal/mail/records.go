package mail

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	// objectNameMessages / objectNameTasks are the objects a webhook event maps
	// to; they match the readable "messages" and "tasks" objects.
	objectNameMessages = "messages"
	objectNameTasks    = "tasks"

	// recordIDSeparator joins a parent id (folderId for messages, groupId for
	// tasks) with the record's own id into the composite id produced by RecordId
	// and consumed by GetRecordsByIds. All parts are numeric, so none contains a
	// slash.
	recordIDSeparator = "/"
)

var (
	errNoRecordIDs     = errors.New("no record ids provided")
	errInvalidRecordID = errors.New(`invalid zoho mail record id, expected "<folderId>/<messageId>"`)
)

// GetRecordsByIds fetches full records for the given record ids.
// Supported objects:
//
//   - "messages": each id is "<folderId>/<messageId>"; fetched from
//     api/accounts/{accountId}/folders/{folderId}/messages/{messageId}/details.
//
//   - "tasks": each id is "<groupId>/<taskId>" for a group task, or bare
//     "<taskId>" for a personal task; fetched from api/tasks/groups/{groupId}/
//     {taskId} or api/tasks/me/{taskId} respectively.
//
// https://www.zoho.com/mail/help/api/get-email-meta-data.html
// https://www.zoho.com/mail/help/api/get-single-task.html
func (a *Adapter) GetRecordsByIds(
	ctx context.Context, objectName string, recordIds []string, fields []string, _ []string,
) ([]common.ReadResultRow, error) {
	if objectName != objectNameMessages && objectName != objectNameTasks {
		return nil, fmt.Errorf("%w: %q", common.ErrGetRecordNotSupportedForObject, objectName)
	}

	if len(recordIds) == 0 {
		return nil, errNoRecordIDs
	}

	rows := make([]common.ReadResultRow, 0, len(recordIds))

	for _, recordID := range recordIds {
		var (
			row common.ReadResultRow
			err error
		)

		if objectName == objectNameTasks {
			row, err = a.getTaskByID(ctx, recordID, fields)
		} else {
			row, err = a.getMessageByID(ctx, recordID, fields)
		}

		if err != nil {
			return nil, err
		}

		rows = append(rows, row)
	}

	return rows, nil
}

// getMessageByID fetches a single email's metadata. The folderId is mandatory,
// so the record id must be the "<folderId>/<messageId>" composite.
func (a *Adapter) getMessageByID(
	ctx context.Context, recordID string, fields []string,
) (common.ReadResultRow, error) {
	folderID, messageID, hasFolder := splitCompositeID(recordID)
	if !hasFolder {
		return common.ReadResultRow{}, fmt.Errorf("%w: %q", errInvalidRecordID, recordID)
	}

	url, err := a.getAccountScopedURL("folders/" + folderID + "/messages/" + messageID + "/details")
	if err != nil {
		return common.ReadResultRow{}, err
	}

	resp, err := a.Client.Get(ctx, url.String())
	if err != nil {
		return common.ReadResultRow{}, err
	}

	return parseSingleRecord(resp, recordID, fields, "data")
}

// getTaskByID fetches a single task. A "<groupId>/<taskId>" composite hits the
// group endpoint; a bare "<taskId>" hits the personal endpoint.
func (a *Adapter) getTaskByID(
	ctx context.Context, recordID string, fields []string,
) (common.ReadResultRow, error) {
	groupID, taskID, hasGroup := splitCompositeID(recordID)

	var path string
	if hasGroup {
		path = "api/tasks/groups/" + groupID + "/" + taskID
	} else {
		path = "api/tasks/me/" + taskID
	}

	url, err := a.getAPIURL(path)
	if err != nil {
		return common.ReadResultRow{}, err
	}

	resp, err := a.Client.Get(ctx, url.String())
	if err != nil {
		return common.ReadResultRow{}, err
	}

	// The single-task response nests the record under data.tasks[].
	return parseSingleRecord(resp, recordID, fields, "data", "tasks")
}

// splitCompositeID parses a "<parent>/<child>" composite id. When there is no
// separator (or either side is empty) it returns the whole input as the child
// with hasParent=false.
func splitCompositeID(recordID string) (parent, child string, hasParent bool) {
	p, c, ok := strings.Cut(recordID, recordIDSeparator)
	if !ok || p == "" || c == "" {
		return "", recordID, false
	}

	return p, c, true
}

// parseSingleRecord extracts one record from a fetch response and returns it as
// a ReadResultRow stamped with the composite record id it was fetched by.
// recordsPath is the key path to the record: a single key ("data") yields an
// object; a path like ("data","tasks") yields the first element of an array.
func parseSingleRecord(
	resp *common.JSONHTTPResponse, recordID string, fields []string, recordsPath ...string,
) (common.ReadResultRow, error) {
	node, ok := resp.Body()
	if !ok {
		return common.ReadResultRow{}, common.ErrEmptyJSONHTTPResponse
	}

	record, err := extractRecord(node, recordsPath)
	if err != nil {
		return common.ReadResultRow{}, err
	}

	rows, err := common.GetMarshaledData([]map[string]any{record}, fields)
	if err != nil {
		return common.ReadResultRow{}, err
	}

	row := rows[0]
	// Correlate the row with the exact (possibly composite) id requested rather
	// than relying on GetMarshaledData's "id"-field heuristic.
	row.Id = recordID

	return row, nil
}

// extractRecord reads a single record map from node. A one-key path points at
// an object; a longer path points at an array whose first element is returned.
func extractRecord(node *ajson.Node, recordsPath []string) (map[string]any, error) {
	if len(recordsPath) == 1 {
		data, err := jsonquery.New(node).ObjectRequired(recordsPath[0])
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ObjectToMap(data)
	}

	records, err := extractRecordsFromKeyPath(recordsPath)(node)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	return records[0], nil
}
