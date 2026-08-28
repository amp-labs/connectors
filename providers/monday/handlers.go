package monday

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	defaultPageSize   = 200
	mondayObjectBoard = "boards"
	mondayObjectUser  = "users"
	mondayObjectItems = "items"
	mondayObjectDocs  = "docs"
)

var (
	// ErrUnsupportedObjectName is returned when an unsupported object name is provided.
	ErrUnsupportedObjectName = errors.New("unsupported object name")
	// ErrBoardNameRequired is returned when board name is missing for creation.
	ErrBoardNameRequired = errors.New("board name is required for creation")
	// ErrWriteUserNotSupported is returned when attempting to write user data.
	ErrWriteUserNotSupported = errors.New("write user not supported")
)


func getBoardsBaseFields() string {
	return `
		id
		name
		state
		permissions
		items_count
		type
		updated_at
		url
		workspace_id`
}

func getBoardsNestedFields() string {
	return `
		columns {
			id
			title
			type
		}
		groups {
			id
			title
			position
		}
		owner {
			id
			name
		}
		owners {
			id
			name
		}
		subscribers {
			id
			name
		}
		tags {
			id
			name
		}
		team_owners {
			id
			name
		}
		team_subscribers {
			id
			name
		}
		top_group {
			id
			title
		}
		updates {
			id
			body
			created_at
		}
		views {
			id
			name
			type
		}
		workspace {
			id
			name
		}`
}

func getBoardsQuery(page *int, limit *int) string {
	paginationParams := ""
	if page != nil && limit != nil {
		paginationParams = fmt.Sprintf("(limit: %d, page: %d)", *limit, *page)
	}

	return fmt.Sprintf(`query {
		boards%s {
			%s
			%s
		}
	}`, paginationParams, getBoardsBaseFields(), getBoardsNestedFields())
}

func getUsersQuery(page *int, limit *int) string {
	paginationParams := ""
	if page != nil && limit != nil {
		paginationParams = fmt.Sprintf("(limit: %d, page: %d)", *limit, *page)
	}

	return fmt.Sprintf(`query {
		users%s {
			id
			email
			name
			enabled
		}
	}`, paginationParams)
}

func getQueryForObject(objectName string, page *int, limit *int) (string, error) {
	switch objectName {
	case mondayObjectBoard:
		return getBoardsQuery(page, limit), nil
	case mondayObjectUser:
		return getUsersQuery(page, limit), nil
	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedObjectName, objectName)
	}
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	query, err := buildReadQuery(params)
	if err != nil {
		return nil, err
	}

	return buildGraphQLHTTPRequest(ctx, c.ProviderInfo().BaseURL, query)
}

func (c *Connector) parseReadResponse(
	_ context.Context,
	params common.ReadParams,
	_ *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	if params.ObjectName == mondayObjectItems {
		return parseItemsReadResponse(params, resp)
	}

	return parseStandardReadResponse(params, resp)
}

func parseItemsReadResponse(params common.ReadParams, resp *common.JSONHTTPResponse) (*common.ReadResult, error) {
	limit := params.PageSize
	if limit <= 0 {
		limit = defaultPageSize
	}

	return common.ParseResult(
		resp,
		extractItemsRecords,
		makeItemsNextRecordsURL(limit),
		common.MakeMarshaledDataFunc(itemReadRecordCustomFieldsTransformer),
		params.Fields,
	)
}

func parseStandardReadResponse(params common.ReadParams, resp *common.JSONHTTPResponse) (*common.ReadResult, error) {
	data, err := common.UnmarshalJSON[Response](resp)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	recordCount, err := validateStandardReadRecords(params.ObjectName, *data)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		resp,
		getRecords(params.ObjectName),
		makeNextRecordsURL(params, recordCount),
		common.GetMarshaledData,
		params.Fields,
	)
}

