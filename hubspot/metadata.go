package hubspot

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
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

	// Use goroutines to fetch metadata for each object in parallel
	metadataChannel := make(chan common.ObjectMetadata, len(objectNames))
	errChannel := make(chan error, 1)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, objectName := range objectNames {
		go func(object string) {
			objectMetadata, err := c.describeObject(ctx, object)
			if err != nil {
				select {
				case errChannel <- err:
					cancel()
				case <-ctx.Done():
				}

				return
			}

			select {
			case metadataChannel <- *objectMetadata:
			case <-ctx.Done():
			}
		}(objectName)
	}

	// Collect metadata for each object
	objectsMap := make(common.ListObjectMetadataResult)

	for range objectNames {
		select {
		case objectMetadata := <-metadataChannel:
			objectsMap[objectMetadata.DisplayName] = objectMetadata
		case err := <-errChannel:
			// Drain to ensure all goroutines can exit
			for range objectNames {
				select {
				case <-metadataChannel:
				default:
				}
			}

			return nil, err
		}
	}

	return &objectsMap, nil
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
	data, err := c.get(ctx, c.BaseURL+"/properties/"+objectName)
	if err != nil {
		return nil, fmt.Errorf("error fetching HubSpot fields: %w", err)
	}

	rawResponse, err := ajson.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error marshalling composite response into byte array: %w", err)
	}

	resp := &describeObjectResponse{}

	err = json.Unmarshal(rawResponse, resp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling composite response into JSON: %w", err)
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
