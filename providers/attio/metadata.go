package attio

import (
	"context"
	"sync"
	"time"

	"github.com/amp-labs/amp-common/simultaneously"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// This struct is used for when the response data having slice of data.
type responseObject struct {
	Data []map[string]any `json:"data"`
}

// This struct is used for when the response data having single data.
type objectAttribute struct {
	Data []Data `json:"data"`
}

type Data struct {
	ID struct {
		WorkspaceID string `json:"workspace_id"` //nolint:tagliatelle
		ObjectID    string `json:"object_id"`    //nolint:tagliatelle
		AttributeID string `json:"attribute_id"` //nolint:tagliatelle
	} `json:"id"`
	Title                 string    `json:"title"`
	APISlug               string    `json:"api_slug"` //nolint:tagliatelle
	Type                  string    `json:"type"`
	IsWritable            bool      `json:"is_writable"`              //nolint:tagliatelle
	IsMultiselect         bool      `json:"is_multiselect"`           //nolint:tagliatelle
	IsDefaultValueEnabled bool      `json:"is_default_value_enabled"` //nolint:tagliatelle
	CreatedAt             time.Time `json:"created_at"`               //nolint:tagliatelle
}

// Attio allows 100 reads/s and the HTTP client doesn't retry 429s, so the fan-out is capped.
// The caps assume ~0.5s per call (5 objects x 10 option calls = 50 requests in flight);
// switch to a real rate limiter if Attio gets faster and 429s show up.
const (
	maxConcurrentObjects     = 5
	maxConcurrentOptionCalls = 10
)

// ListObjectMetadata creates metadata of object via reading objects using Attio API.
func (c *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	var mutex sync.Mutex

	metadataResult := common.NewListObjectMetadataResult()

	callbacks := make([]simultaneously.Job, 0, len(objectNames))

	for _, obj := range objectNames {
		callbacks = append(callbacks, func(ctx context.Context) error {
			metadata, err := c.getMetadata(ctx, obj)

			mutex.Lock()
			defer mutex.Unlock()

			if err != nil {
				metadataResult.Errors[obj] = err

				return nil //nolint:nilerr // intentionally collecting errors in map, not failing fast
			}

			metadataResult.Result[obj] = *metadata

			return nil
		})
	}

	// This will block until all callbacks are done.
	if err := simultaneously.DoCtx(ctx, maxConcurrentObjects, callbacks...); err != nil {
		return nil, err
	}

	return metadataResult, nil
}

func (c *Connector) getMetadata(ctx context.Context, obj string) (*common.ObjectMetadata, error) {
	metadata, isCustom, err := c.getObjectAttributes(ctx, obj)
	if err != nil {
		return nil, err
	}

	displayName := obj

	if isCustom {
		name, err := c.getObjectDisplayName(ctx, obj)
		if err != nil {
			return nil, err
		}

		if name != "" {
			displayName = name
		}
	}

	return common.NewObjectMetadata(displayName, metadata), nil
}

func (c *Connector) getObjectAttributes(
	ctx context.Context, obj string,
) (map[string]common.FieldMetadata, bool, error) {
	// Standard isn't a term we commonly use, but rather a concept defined by Attio itself.
	// supportAttioApi represents non-standarad/custom objects.
	isAttioStandardOrCustomObj := !supportAttioApi.Has(obj)

	var (
		url *urlbuilder.URL
		err error
	)

	if isAttioStandardOrCustomObj {
		url, err = c.getObjectAttributesURL(obj)
		if err != nil {
			return nil, false, err
		}
	} else {
		url, err = c.getApiURL(obj)
		if err != nil {
			return nil, false, err
		}
	}

	resp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, false, err
	}

	metadata, err := c.parseMetadataFromResponse(ctx, resp, isAttioStandardOrCustomObj)
	if err != nil {
		return nil, false, err
	}

	return metadata, isAttioStandardOrCustomObj, nil
}

// getObjectDisplayName fetches the display name for custom objects.
func (c *Connector) getObjectDisplayName(ctx context.Context, obj string) (string, error) {
	url, err := c.getObjectsURL(obj)
	if err != nil {
		return "", err
	}

	resp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return "", err
	}

	return getDisplayName(resp)
}

