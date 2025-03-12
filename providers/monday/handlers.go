package monday

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objectName)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
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

	data, err := common.UnmarshalJSON[map[string]any](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	rawRecords, exists := (*data)[objectName]
	if !exists {
		return nil, fmt.Errorf("missing expected values for object: %s, error: %w", objectName, common.ErrMissingExpectedValues) //nolint:lll
	}

	records, ok := rawRecords.([]any)
	if len(records) == 0 || !ok {
		return nil, fmt.Errorf("unexpected type or empty records for object: %s, error: %w", objectName, common.ErrMissingExpectedValues) //nolint:lll
	}

	firstRecord, ok := records[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected record format for object: %s, error: %w", objectName, common.ErrMissingExpectedValues) //nolint:lll
	}

	for fld := range firstRecord {
		objectMetadata.FieldsMap[fld] = fld
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		resp,
		common.GetOptionalRecordsUnderJSONPath(""),
		getNextRecordsURL,
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	// TODO: Implement write request builder
	return nil, common.ErrNotImplemented
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	// Get the JSON node from response
	node, ok := resp.Body()
	if !ok {
		// Handle empty response
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	// Get the appropriate ID field path for this object
	idPath, valid := RecordIDPaths[params.ObjectName]
	if !valid {
		return nil, fmt.Errorf("%w: %s", common.ErrOperationNotSupportedForObject, params.ObjectName)
	}

	// ID is integer that is always stored under different field name.
	rawID, err := jsonquery.New(node).IntegerOptional(idPath)
	if err != nil {
		return nil, err
	}

	recordID := ""
	if rawID != nil {
		// optional
		recordID = strconv.FormatInt(*rawID, 10)
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     nil,
	}, nil
}

func (c *Connector) buildDeleteRequest(ctx context.Context, params common.DeleteParams) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, params.ObjectName)
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
