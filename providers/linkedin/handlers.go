package linkedin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const LinkedInVersion = "202504"

type responseObject struct {
	Elements []map[string]any `json:"elements"`
}

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "rest", objectName)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("LinkedIn-Version", LinkedInVersion) // nolint:canonicalheader
	req.Header.Add("X-Restli-Protocol-Version", "2.0.0")

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
		DisplayName: naming.CapitalizeFirstLetter(objectName),
	}

	data, err := common.UnmarshalJSON[responseObject](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if len(data.Elements) == 0 {
		return nil, ErrMetadataNotFound
	}

	for field := range data.Elements[0] {
		objectMetadata.Fields[field] = common.FieldMetadata{
			DisplayName:  field,
			ValueType:    common.ValueTypeOther,
			ProviderType: "",
			ReadOnly:     false,
			Values:       nil,
		}
	}

	return &objectMetadata, nil
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, "rest", params.ObjectName)
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		url.AddPath(params.RecordId)
	}

	jsonData, err := json.Marshal(params.RecordData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url.String(), bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	if params.RecordId != "" {
		req.Header.Add("X-Restli-Method", "PARTIAL_UPDATE")
	}

	req.Header.Add("LinkedIn-Version", LinkedInVersion) // nolint:canonicalheader
	req.Header.Add("X-Restli-Protocol-Version", "2.0.0")

	return req, nil
}

func (c *Connector) parseWriteResponse(
	ctx context.Context,
	params common.WriteParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		return &common.WriteResult{ // nolint:nilerr
			Success: true,
		}, nil
	}

	RecordId := strings.TrimPrefix(request.Header.Get("Location"), params.ObjectName+"/")

	resp, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: RecordId,
		Errors:   nil,
		Data:     resp,
	}, nil
}
