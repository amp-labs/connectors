package fireflies

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

//go:embed graphql/*.graphql
var queryFS embed.FS

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

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
					  kind
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		Fields:      make(map[string]common.FieldMetadata),
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
		valueType := field.Type.Name

		if valueType == "" {
			valueType = field.Type.OfType.Name
		}

		objectMetadata.Fields[field.Name] = common.FieldMetadata{
			DisplayName:  field.Name,
			ValueType:    getFieldValueType(valueType),
			ProviderType: valueType,
			ReadOnly:     false,
			Values:       nil,
		}
	}

	return &objectMetadata, nil
}

func getFieldValueType(field string) common.ValueType {
	if field == "" {
		return ""
	}

	switch strings.ToLower(field) {
	case "float":
		return common.ValueTypeFloat
	case "string", "id":
		return common.ValueTypeString
	case "boolean":
		return common.ValueTypeBoolean
	default:
		return common.ValueTypeOther
	}
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL)
	if err != nil {
		return nil, err
	}

	var (
		skip  = 0
		limit int
		query string
	)

	if params.NextPage != "" {
		// Parse the page number from NextPage
		skip, err = strconv.Atoi(params.NextPage.String())
		if err != nil {
			return nil, err
		}
	}

	limit = defaultPageSize

	switch params.ObjectName {
	case transcriptsObjectName:
		query = getQuery(limit, skip, "graphql/transcripts.graphql", "transcriptsQuery")
	case bitesObjectName:
		query = getQuery(limit, skip, "graphql/bites.graphql", "bitesQuery")
	case usersObjectName:
		query = getQuery(0, 0, "graphql/users.graphql", "usersQuery")
	default:
		return nil, common.ErrObjectNotSupported
	}

	requestBody := map[string]string{
		"query": query,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	return req, nil
}

func getQuery(limit, skip int, filePath, queryName string) string {
	queryBytes, err := queryFS.ReadFile(filePath)
	if err != nil {
		return ""
	}

	tmpl, err := template.New(queryName).Parse(string(queryBytes))
	if err != nil {
		return ""
	}

	var (
		pageInfo PageInfo
		queryBuf bytes.Buffer
	)

	pageInfo.Limit = limit
	pageInfo.Skip = skip

	err = tmpl.Execute(&queryBuf, pageInfo)
	if err != nil {
		return ""
	}

	return queryBuf.String()
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	data, err := common.UnmarshalJSON[Response](resp)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	var (
		records      []any
		responseData []map[string]any
	)

	switch params.ObjectName {
	case usersObjectName:
		responseData = data.Data.Users
	case transcriptsObjectName:
		responseData = data.Data.Transcripts
	case bitesObjectName:
		responseData = data.Data.Bites
	default:
		return nil, fmt.Errorf("%w: %s", common.ErrObjectNotSupported, params.ObjectName)
	}

	if len(responseData) == 0 {
		errMsg := "missing expected values for object: " + params.ObjectName

		return nil, fmt.Errorf("%s, error: %w", errMsg, common.ErrMissingExpectedValues)
	}

	records = make([]any, len(responseData))
	for i, value := range responseData {
		records[i] = value
	}

	return common.ParseResult(
		resp,
		common.ExtractOptionalRecordsFromPath(params.ObjectName, "data"),
		makeNextRecordsURL(params, len(records)),
		common.GetMarshaledData,
		params.Fields,
	)
}

