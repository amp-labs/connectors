package hubspot

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/spyzhov/ajson"
)

type objectMetadataResult struct {
	ObjectName string
	Response   common.ObjectMetadata
}

type objectMetadataError struct {
	ObjectName string
	Error      error
}

// ListObjectMetadata returns object metadata for each object name provided.
func (c *Connector) ListObjectMetadata( // nolint:cyclop,funlen
	ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	ctx = logging.With(ctx, "connector", "hubspot")

	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	// Use goroutines to fetch metadata for each object in parallel
	metadataChannel := make(chan *objectMetadataResult, len(objectNames))
	errChannel := make(chan *objectMetadataError, len(objectNames))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, objectName := range objectNames {
		go func(object string) {
			objectMetadata, err := c.describeObject(ctx, object)
			if err != nil {
				errChannel <- &objectMetadataError{
					ObjectName: object,
					Error:      err,
				}

				return
			}

			// Send object metadata to metadataChannel
			metadataChannel <- &objectMetadataResult{
				ObjectName: object,
				Response:   *objectMetadata,
			}
		}(objectName)
	}

	// Collect metadata for each object
	objectsMap := &common.ListObjectMetadataResult{}
	objectsMap.Result = make(map[string]common.ObjectMetadata)
	objectsMap.Errors = make(map[string]error)

	for range objectNames {
		select {
		// Add object metadata to objectsMap
		case objectMetadataResult := <-metadataChannel:
			objectsMap.Result[objectMetadataResult.ObjectName] = objectMetadataResult.Response
		case objectMetadataError := <-errChannel:
			objectsMap.Errors[objectMetadataError.ObjectName] = objectMetadataError.Error
		}
	}

	return objectsMap, nil
}

type describeObjectResponse struct {
	Results []describeObjectResult `json:"results"`
}

type describeObjectResult struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

// describeObject returns object metadata for the given object name.
func (c *Connector) describeObject(ctx context.Context, objectName string) (*common.ObjectMetadata, error) {
	relativeURL := strings.Join([]string{"properties", objectName}, "/")

	u, err := c.getURL(relativeURL)
	if err != nil {
		return nil, err
	}

	rsp, err := c.Client.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("error fetching HubSpot fields: %w", err)
	}

	body, ok := rsp.Body()
	if !ok {
		return nil, fmt.Errorf("cannot get HubSpot fields %w", common.ErrEmptyJSONHTTPResponse)
	}

	rawResponse, err := ajson.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("error marshalling object metadata response into byte array: %w", err)
	}

	resp := &describeObjectResponse{}

	err = json.Unmarshal(rawResponse, resp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling object metadata response into JSON: %w", err)
	}

	return &common.ObjectMetadata{
		DisplayName: objectName,
		FieldsMap:   makeFieldsMap(resp),
	}, nil
}

// makeFieldsMap returns a map of field name to field label.
func makeFieldsMap(data *describeObjectResponse) map[string]string {
	fieldsMap := make(map[string]string)

	for _, field := range data.Results {
		fieldName := strings.ToLower(field.Name)

		// Add entry to fieldsMap
		fieldsMap[fieldName] = field.Label
	}

	return fieldsMap
}

func (c *Connector) GetPostAuthInfo(
	ctx context.Context,
) (*common.PostAuthInfo, error) {
	accInfo, resp, err := c.GetAccountInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("error fetching HubSpot account info: %w", err)
	}

	return &common.PostAuthInfo{
		ProviderWorkspaceRef: strconv.Itoa(accInfo.PortalId),
		RawResponse:          resp,
	}, nil
}

type AccountInfo struct {
	PortalId              int    `json:"portalId"`
	TimeZone              string `json:"timeZone"`
	CompanyCurrency       string `json:"companyCurrency"`
	AdditionalCurrencies  []string
	UTCOffset             string `json:"utcOffset"`
	UTCOffsetMilliseconds int    `json:"utcOffsetMilliseconds"`
	UIDomain              string `json:"uiDomain"`
	DataHostingLocation   string `json:"dataHostingLocation"`
}

func (c *Connector) GetAccountInfo(ctx context.Context) (*AccountInfo, *common.JSONHTTPResponse, error) {
	ctx = logging.With(ctx, "connector", "hubspot")

	resp, err := c.Client.Get(ctx, "account-info/v3/details")
	if err != nil {
		return nil, resp, fmt.Errorf("error fetching HubSpot token info: %w", err)
	}

	accountInfo, err := common.UnmarshalJSON[AccountInfo](resp)
	if err != nil {
		return nil, resp, fmt.Errorf("error unmarshalling account info response into JSON: %w", err)
	}

	return accountInfo, resp, nil
}
