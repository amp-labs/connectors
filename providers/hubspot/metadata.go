package hubspot

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/simultaneously"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/metadata"
)

type objectMetadataResult struct {
	ObjectName string
	Response   common.ObjectMetadata
}

type objectMetadataError struct {
	ObjectName string
	Error      error
}

func (c *Connector) UpsertMetadata(
	ctx context.Context, params *common.UpsertMetadataParams,
) (*common.UpsertMetadataResult, error) {
	// Delegated.
	return c.crmAdapter.UpsertMetadata(ctx, params)
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

	callbacks := make([]simultaneously.Job, 0, len(objectNames))

	for _, objectName := range objectNames {
		obj := objectName // capture loop variable

		callbacks = append(callbacks, func(ctx context.Context) error {
			objectMetadata, err := c.getObjectMetadata(ctx, obj)
			if err != nil {
				errChannel <- &objectMetadataError{
					ObjectName: obj,
					Error:      err,
				}

				return nil //nolint:nilerr // intentionally collecting errors in channel, not failing fast
			}

			// Send object metadata to metadataChannel
			metadataChannel <- &objectMetadataResult{
				ObjectName: obj,
				Response:   *objectMetadata,
			}

			return nil
		})
	}

	// This will block until all callbacks are done. Note that since the
	// channels are buffered, the above code won't block on sending to them
	// even if we're not receiving yet.
	if err := simultaneously.DoCtx(ctx, -1, callbacks...); err != nil {
		close(metadataChannel)
		close(errChannel)

		return nil, err
	}

	// Since all callbacks are done, we can close the channels.
	// This ensures that the following range loops will terminate.
	close(metadataChannel)
	close(errChannel)

	// Collect metadata for each object
	objectsMap := &common.ListObjectMetadataResult{}
	objectsMap.Result = make(map[string]common.ObjectMetadata)
	objectsMap.Errors = make(map[string]error)

	for object := range metadataChannel {
		objectsMap.Result[object.ObjectName] = object.Response
	}

	for object := range errChannel {
		objectsMap.Errors[object.ObjectName] = object.Error
	}

	return objectsMap, nil
}

// getObjectMetadata returns object metadata for the given object name.
func (c *Connector) getObjectMetadata(ctx context.Context, objectName string) (*common.ObjectMetadata, error) {
	if crmObjectsWithoutPropertiesAPISupport.Has(objectName) {
		return c.getObjectMetadataFromCRMSearch(ctx, objectName)
	}

	return c.getObjectMetadataFromPropertyAPI(ctx, objectName)
}

// This method describes objects that are part of Objects API using properties endpoint.
// There is a dedicated API endpoint that is used for discovery of object properties.
func (c *Connector) getObjectMetadataFromPropertyAPI(
	ctx context.Context, objectName string,
) (*common.ObjectMetadata, error) {
	u, err := c.getPropertiesURL(objectName)
	if err != nil {
		return nil, err
	}

	rsp, err := c.Client.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("error fetching HubSpot fields: %w", err)
	}

	resp, err := common.UnmarshalJSON[fieldDescriptionResponse](rsp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling object metadata response into JSON: %w", err)
	}

	// Attached enum value options to each field if any.
	fields, err := c.fetchExternalMetadataEnumValues(ctx, objectName, resp.transformToFields())
	if err != nil {
		return nil, err
	}

	// Mark required fields.
	fields, err = c.fetchRequiredFieldsBestEffort(ctx, objectName, fields)
	if err != nil {
		return nil, err
	}

	return common.NewObjectMetadata(
		objectName, fields,
	), nil
}

