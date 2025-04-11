package monday

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	defaultPageSize   = 200
	mondayObjectBoard = "boards"
	mondayObjectUser  = "users"
)

var (
	// ErrUnsupportedObject is returned when an unsupported object type is requested.
	ErrUnsupportedObject = errors.New("unsupported object")
	// ErrInvalidResponseFormat is returned when the API response format is unexpected.
	ErrInvalidResponseFormat = errors.New("invalid response format")
	// ErrUnsupportedObjectName is returned when an unsupported object name is provided.
	ErrUnsupportedObjectName = errors.New("unsupported object name")
	// ErrBoardNameRequired is returned when board name is missing for creation.
	ErrBoardNameRequired = errors.New("board name is required for creation")
	// ErrWriteUserNotSupported is returned when attempting to write user data.
	ErrWriteUserNotSupported = errors.New("write user not supported")
)

// Record ID paths in GraphQL response.
const (
	mondayBoardsIDPath = "data.create_board.id"
	mondayUsersIDPath  = "data.create_user.id"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	// Use introspection query to get field information
	query := fmt.Sprintf(`{
		__type(name: "%s") {
			name
			fields {
				name
				type {
					name
					kind
					ofType {
						name
					}
				}
			}
		}
	}`, naming.NewSingularString(naming.CapitalizeFirstLetterEveryWord(objectName)).String())

	// Create the request body as a map
	requestBody := map[string]string{
		"query": query,
	}

	// Marshal the request body to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	url, err := c.ModuleClient.URL()
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		FieldsMap:   make(map[string]string),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	metadataResp, err := common.UnmarshalJSON[MetadataResponse](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if len(metadataResp.Data.Type.Fields) == 0 {
		return nil, fmt.Errorf(
			"missing or empty fields for object: %s, error: %w",
			objectName,
			common.ErrMissingExpectedValues,
		)
	}

	// Process each field from the introspection result
	for _, field := range metadataResp.Data.Type.Fields {
		objectMetadata.FieldsMap[field.Name] = field.Name
	}

	return &objectMetadata, nil
}

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
	var page *int
	var limit int

	if params.NextPage != "" {
		// Parse the page number from NextPage
		var pageNum int

		_, err := fmt.Sscanf(string(params.NextPage), "%d", &pageNum)
		if err != nil {
			return nil, fmt.Errorf("invalid next page format: %w", err)
		}

		page = &pageNum
		limit = defaultPageSize
	}

	query, err := getQueryForObject(params.ObjectName, page, &limit)
	if err != nil {
		return nil, err
	}

	requestBody := map[string]string{
		"query": query,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	url, err := c.ModuleClient.URL()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	data, err := common.UnmarshalJSON[MondayResponse](resp)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	var records []any

	switch params.ObjectName {
	case mondayObjectUser:
		if len(data.Data.Users) == 0 {
			errMsg := "missing expected values for object: " + params.ObjectName

			return nil, fmt.Errorf("%s, error: %w", errMsg, common.ErrMissingExpectedValues)
		}

		records = make([]any, len(data.Data.Users))
		for i, user := range data.Data.Users {
			records[i] = user
		}
	case mondayObjectBoard:
		if len(data.Data.Boards) == 0 {
			errMsg := "missing expected values for object: " + params.ObjectName

			return nil, fmt.Errorf("%s, error: %w", errMsg, common.ErrMissingExpectedValues)
		}

		records = make([]any, len(data.Data.Boards))
		for i, board := range data.Data.Boards {
			records[i] = board
		}
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedObjectName, params.ObjectName)
	}

	return common.ParseResult(
		resp,
		getRecords(params.ObjectName),
		makeNextRecordsURL(params, len(records)),
		common.GetMarshaledData,
		params.Fields,
	)
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

	var mutation string

	switch params.ObjectName {
	case mondayObjectBoard:
		if params.RecordId == "" {
			boardName, ok := recordData["name"].(string)
			if !ok {
				return nil, ErrBoardNameRequired
			}

			mutation = fmt.Sprintf(`mutation {
				create_board(board_name: "%s", board_kind: public) {
					id
					name
				}
			}`, boardName)
		} else {
			mutation = fmt.Sprintf(`mutation {
				update_board(board_id: %s, board_attribute: name, new_value: "%v") {
					id
				}
			}`, params.RecordId, recordData["name"])
		}
	case mondayObjectUser:
		return nil, ErrWriteUserNotSupported
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedObjectName, params.ObjectName)
	}

	requestBody := map[string]string{
		"query": mutation,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	url, err := c.ModuleClient.URL()
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
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

func extractRecordID(node *ajson.Node, objectName string) (string, error) {
	createRecordIDPaths := map[string]string{
		mondayObjectBoard: mondayBoardsIDPath,
		mondayObjectUser:  mondayUsersIDPath,
	}

	idPath, valid := createRecordIDPaths[objectName]
	if !valid {
		return "", fmt.Errorf("%w: %s", common.ErrOperationNotSupportedForObject, objectName)
	}

	rawID, err := jsonquery.New(node).IntegerOptional(idPath)
	if err != nil {
		return "", err
	}

	if rawID == nil {
		return "", nil
	}

	return strconv.FormatInt(*rawID, 10), nil
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

	recordId, err := extractRecordID(node, params.ObjectName)
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
	url, err := c.ModuleClient.URL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), nil)
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
