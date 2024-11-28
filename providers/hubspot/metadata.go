package hubspot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
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

	rsp, err := c.Client.Get(ctx, c.getURL(relativeURL))
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

var (
	errMissingAccessToken   = errors.New("missing access token")
	errFailedToGetTokenInfo = errors.New("failed to get token info")
	errFailedToGetHubId     = errors.New("failed to get hub id")
)

func (c *Connector) GetPostAuthInfo(
	ctx context.Context,
	params *common.PostAuthInfoParams,
) (*common.PostAuthInfo, error) {
	if params.AccessToken == "" {
		return nil, errMissingAccessToken
	}

	resp, err := c.Client.Get(ctx, "/oauth/v1/access-tokens/"+params.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("error fetching HubSpot token info: %w", err)
	}

	body, ok := resp.Body()
	if !ok {
		return nil, errors.Join(errFailedToGetTokenInfo, common.ErrEmptyJSONHTTPResponse)
	}

	hubspotId, err := body.GetKey("hub_id")
	if err != nil {
		return nil, errors.Join(errFailedToGetHubId, err)
	}

	hubId, err := hubspotId.GetString()
	if err != nil {
		return nil, fmt.Errorf("error parsing 'hub_id': %w", err)
	}

	return &common.PostAuthInfo{
		ProviderWorkspaceRef: hubId,
		RawResponse:          resp,
	}, nil
}