// nolint:gocognit,cyclop,funlen
func (c *Connector) buildWriteRequest(
	ctx context.Context, params common.WriteParams,
) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	recordData, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert record data to map: %w", err)
	}

	var mutation string

	switch params.ObjectName {
	case objectNameLiveMeeting:
		if params.RecordId == "" {
			mutation = getMutation("graphql/mutation_meeting.graphql", "liveMeeting", recordData)
		} else {
			return nil, ErrUpdateMeetingLinkNotSupported
		}
	case objectNameCreateBite:
		if params.RecordId == "" {
			mutation = getMutation("graphql/mutation_bite.graphql", "bite", recordData)
		} else {
			return nil, ErrUpdateBiteNotSupported
		}
	case objectNameSetUserRole:
		if params.RecordId == "" {
			mutation = getMutation("graphql/mutation_user_role.graphql", "userRole", recordData)
		} else {
			return nil, ErrUpdateRoleNotSupported
		}
	case objectNameUploadAudio:
		if params.RecordId == "" {
			mutationInput, err := ExtractAudioFields(params.RecordData)
			if err != nil {
				return nil, err
			}

			mutation = getMutation("graphql/mutation_audio.graphql", "audio", mutationInput)
		} else {
			return nil, ErrUpdateAudioSupported
		}
	case objectNameUpdateMeetingTitle:
		if params.RecordId != "" {
			input, ok := params.RecordData.(map[string]any)["input"].(map[string]any)
			if !ok {
				return nil, ErrInvalidResponseFormat
			}

			title, ok := input["title"].(string)
			if !ok {
				return nil, ErrTitleRequired
			}

			mutation = getMutation("graphql/mutation_meeting_title.graphql", "meetingTitle", map[string]string{
				"id":    params.RecordId,
				"title": title,
			})
		} else {
			return nil, ErrCreateMeetingSupported
		}
	default:
		return nil, common.ErrObjectNotSupported
	}

	requestBody := map[string]string{
		"query": mutation,
	}

	jsonBody, err := json.Marshal(requestBody)
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

func getMutation(filePath, queryName string, data any) string {
	queryBytes, err := queryFS.ReadFile(filePath)
	if err != nil {
		return ""
	}

	tmpl, err := template.New(queryName).Parse(string(queryBytes))
	if err != nil {
		return ""
	}

	var queryBuf bytes.Buffer

	err = tmpl.Execute(&queryBuf, data)
	if err != nil {
		return ""
	}

	return queryBuf.String()
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	var (
		recordID string
		err      error
	)

	node, ok := resp.Body()
	if !ok {
		return &common.WriteResult{Success: true}, nil
	}

	objectResponse, err := jsonquery.New(node).ObjectRequired("data")
	if err != nil {
		return nil, err
	}

	if params.ObjectName == "bite" {
		recordID, err = jsonquery.New(objectResponse, params.ObjectName).StrWithDefault("id", "")
		if err != nil {
			return nil, err
		}
	}

	response, err := jsonquery.Convertor.ObjectToMap(objectResponse)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     response,
	}, nil
}

// nolint
func ExtractAudioFields(RecordData any) (map[string]any, error) {
	input, ok := RecordData.(map[string]any)["input"].(map[string]any)
	if !ok {
		return nil, ErrInvalidResponseFormat
	}

	url, ok := input["url"].(string)
	if !ok {
		return nil, ErrURLIsRequired
	}

	inputParts := map[string]any{
		"url": url,
	}

	// Optional fields: only include if non-zero
	if title, ok := input["title"].(string); ok && title != "" {
		inputParts["title"] = title
	}

	if customLanguage, ok := input["custom_language"].(string); ok && customLanguage != "" {
		inputParts["custom_language"] = customLanguage
	}

	if clientReferenceId, ok := input["client_reference_id"].(string); ok && clientReferenceId != "" {
		inputParts["client_reference_id"] = clientReferenceId
	}

	if saveVideo, ok := input["save_video"].(bool); ok && saveVideo {
		inputParts["save_video"] = saveVideo
	}

	if attendees, ok := input["attendees"].([]any); ok && len(attendees) > 0 {
		inputParts["attendees"] = attendees
	}

	return inputParts, nil
}

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL)
	if err != nil {
		return nil, err
	}

	var mutation string

	switch params.ObjectName {
	case objectNamedeleteTranscript:
		if params.RecordId != "" {
			mutation = getMutation("graphql/mutation_transcript.graphql", "transcript", map[string]string{
				"transcript_id": params.RecordId,
			})
		} else {
			return nil, ErrUpdateMeetingLinkNotSupported
		}
	default:
		return nil, common.ErrObjectNotSupported
	}

	requestBody := map[string]string{
		"query": mutation,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Connector) parseDeleteResponse(
	ctx context.Context,
	params common.DeleteParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.DeleteResult, error) {
	if resp.Code != http.StatusOK {
		return nil, fmt.Errorf("%w: failed to delete record: %d", common.ErrRequestFailed, resp.Code)
	}

	// A successful delete returns 200 OK
	return &common.DeleteResult{
		Success: true,
	}, nil
}
