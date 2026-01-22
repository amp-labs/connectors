package ringcentral

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

var (
	// JSON file containing object read,write details.
	//
	//go:embed objectCfg.json
	objectCfg []byte

	pathURLs = make(map[string]ObjectsOperationURLs) //nolint: gochecknoglobals
)

func init() {
	logger := logging.Logger(context.TODO())
	if err := json.Unmarshal(objectCfg, &pathURLs); err != nil {
		logger.Error("ringcentral: couldn't unmarshal object configuration json file", "err", err)
	}
}

type Response struct {
	Records              []map[string]any `json:"records"`
	ReferencedExtensions []map[string]any `json:"referencedExtensions"`
	Mappings             []map[string]any `json:"mappings"`
	Resources            []map[string]any `json:"Resources"`
	Meetings             []map[string]any `json:"meetings"`
	Recordings           []map[string]any `json:"recordings"`
	Items                []map[string]any `json:"items"`
	Tasks                []map[string]any `json:"tasks"`
	SyncInfo             map[string]any   `json:"syncInfo"`
	Navigation           map[string]any   `json:"navigation"`
	Paging               map[string]any   `json:"paging"`
}

var ErrUnexpectedRecordsField = errors.New("unexpected records field used")

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	endpointPath, exists := pathURLs[objectName]
	if !exists {
		endpointPath.ReadPath = objectName
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, endpointPath.ReadPath)
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
	objectMetadata := &common.ObjectMetadata{
		FieldsMap:   make(map[string]string),
		Fields:      make(common.FieldsMetadata),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	var recordField string

	objInfo, exist := pathURLs[objectName]
	if !exist {
		recordField = records
	} else {
		recordField = objInfo.RecordsField
	}

	data, err := common.UnmarshalJSON[Response](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	records, err := GetFieldByJSONTag(data, recordField)
	if err != nil {
		return nil, fmt.Errorf("couldn't fetch records for object: %s expecting record field %s: %w",
			objectName, recordField, err)
	}

	if len(records) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	firstRecord := records[0]

	for fld, val := range firstRecord {
		objectMetadata.Fields[fld] = common.FieldMetadata{
			DisplayName: fld,
			ValueType:   inferValue(val),
		}
	}

	return objectMetadata, nil
}

func (c *Connector) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if params.NextPage != "" {
		return urlbuilder.New(params.NextPage.String())
	}

	objectPaths, exists := pathURLs[params.ObjectName]
	if !exists {
		return urlbuilder.New(c.ProviderInfo().BaseURL, "restapi/v1.0", params.ObjectName)
	}

	return urlbuilder.New(c.ProviderInfo().BaseURL, objectPaths.ReadPath)
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	if !params.Since.IsZero() {
		if creationTimeFrom.Has(params.ObjectName) {
			url.WithQueryParam("creationTimeFrom", params.Since.Format(time.RFC3339))
		}
	}

	if !params.Until.IsZero() {
		if creationTimeFrom.Has(params.ObjectName) {
			url.WithQueryParam("creationTimeTo", params.Until.Format(time.RFC3339))
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	var recordsField string

	url, err := urlbuilder.New(request.URL.String())
	if err != nil {
		return nil, err
	}

	objectPaths, exists := pathURLs[params.ObjectName]
	if !exists {
		recordsField = records
	} else {
		recordsField = objectPaths.RecordsField
	}

	return common.ParseResult(
		response,
		common.ExtractRecordsFromPath(recordsField),
		nextRecordsURL(params.ObjectName, url),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) buildWriteURL(params common.WriteParams) (*urlbuilder.URL, string, error) {
	method := http.MethodPost

	objectPaths, exists := pathURLs[params.ObjectName]
	if !exists {
		url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "restapi/v1.0", params.ObjectName)
		if err != nil {
			return nil, "", err
		}

		return url, method, nil
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, objectPaths.WritePath)
	if err != nil {
		return nil, "", err
	}

	if params.RecordId != "" {
		url.AddPath(params.RecordId)

		if objectPaths.UpdateMethod != "" {
			method = objectPaths.UpdateMethod
		} else {
			method = http.MethodPut
		}
	}

	return url, method, nil
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	url, method, err := c.buildWriteURL(params)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	return req, nil
}

func retrieveRecordId(body *ajson.Node) (string, error) {
	var idVal string

	// 1. we try integer
	recordID, err := jsonquery.New(body).IntegerWithDefault("id", 0)
	if !errors.Is(err, jsonquery.ErrNotNumeric) {
		return "", err
	}

	idVal = strconv.Itoa(int(recordID))

	// 2. we try string
	if recordID == 0 {
		recordId, err := jsonquery.New(body).StrWithDefault("id", "")
		if err != nil {
			return "nil", err
		}

		idVal = recordId
	}

	return idVal, nil
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	recordID, err := retrieveRecordId(body)
	if err != nil {
		return nil, err
	}

	resp, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     resp,
	}, nil
}
