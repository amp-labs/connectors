package drift

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const (
	users         = "users"
	conversations = "conversations"
	playbooks     = "playbooks"
	accounts      = "accounts"

	ListSuffix = "/list"
)

var ErrUnexpectedFieldId = errors.New("received unexpected data type in recordId field") //nolint: gochecknoglobals

// Create a set of endpoints that require the list suffix.
var endpointsRequiringListSuffix = datautils.NewSet(users, conversations, playbooks) //nolint: gochecknoglobals

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.constructReadURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	if params.NextPage != "" {
		url, err = urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
		}
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		records(params.ObjectName),
		nextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}

func (c *Connector) constructReadURL(objectName string) (*urlbuilder.URL, error) {
	lowerCaseObject := strings.ToLower(objectName)

	// Check if this endpoint requires the list suffix
	if endpointsRequiringListSuffix.Has(lowerCaseObject) {
		objectName += ListSuffix
	}

	return urlbuilder.New(c.ProviderInfo().BaseURL, objectName)
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	method := http.MethodPost

	url, err := c.constructWriteURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		method = http.MethodPatch

		url = constructUpdateEndpoint(url, params.ObjectName, params.RecordId)
	} else {
		url = constructCreateEndpoint(url, params.ObjectName)
	}

	if params.ObjectName == updateAccount {
		method = http.MethodPatch
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(jsonData))
}

func (c *Connector) constructWriteURL(objectName string) (*urlbuilder.URL, error) {
	return urlbuilder.New(c.ProviderInfo().BaseURL, objectName)
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	logging.With(ctx, "drift connector")

	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	resp, err := jsonquery.New(body).ObjectRequired(writeResponseField(params.ObjectName))
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(resp)
	if err != nil {
		return nil, err
	}

	recordId, err := retrieveRecordId(params.ObjectName, data)
	if err != nil {
		logging.Logger(ctx).Error("failed to retrieve the recordId from response data",
			"object", params.ObjectName, "response", data)
	}

	return &common.WriteResult{
		Success:  true,
		Data:     data,
		RecordId: recordId,
	}, nil
}

func retrieveRecordId(objectName string, data map[string]any) (string, error) {
	if !recordIdFields.Has(objectName) {
		return "", nil
	}

	fld := recordIdFields.Get(objectName)

	id, ok := data[fld].(float64)
	if !ok {
		return "", ErrUnexpectedFieldId
	}

	recordId := strconv.Itoa(int(id))

	return recordId, nil
}
