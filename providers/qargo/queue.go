package qargo

import (
	"context"
	"slices"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/qargo/internal/metadata"
	"github.com/spyzhov/ajson"
)

// Sales invoices and credit notes have no list endpoint. Qargo exposes them to accounting integrations
// only through the sync-task queue: GET /v1/accounting/sync-tasks returns every task waiting for this
// integration, and GET /v1/accounting/sync-tasks/{id} returns one with the document in payload.instance.
// A task stays in the queue until the integration reports it with
// POST /v1/accounting/sync-tasks/{id}/update-status, which builders do through the proxy.
// https://api-docs.qargo.com/docs/section/accounting/synchronize-invoicescredit-notes
//
// The queue has no paging, no order and no time filter, and documents carry no update time, so:
//   - every read returns every task still waiting, whatever Since and Until are;
//   - pages are cut by the connector: the queue is listed again for each page, its task ids sorted,
//     and the page holds the next PageSize ids after the last one read (the NextPage token). Paging is
//     by id, so a task completed while the read runs only drops out, and a new task with a lower id is
//     read on the next run, since it stays in the queue;
//   - the read ends when no waiting task id is greater than the last one read.
//
// Each record costs one request for the task's detail, against Qargo's 2 requests per second.
const defaultQueuePageSize = 50

func (c *Connector) parseQueueResponse(
	ctx context.Context,
	params common.ReadParams,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	body, ok := response.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	taskIDs, err := waitingTasks(body, queueTaskTypes[params.ObjectName])
	if err != nil {
		return nil, err
	}

	page, next := queuePage(taskIDs, params.NextPage.String(),
		readhelper.PageSizeWithDefault(params, defaultQueuePageSize))

	tasks := make([]*ajson.Node, 0, len(page))

	for _, taskID := range page {
		task, err := c.fetchTask(ctx, params.ObjectName, taskID)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	// The record is the document; Raw keeps the whole task, whose id builders need to report it.
	rows, err := readhelper.MakeMarshaledDataFuncWithId(
		taskDocument, readhelper.NewNestedIdField([]string{"payload", "instance"}, "id"),
	)(tasks, params.Fields.List())
	if err != nil {
		return nil, err
	}

	return &common.ReadResult{
		Rows:     int64(len(rows)),
		Data:     rows,
		NextPage: common.NextPageToken(next),
		Done:     next == "",
	}, nil
}

// waitingTasks returns the sorted ids of the listed tasks of one type.
func waitingTasks(body *ajson.Node, taskType string) ([]string, error) {
	tasks, err := jsonquery.New(body).ArrayRequired("tasks")
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(tasks))

	for _, task := range tasks {
		query := jsonquery.New(task)

		kind, err := query.StringRequired("task_type")
		if err != nil {
			return nil, err
		}

		if kind != taskType {
			continue
		}

		id, err := query.StringRequired("id")
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	slices.Sort(ids)

	return ids, nil
}

// queuePage returns up to size ids after the last one read, and the token for the next page.
func queuePage(ids []string, lastRead string, size int) ([]string, string) {
	start := slices.IndexFunc(ids, func(id string) bool { return id > lastRead })
	if start == -1 {
		return nil, ""
	}

	end := min(start+size, len(ids))
	if end == len(ids) {
		return ids[start:end], ""
	}

	return ids[start:end], ids[end-1]
}

func (c *Connector) fetchTask(ctx context.Context, objectName, taskID string) (*ajson.Node, error) {
	path, err := metadata.Catalog.ReadPath(objectName)
	if err != nil {
		return nil, err
	}

	taskURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, path, taskID)
	if err != nil {
		return nil, err
	}

	response, err := c.JSONHTTPClient().Get(ctx, taskURL.String())
	if err != nil {
		return nil, err
	}

	task, ok := response.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	return task, nil
}

func taskDocument(task *ajson.Node) (map[string]any, error) {
	document, err := jsonquery.New(task, "payload").ObjectRequired("instance")
	if err != nil {
		return nil, err
	}

	return jsonquery.Convertor.ObjectToMap(document)
}
