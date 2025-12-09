package kaseyavsax

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	limitQuery       = "$top"
	filterQuery      = "$filter"
	metadataPageSize = "1"
	metaField        = "Meta"
	dataField        = "Data"
	NextRecordsField = "NextQueryLink"
	pageSize         = 100
)

type readResponse struct {
	Data []map[string]any `json:"Data"`
	Meta map[string]any   `json:"Meta"`
}

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion, objectName)
	if err != nil {
		return nil, err
	}

	// Limit response to 1 record data.
	url.WithQueryParam(limitQuery, metadataPageSize)

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
		Fields:      make(common.FieldsMetadata),
		DisplayName: naming.CapitalizeFirstLetterEveryWord(objectName),
	}

	resp, err := common.UnmarshalJSON[readResponse](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("%w: could not find a record to sample fields from", common.ErrMissingExpectedValues)
	}

	firstRecord := resp.Data[0]

	for field, value := range firstRecord {
		objectMetadata.Fields[field] = common.FieldMetadata{
			DisplayName:  field,
			ValueType:    inferValueTypeFromData(value),
			ProviderType: "", // not available
			Values:       nil,
		}
	}

	return &objectMetadata, nil
}

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	url, err = urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion, params.ObjectName)
	if err != nil {
		return nil, err
	}

	if !params.Since.IsZero() && supportsFiltering(params.ObjectName) {
		f := fmt.Sprintf("UpdatedAt ge %v", params.Since.Format(time.RFC3339))
		url.WithQueryParam(filterQuery, f)
	}

	url.WithQueryParam(limitQuery, strconv.Itoa(pageSize))

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
		records(),
		nextRecordsURL(),
		common.GetMarshaledData,
		params.Fields,
	)
}
