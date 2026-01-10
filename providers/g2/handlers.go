package g2

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	limitQuery       = "page[size]"
	metadataPageSize = "1"
)

type Response struct {
	Data  []record       `json:"data"`
	Links map[string]any `json:"links"`
}

type record struct {
	Id            string         `json:"id"`
	Type          string         `json:"type"`
	Attributes    map[string]any `json:"attributes"`
	Relationships map[string]any `json:"relationships"`
}

var restAPIVersion string = "api/v2" //nolint: gochecknoglobals

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	path, err := PathsConfig(c.productId, objectName)
	if err != nil {
		return nil, err
	}

	// We couldn't test on the product's buyer_intent API
	// we used this sandbox object and since it's a replica of the product it should work fine.
	if objectName == PathSandboxBuyerIntent {
		restAPIVersion = "api/sandbox"
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, restAPIVersion, path)
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

	resp, err := common.UnmarshalJSON[Response](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("%w: could not find a record to sample fields from", common.ErrMissingExpectedValues)
	}

	// Add attributes fields to metadata
	firstRecord := resp.Data[0].Attributes
	for fld, val := range firstRecord {
		objectMetadata.Fields[fld] = common.FieldMetadata{
			DisplayName: fld,
			ValueType:   inferValueTypeFromData(val),
		}
	}

	// Add the id of the data layer into fields.
	objectMetadata.Fields["id"] = common.FieldMetadata{
		DisplayName: "Id",
		ValueType:   common.ValueTypeString,
	}

	return &objectMetadata, nil
}