func validateStandardReadRecords(objectName string, data Response) (int, error) {
	switch objectName {
	case mondayObjectUser:
		if len(data.Data.Users) == 0 {
			return 0, fmt.Errorf(
				"missing expected values for object: %s, error: %w",
				objectName,
				common.ErrMissingExpectedValues,
			)
		}

		return len(data.Data.Users), nil
	case mondayObjectBoard:
		if len(data.Data.Boards) == 0 {
			return 0, fmt.Errorf(
				"missing expected values for object: %s, error: %w",
				objectName,
				common.ErrMissingExpectedValues,
			)
		}

		return len(data.Data.Boards), nil
	default:
		return 0, fmt.Errorf("%w: %s", ErrUnsupportedObjectName, objectName)
	}
}

func getRecords(objectName string) func(*ajson.Node) ([]map[string]any, error) {
	return func(node *ajson.Node) ([]map[string]any, error) {
		// First get the data object
		dataNode, err := node.GetKey("data")
		if err != nil {
			return nil, err
		}

		// Then get the array under the object name (e.g., "boards" or "users")
		records, err := jsonquery.New(dataNode).ArrayOptional(objectName)
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(records)
	}
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	recordData, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert record data to map: %w", err)
	}

	mutation, err := c.buildWriteMutation(ctx, params, recordData)
	if err != nil {
		return nil, err
	}

	return buildGraphQLHTTPRequest(ctx, c.ProviderInfo().BaseURL, mutation)
}

func extractResponseErrors(node *ajson.Node) ([]any, error) {
	errors, err := jsonquery.New(node).ArrayOptional("errors")
	if err != nil {
		return nil, err
	}

	if len(errors) == 0 {
		return nil, nil
	}

	errorMsgs := make([]any, 0, len(errors))

	for _, e := range errors {
		if msg, err := e.GetKey("message"); err == nil {
			errorMsgs = append(errorMsgs, msg.String())
		}
	}

	return errorMsgs, nil
}

func extractRecordID(node *ajson.Node, objectName string, recordID string) (string, error) {
	type mutationPath struct {
		parent string
		child  string
	}

	paths := map[string]mutationPath{
		mondayObjectBoard: {parent: "create_board", child: "id"},
		mondayObjectUser:  {parent: "create_user", child: "id"},
		mondayObjectItems: {parent: "create_item", child: "id"},
	}

	path, valid := paths[objectName]
	if !valid {
		return "", fmt.Errorf("%w: %s", common.ErrOperationNotSupportedForObject, objectName)
	}

	if objectName == mondayObjectItems && recordID != "" {
		path = mutationPath{parent: "change_multiple_column_values", child: "id"}
	}

	rawID, err := jsonquery.New(node, "data", path.parent).TextWithDefault(path.child, "")
	if err != nil {
		return "", err
	}

	if rawID != "" {
		return rawID, nil
	}

	if recordID != "" {
		return recordID, nil
	}

	return "", nil
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	node, ok := resp.Body()
	if !ok {
		return &common.WriteResult{Success: true}, nil
	}

	if errors, err := extractResponseErrors(node); err != nil {
		return nil, err
	} else if errors != nil {
		return &common.WriteResult{
			Success: false,
			Errors:  errors,
		}, nil
	}

	recordId, err := extractRecordID(node, params.ObjectName, params.RecordId)
	if err != nil {
		return nil, err
	}

	data, err := common.UnmarshalJSON[map[string]any](resp)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordId,
		Data:     *data,
	}, nil
}

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	if params.ObjectName != mondayObjectItems {
		return nil, fmt.Errorf("%w: %s", common.ErrOperationNotSupportedForObject, params.ObjectName)
	}

	if params.RecordId == "" {
		return nil, common.ErrMissingRecordID
	}

	return buildGraphQLHTTPRequest(ctx, c.ProviderInfo().BaseURL, getDeleteItemMutation(params.RecordId))
}

func (c *Connector) parseDeleteResponse(
	ctx context.Context,
	params common.DeleteParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	if resp.Code != http.StatusOK && resp.Code != http.StatusNoContent {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, resp.Code)
	}

	// A successful delete returns 200 OK
	return &common.DeleteResult{
		Success: true,
	}, nil
}
