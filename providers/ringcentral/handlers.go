package ringcentral

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
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
