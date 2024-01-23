package salesforce

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/amp-labs/connectors/common"
)

// ListObjectMetadata returns object metadata for each object name provided.
func (c *Connector) ListObjectMetadata(
	ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	// Ensure that objectNames is not empty
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	requests := make([]compositeRequestItem, len(objectNames))

	// Construct describe requests for each object name
	for idx, objectName := range objectNames {
		describeObjectURL, err := url.JoinPath(
			fmt.Sprintf("/services/data/%s/sobjects/%s/describe", APIVersion(), objectName),
		)
		if err != nil {
			return nil, err
		}

		requests[idx] = compositeRequestItem{
			Method:      "GET",
			URL:         describeObjectURL,
			ReferenceId: objectName,
		}
	}

	// Construct endpoint for the request
	compositeRequestEndpoint, err := url.JoinPath(c.BaseURL, "composite")
	if err != nil {
		return nil, err
	}

	// Make the request
	result, err := c.post(
		ctx,
		compositeRequestEndpoint,
		compositeRequest{
			CompositeRequest: requests,
			// If we fail to fetch metadata for one object, we don't want to fail the entire request.
			AllOrNone: false,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error fetching Salesforce fields: %w", err)
	}

	// Construct map of object names to object metadata
	return constructResponseMap(result)
}

// constructResponseMap constructs a map of object names to object metadata from the composite response.
func constructResponseMap(result *common.JSONHTTPResponse) (*common.ListObjectMetadataResult, error) {
	objectsMap := &common.ListObjectMetadataResult{}
	objectsMap.Result = make(map[string]common.ObjectMetadata)
	objectsMap.Errors = make(map[string]error)

	resp, err := common.UnmarshalJSON[compositeResponse](result)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response from JSON: %w", err)
	}

	// Construct map of object names to object metadata
	for _, subRes := range resp.CompositeResponse {
		result := &describeSObjectResult{}

		err = json.Unmarshal(subRes.Body, result)
		if err != nil {
			// If one of the sub-requests of the composite request fails, then subRes.Body will look like:
			// "[{\"errorCode\":\"NOT_FOUND\",\"message\":\"The requested resource does not exist\"}]"
			// which will fail the json.Unmarshall
			objectsMap.Errors[strings.ToLower(subRes.ReferenceId)] = fmt.Errorf(
				"%w: %s", ErrCannotReadMetadata, string(subRes.Body),
			)
		} else {
			objectsMap.Result[strings.ToLower(result.Name)] = common.ObjectMetadata{
				DisplayName: result.Label,
				// Map that satisfies type constraint
				FieldsMap: makeFieldsMap(result.Fields),
			}
		}
	}

	return objectsMap, nil
}

// makeFieldsMap constructs a map of field names to field labels from a describeSObjectResult.
func makeFieldsMap(fields []fieldResult) map[string]string {
	fieldsMap := make(map[string]string)

	for _, field := range fields {
		fieldName := strings.ToLower(field.Name)

		// Add entry to fieldsMap
		fieldsMap[fieldName] = field.Label
	}

	return fieldsMap
}

type compositeRequest struct {
	AllOrNone        bool                   `json:"allOrNone"`
	CompositeRequest []compositeRequestItem `json:"compositeRequest"`
}

type compositeResponse struct {
	CompositeResponse []compositeResponseItem `json:"compositeResponse"`
}

type compositeRequestItem struct {
	// ReferenceId allows us to map the result to the original request
	ReferenceId string `json:"referenceId"`
	Method      string `json:"method"`
	URL         string `json:"url"`
	Body        any    `json:"body,omitempty"`
}

type compositeResponseItem struct {
	// ReferenceId comes from the original request
	ReferenceId    string            `json:"referenceId"`
	Body           json.RawMessage   `json:"body"`
	HttpHeaders    map[string]string `json:"httpHeaders"`    //nolint:revive
	HttpStatusCode int               `json:"httpStatusCode"` //nolint:revive
}

// See https://developer.salesforce.com/docs/atlas.en-us.244.0.api.meta/api/sforce_api_calls_describesobjects_describesobjectresult.htm.
// NOTE: doc page is for SOAP API, but REST API returns the same result.
//
//nolint:lll
type describeSObjectResult struct {
	Name   string        `json:"name"`
	Label  string        `json:"label"`
	Fields []fieldResult `json:"fields" validate:"required"`
}

// See https://developer.salesforce.com/docs/atlas.en-us.244.0.api.meta/api/sforce_api_calls_describesobjects_describesobjectresult.htm#field.
//
//nolint:lll
type fieldResult struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}