func (c *Connector) parseMetadataFromResponse(ctx context.Context, resp *common.JSONHTTPResponse,
	isAttioStandardOrCustomObj bool,
) (map[string]common.FieldMetadata, error) {
	// Retrieving metadata for standard and custom objects in Attio using the api_slug field.
	if isAttioStandardOrCustomObj {
		return c.parseStandardOrCustomMetadata(ctx, resp)
	}

	return c.parseMetadata(resp)
}

// Parsing the metadata response for standard or custom objects.
func (c *Connector) parseStandardOrCustomMetadata(
	ctx context.Context, resp *common.JSONHTTPResponse,
) (map[string]common.FieldMetadata, error) {
	response, err := common.UnmarshalJSON[objectAttribute](resp)
	if err != nil {
		return nil, err
	}

	if len(response.Data) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	// Attio has no bulk endpoint for select options, so each select attribute needs its own call.
	// Every callback writes only its own index, so no lock is needed.
	defaultValues := make([][]common.FieldValue, len(response.Data))
	callbacks := make([]simultaneously.Job, 0, len(response.Data))

	for idx, value := range response.Data {
		if value.Type != "select" {
			continue
		}

		callbacks = append(callbacks, func(ctx context.Context) error {
			values, err := c.getDefaultValues(ctx, value)
			defaultValues[idx] = values

			return err
		})
	}

	if err := simultaneously.DoCtx(ctx, maxConcurrentOptionCalls, callbacks...); err != nil {
		return nil, err
	}

	metadata := make(map[string]common.FieldMetadata, len(response.Data))

	for idx, value := range response.Data {
		metadata[value.APISlug] = common.FieldMetadata{
			DisplayName:  value.APISlug,
			ValueType:    getFieldValueType(value.Type, value.IsMultiselect),
			ProviderType: value.Type,
			ReadOnly:     new(!value.IsWritable),
			FieldId:      new(value.ID.AttributeID),
			Values:       defaultValues[idx],
		}
	}

	return metadata, nil
}

// Parsing the metadata response for non-standard or custom objects.
func (c *Connector) parseMetadata(resp *common.JSONHTTPResponse) (map[string]common.FieldMetadata, error) {
	metadata := make(map[string]common.FieldMetadata)

	response, err := common.UnmarshalJSON[responseObject](resp)
	if err != nil {
		return nil, err
	}

	if len(response.Data) == 0 {
		return nil, common.ErrMissingExpectedValues
	}
	// Using the first result data to generate the metadata.
	for k := range response.Data[0] {
		metadata[k] = common.FieldMetadata{
			DisplayName:  k,
			ValueType:    common.ValueTypeOther,
			ProviderType: "", // not available
			Values:       nil,
		}
	}

	return metadata, nil
}

// Getting display name for both standard and custom objects.
func getDisplayName(resp *common.JSONHTTPResponse) (string, error) {
	response, err := common.UnmarshalJSON[objectResponse](resp)
	if err != nil {
		return "", err
	}

	if response == nil {
		return "", common.ErrMissingExpectedValues
	}

	res := response.Data.PluralNoun
	if res == "" {
		return "", common.ErrNotFound
	}

	return res, nil
}

func getFieldValueType(field string, ismultiselect bool) common.ValueType {
	switch field {
	case "number":
		return common.ValueTypeInt
	case "text":
		return common.ValueTypeString
	case "select", "record-reference", "domain":
		if ismultiselect {
			return common.ValueTypeMultiSelect
		}

		return common.ValueTypeSingleSelect
	case "date":
		return common.ValueTypeDate
	case "timestamp":
		return common.ValueTypeDateTime
	default:
		// location, currency, interaction, actor-reference
		return common.ValueTypeOther
	}
}

func (c *Connector) getDefaultValues(ctx context.Context, o Data) (fields []common.FieldValue, err error) {
	url, err := c.getOptionsURL(o.ID.ObjectID, o.ID.AttributeID)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	objectAttribute, err := common.UnmarshalJSON[objectAttribute](resp)
	if err != nil {
		return nil, err
	}

	for _, title := range objectAttribute.Data {
		fields = append(fields, common.FieldValue{
			Value:        title.Title,
			DisplayValue: title.Title,
		})
	}

	return fields, nil
}