// Method focuses on acquiring object properties for those objects that are not part of CRM Properties API.
// https://developers.hubspot.com/docs/guides/api/crm/objects/companies
// There is no discovery endpoint to acquire object properties, therefore, manual read is used.
func (c *Connector) getObjectMetadataFromCRMSearch(
	ctx context.Context, objectName string,
) (*common.ObjectMetadata, error) {
	readResult, err := c.searchCRM(ctx, searchCRMParams{
		SearchParams: SearchParams{
			ObjectName: objectName,
			Fields:     connectors.Fields(""), // passed to satisfy validation
			NextPage:   "",
		},
		PageSize: 1,
	})
	if err != nil {
		// Ignore an error and fallback to static schema.
		return metadata.Schemas.SelectOne(c.moduleID, objectName)
	}

	if len(readResult.Data) == 0 {
		// Read returned no rows.
		return metadata.Schemas.SelectOne(c.moduleID, objectName)
	}

	fields := make(map[string]common.FieldMetadata)
	for fieldName := range readResult.Data[0].Raw {
		fields[fieldName] = common.FieldMetadata{
			DisplayName:  fieldName,
			ValueType:    common.ValueTypeOther,
			ProviderType: "", // not available
			Values:       nil,
		}
	}

	return common.NewObjectMetadata(objectName, fields), nil
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

type fieldDescriptionResponse struct {
	Results []fieldDescription `json:"results"`
}

type fieldDescription struct {
	Name      string `json:"name"`
	Label     string `json:"label"`
	Type      string `json:"type"`
	FieldType string `json:"fieldType"`
	// IsBuiltIn indicates whether the field is HubSpot-defined (built-in).
	// If false or omitted, the field is custom.
	IsBuiltIn            bool                      `json:"hubspotDefined"`
	Options              []fieldEnumerationOption  `json:"options"`
	ModificationMetadata fieldModificationMetadata `json:"modificationMetadata"`
}

type fieldEnumerationOption struct {
	Label       string `json:"label"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

type fieldModificationMetadata struct {
	ReadOnlyValue bool `json:"readOnlyValue"`
}

func (r fieldDescriptionResponse) transformToFields() map[string]common.FieldMetadata {
	fieldsMap := make(map[string]common.FieldMetadata)

	for _, field := range r.Results {
		fieldName := strings.ToLower(field.Name)
		fieldsMap[fieldName] = field.transformToFieldMetadata()
	}

	return fieldsMap
}

// transformToFieldMetadata converts Provider model of a field into Ampersand's common.FieldMetadata.
// This normalizes provider response to the unified standard across all providers.
func (f fieldDescription) transformToFieldMetadata() common.FieldMetadata {
	var (
		valueType common.ValueType
		values    []common.FieldValue
	)

	// Based on type and field type properties from Hubspot object model map value to Ampersand value type.
	switch f.Type {
	case "string":
		valueType = common.ValueTypeString
	case "number":
		valueType = common.ValueTypeFloat
	case "bool":
		valueType = common.ValueTypeBoolean
	case "datetime":
		valueType = common.ValueTypeDateTime
	case "enumeration":
		valueType, values = f.implyEnumerationType(f.Name)
		// Enumeration type means there are predefined field values.
	default:
		// ex: object_coordinates, phone_number
		valueType = common.ValueTypeOther
	}

	return common.FieldMetadata{
		DisplayName:  f.Label,
		ValueType:    valueType,
		ProviderType: f.Type + "." + f.FieldType,
		ReadOnly:     goutils.Pointer(f.ModificationMetadata.ReadOnlyValue),
		IsCustom:     goutils.Pointer(!f.IsBuiltIn),
		// IsRequired is not known from current struct,
		// info is acquired by different API call and set by fetchRequiredFieldsBestEffort.
		IsRequired: nil,
		Values:     values,
	}
}

func (f fieldDescription) implyEnumerationType(fieldName string) (common.ValueType, []common.FieldValue) {
	var values []common.FieldValue

	if len(f.Options) != 0 {
		// List of values is not nil, at least one option exists.
		values = make([]common.FieldValue, len(f.Options))

		for index, option := range f.Options {
			displayValue := option.Label
			// For persona field, use description if it exists, otherwise fall back to label
			// https://community.hubspot.com/t5/APIs-Integrations/Getting-Wrong-Value-from-Persona-in-API/
			// m-p/1193587/highlight/true#M84004
			if strings.EqualFold(fieldName, "hs_persona") && option.Description != "" {
				displayValue = option.Description
			}

			values[index] = common.FieldValue{
				Value:        option.Value,
				DisplayValue: displayValue,
			}
		}
	}

	switch f.FieldType {
	case "checkbox":
		return common.ValueTypeMultiSelect, values
	case "booleancheckbox":
		// Boolean values are ignored.
		return common.ValueTypeBoolean, nil
	case "radio":
		return common.ValueTypeSingleSelect, values
	case "select":
		return common.ValueTypeSingleSelect, values
	default:
		// ex: enumeration.calculation_equation
		return common.ValueTypeOther, values
	}
}

// Registry of objects to the fields with external metadata.
// It provides information on location of data (URLs) and how to process JSON to infer enum options.
// If you want to support more objects and their fields extend this registry.
//
// NOTE: ResponseProcessor may retrieve values for multiple fields in a single API call.
var objectsWithExternalMetadataFields = datautils.Map[string, []externalFieldDiscovery]{ // nolint:gochecknoglobals
	"contacts": {
		{
			FieldNames:        []string{"hs_pipeline"},
			EndpointPath:      "/crm/v3/pipelines/contacts",
			ResponseProcessor: parsePipelineFieldValues,
		},
	},
	"deals": {
		{
			FieldNames:        []string{"pipeline", "dealstage"},
			EndpointPath:      "/crm/v3/pipelines/deals",
			ResponseProcessor: parsePipelineFieldValuesWithStages,
		},
	},
}

// Hubspot may have common.ValueTypeSingleSelect or common.ValueTypeMultiSelect without values.
// This means we have to make additional API calls to resolve missing values.
// Current procedure doesn't resolve all fields, where fieldDescription.externalOptions == true.
func (c *Connector) fetchExternalMetadataEnumValues(
	ctx context.Context,
	objectName string, fields map[string]common.FieldMetadata,
) (map[string]common.FieldMetadata, error) {
	externalFields, ok := objectsWithExternalMetadataFields[objectName]
	if !ok {
		// Nothing to retrieve. This object doesn't have or doesn't support external field discovery.
		return fields, nil
	}

	// For each external field that we support make an API call to fetch enumeration options.
	// Store this values for each field within each object.
	for _, discovery := range externalFields {
		rsp, err := c.Client.Get(ctx, c.providerInfo.BaseURL+discovery.EndpointPath)
		if err != nil {
			return nil, fmt.Errorf("error resolving external metadata values for HubSpot: %w", err)
		}

		perFieldValues, err := discovery.ResponseProcessor(rsp)
		if err != nil {
			return nil, err
		}

		// Response processor returns an array for each field in that order.
		if len(perFieldValues) != len(discovery.FieldNames) {
			return nil, common.ErrInvalidImplementation
		}

		// Store field values associated with each field.
		for index, discoveryFieldName := range discovery.FieldNames {
			fieldMetadata, ok := fields[discoveryFieldName]
			if !ok {
				// Provider no longer has this field.
				continue
			}

			// Store discovered values.
			fieldMetadata.Values = perFieldValues[index]
			fields[discoveryFieldName] = fieldMetadata
		}
	}

	return fields, nil
}

// Represents metadata discovery details for an object's fields.
type externalFieldDiscovery struct {
	// FieldNames lists the fields that have enum options available at EndpointPath.
	// A single API response may provide values for multiple fields.
	FieldNames []string
	// EndpointPath specifies the API location where the list of enum values can be retrieved.
	EndpointPath string
	// ResponseProcessor extracts and parses []common.FieldValue for each field.
	ResponseProcessor externalFieldProcessor
}

// Extracts values for multiple fields from the response.
// The number of returned lists must match externalFieldDiscovery.FieldNames, maintaining the same order.
type externalFieldProcessor func(response *common.JSONHTTPResponse) ([]common.FieldValues, error)

// Parses field values from an API response, producing two value arrays:
//  1. The first array contains pipeline values.
//  2. The second array contains aggregated stage values,
//     where each stage ID is prefixed with the ID of its corresponding pipeline.
func parsePipelineFieldValuesWithStages(response *common.JSONHTTPResponse) ([]common.FieldValues, error) {
	resp, err := common.UnmarshalJSON[pipelineResponse](response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling pipelines into JSON: %w", err)
	}

	result := make([]common.FieldValues, 2) // nolint:mnd

	// Aggregate stages across all pipelines.
	// Note: Multiple pipelines can exist for each object type, with each instance referencing a specific pipeline.
	//
	// Problem: In HubSpot, field options are not defined at the object level but at the instance level.
	// This means the object identifier is required to locate the pipeline and subsequently retrieve its options.
	pipelines := make(common.FieldValues, 0)
	stages := make(common.FieldValues, 0)

	for _, pipelineItem := range resp.Pipelines {
		pipelines = append(pipelines, common.FieldValue{
			Value:        pipelineItem.ID,
			DisplayValue: pipelineItem.DisplayName,
		})

		for _, s := range pipelineItem.Stages {
			stages = append(stages, common.FieldValue{
				Value:        fmt.Sprintf("%v:%v", pipelineItem.ID, s.Value),
				DisplayValue: s.DisplayName,
			})
		}
	}

	result[0] = pipelines
	result[1] = stages

	return result, nil
}

// Parses and returns only pipeline values.
// Stage values are not included, because there is no corresponding field on the object to store them.
func parsePipelineFieldValues(response *common.JSONHTTPResponse) ([]common.FieldValues, error) {
	listOfValues, err := parsePipelineFieldValuesWithStages(response)
	if err != nil {
		return nil, err
	}

	if len(listOfValues) != 2 { // nolint:mnd
		return nil, common.ErrInvalidImplementation
	}

	return []common.FieldValues{
		listOfValues[0],
	}, nil
}

// For more details, refer to the HubSpot documentation on Pipelines.
// https://developers.hubspot.com/docs/guides/api/crm/pipelines#retrieve-pipelines
type pipelineResponse struct {
	Pipelines []pipeline `json:"results"`
}

type pipeline struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"label"`
	Stages      []stage `json:"stages"`
}

type stage struct {
	Value       string `json:"id"`
	DisplayName string `json:"label"`
}

// fetchRequiredFieldsBestEffort fetches the object's schema and marks required fields.
// If the schema cannot be fetched due to the `crm.schemas.custom.read` scope missing, it returns the original
// fields unchanged without an error. Other errors are returned normally.
func (c *Connector) fetchRequiredFieldsBestEffort(
	ctx context.Context, objectName string, fields map[string]common.FieldMetadata,
) (map[string]common.FieldMetadata, error) {
	url, err := c.getObjectSchemaURL(objectName)
	if err != nil {
		return nil, err
	}

	rsp, err := c.Client.Get(ctx, url.String())
	if err != nil {
		if isMissingSchemasScope(err) {
			// User does not have permission to access the schema endpoint.
			// Return the original fields without enrichment.
			logging.VerboseLogger(ctx).Debug(fmt.Sprintf(
				"Not populating isRequired for fields of %s because scopes are missing", objectName,
			))

			return fields, nil
		}

		return nil, fmt.Errorf("error fetching HubSpot fields: %w", err)
	}

	resp, err := common.UnmarshalJSON[schemaResponse](rsp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling schemaResponse response into JSON: %w", err)
	}

	required := datautils.NewSetFromList(resp.RequiredProperties)

	for name, meta := range fields {
		isRequired := required.Has(name)
		meta.IsRequired = goutils.Pointer(isRequired)
		fields[name] = meta
	}

	return fields, nil
}

func isMissingSchemasScope(err error) bool {
	httpErr := &common.HTTPError{}
	if errors.As(err, &httpErr) {
		body := string(httpErr.Body)

		return strings.Contains(body, "custom-object-read") &&
			strings.Contains(body, "MISSING_SCOPES") &&
			httpErr.Status == http.StatusForbidden
	}

	return false
}

type schemaResponse struct {
	RequiredProperties []string           `json:"requiredProperties"`
	Properties         []fieldDescription `json:"properties"`
}
