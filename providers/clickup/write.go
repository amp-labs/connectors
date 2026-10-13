package clickup

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

type writeObjectSpec struct {
	createVerb string
	createPath string
	createId   readhelper.IdFieldQuery
	updateVerb string
	updatePath string
	updateId   readhelper.IdFieldQuery
	deletePath string
}

var writeObjects = map[string]writeObjectSpec{ // nolint:gochecknoglobals
	"channels": {
		// https://developer.clickup.com/reference/createchatchannel
		createVerb: http.MethodPost,
		createPath: "/v3/workspaces/{workspace_id}/chat/channels",
		createId:   readhelper.NewNestedIdField([]string{"data"}, "id"),
		// https://developer.clickup.com/reference/updatechatchannel
		updateVerb: http.MethodPatch,
		updatePath: "/v3/workspaces/{workspace_id}/chat/channels/{record_id}",
		updateId:   readhelper.NewNestedIdField([]string{"data"}, "id"),
		// https://developer.clickup.com/reference/deletechatchannel
		deletePath: "/v3/workspaces/{workspace_id}/chat/channels/{record_id}",
	},
	"docs": {
		// https://developer.clickup.com/reference/createdocpublic
		createVerb: http.MethodPost,
		createPath: "/v3/workspaces/{workspace_id}/docs",
		createId:   readhelper.NewIdField("id"),
	},
	"goals": {
		// https://developer.clickup.com/reference/creategoal
		createVerb: http.MethodPost,
		createPath: "/v2/team/{workspace_id}/goal",
		createId:   readhelper.NewNestedIdField([]string{"goal"}, "id"),
		// https://developer.clickup.com/reference/updategoal
		updateVerb: http.MethodPut,
		updatePath: "/v2/goal/{record_id}",
		updateId:   readhelper.NewNestedIdField([]string{"goal"}, "id"),
		// https://developer.clickup.com/reference/deletegoal
		deletePath: "/v2/goal/{record_id}",
	},
	"groups": {
		// https://developer.clickup.com/reference/createusergroup
		createVerb: http.MethodPost,
		createPath: "/v2/team/{workspace_id}/group",
		createId:   readhelper.NewIdField("id"),
		// https://developer.clickup.com/reference/updateteam
		updateVerb: http.MethodPut,
		updatePath: "/v2/group/{record_id}",
		updateId:   readhelper.NewIdField("id"),
		// https://developer.clickup.com/reference/deleteteam
		deletePath: "/v2/group/{record_id}",
	},
	"spaces": {
		// https://developer.clickup.com/reference/createspace
		createVerb: http.MethodPost,
		createPath: "/v2/team/{workspace_id}/space",
		createId:   readhelper.NewIdField("id"),
		// https://developer.clickup.com/reference/updatespace
		updateVerb: http.MethodPut,
		updatePath: "/v2/space/{record_id}",
		updateId:   readhelper.NewIdField("id"),
		// https://developer.clickup.com/reference/deletespace
		deletePath: "/v2/space/{record_id}",
	},
	"tasks": {
		// https://developer.clickup.com/reference/createtask
		createVerb: "",
		createPath: "",
		// https://developer.clickup.com/reference/updatetask
		updateVerb: http.MethodPut,
		updatePath: "/v2/task/{record_id}",
		updateId:   readhelper.NewIdField("id"),
		// https://developer.clickup.com/reference/deletetask
		deletePath: "/v2/task/{record_id}",
	},
	"time_entries": {
		// https://developer.clickup.com/reference/createatimeentry
		createVerb: http.MethodPost,
		createPath: "/v2/team/{workspace_id}/time_entries",
		createId:   readhelper.NewIdField("start"),
		// https://developer.clickup.com/reference/updateatimeentry
		updateVerb: http.MethodPut,
		updatePath: "/v2/team/{workspace_id}/time_entries/{record_id}",
		updateId:   readhelper.IdFieldQuery{}, // not available
		// https://developer.clickup.com/reference/deleteatimeentry
		deletePath: "/v2/team/{workspace_id}/time_entries/{record_id}",
	},
	"views": {
		createPath: "",
		createVerb: "",
		// https://developer.clickup.com/reference/updateview
		updateVerb: http.MethodPut,
		updatePath: "/v2/view/{record_id}",
		updateId:   readhelper.NewNestedIdField([]string{"view"}, "id"),
		// https://developer.clickup.com/reference/deleteview
		deletePath: "/v2/view/{record_id}",
	},
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	spec, ok := writeObjects[params.ObjectName]
	if !ok {
		return nil, common.ErrObjectNotSupported
	}

	var (
		url    *urlbuilder.URL
		err    error
		method string
	)

	if params.IsCreate() {
		if spec.createPath == "" {
			return nil, fmt.Errorf("%w: object %v", common.ErrOperationNotSupportedForObject, params.ObjectName)
		}

		method = spec.createVerb
		url, err = c.getWriteDeleteURL(spec.createPath, "")
	} else {
		if spec.updatePath == "" {
			return nil, fmt.Errorf("%w: object %v", common.ErrOperationNotSupportedForObject, params.ObjectName)
		}

		method = spec.updateVerb
		url, err = c.getWriteDeleteURL(spec.updatePath, params.RecordId)
	}

	if err != nil {
		return nil, err
	}

	reader, err := params.GetRecordReader()
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, method, url.String(), reader)
}

func (c *Connector) parseWriteResponse(
	ctx context.Context, //nolint:revive
	params common.WriteParams,
	request *http.Request, //nolint:revive
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	spec := writeObjects[params.ObjectName]

	var identifier readhelper.IdFieldQuery
	if params.IsCreate() {
		identifier = spec.createId
	} else {
		identifier = spec.updateId
	}

	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{ // nolint:nilerr
			Success: true,
		}, nil
	}

	nestedObject := body

	if len(identifier.Zoom) != 0 {
		size := len(identifier.Zoom)
		nestedId := identifier.Zoom[size-1]
		zoom := identifier.Zoom[:size-1]

		var err error

		nestedObject, err = jsonquery.New(body, zoom...).ObjectRequired(nestedId)
		if err != nil {
			return nil, err
		}
	}

	recordID := params.RecordId

	if field := identifier.Field; field != "" {
		var err error

		recordID, err = jsonquery.New(nestedObject).TextWithDefault(field, params.RecordId)
		if err != nil {
			return &common.WriteResult{ // nolint:nilerr
				Success: true,
			}, nil
		}
	}

	data, err := jsonquery.Convertor.ObjectToMap(nestedObject)
	if err != nil {
		return &common.WriteResult{ // nolint:nilerr
			Success:  true,
			RecordId: recordID,
		}, nil
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     data,
	}, nil
}
