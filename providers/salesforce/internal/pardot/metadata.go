package pardot

import (
	"context"
	_ "embed"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/tools/scrapper"
)

var (
	// Static file containing a list of object metadata is embedded and can be served.
	//
	//go:embed schemas.json
	schemas []byte

	fileManager = scrapper.NewReader[staticschema.FieldMetadataMapV2](schemas) // nolint:gochecknoglobals

	// Schemas is cached Object schemas.
	Schemas = pardotSchemas{ // nolint:gochecknoglobals
		Metadata: fileManager.MustLoadSchemas(),
	}
)

const (
	prospectsObjectName  = "prospects"
	customFieldsEndpoint = "api/v5/objects/custom-fields"
	customFieldsPageSize = "1000"
	customFieldsFields   = "id,name,fieldId,type,crmManaged,isRecordMultipleResponseType"
)

func (a *Adapter) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	result, err := Schemas.Select(objectNames)
	if err != nil {
		return nil, err
	}

	if !containsFold(objectNames, prospectsObjectName) {
		return result, nil
	}

	customFields, fetchErr := a.fetchProspectCustomFields(ctx)
	if fetchErr != nil {
		logging.Logger(ctx).Warn(
			"pardot: unable to fetch prospect custom fields; returning static schema only",
			"error", fetchErr,
		)

		return result, nil
	}

	mergeProspectCustomFields(result, customFields)

	return result, nil
}

type pardotSchemas struct {
	*staticschema.Metadata[staticschema.FieldMetadataMapV2, any]
}

func (s *pardotSchemas) Select(objectNames []string) (*common.ListObjectMetadataResult, error) {
	// Case-insensitive object names.
	objects := make([]string, len(objectNames))
	for index, name := range objectNames {
		objects[index] = strings.ToLower(name)
	}

	return s.Metadata.Select(providers.ModuleSalesforceAccountEngagement, objects)
}

// pardotCustomField models a single entry returned by the Pardot v5
// /api/v5/objects/custom-fields endpoint. Only the fields needed to expose
// the field in metadata are modeled.
//
// https://developer.salesforce.com/docs/marketing/pardot/guide/custom-field-v5.html
type pardotCustomField struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	FieldID string `json:"fieldId"`
	Type    string `json:"type"`
}

type pardotCustomFieldsResponse struct {
	Values      []pardotCustomField `json:"values"`
	NextPageURL string              `json:"nextPageUrl"`
}

func (a *Adapter) fetchProspectCustomFields(ctx context.Context) ([]pardotCustomField, error) {
	url, err := urlbuilder.New(a.getModuleURL(), customFieldsEndpoint)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("fields", customFieldsFields)
	url.WithQueryParam("limit", customFieldsPageSize)

	header := common.Header{Key: "Pardot-Business-Unit-Id", Value: a.businessUnitID}

	var all []pardotCustomField

	nextURL := url.String()

	for {
		resp, err := a.JSONHTTPClient().Get(ctx, nextURL, header)
		if err != nil {
			return nil, err
		}

		page, err := common.UnmarshalJSON[pardotCustomFieldsResponse](resp)
		if err != nil {
			return nil, err
		}

		all = append(all, page.Values...)

		if page.NextPageURL == "" {
			break
		}

		nextURL = page.NextPageURL
	}

	return all, nil
}

func mergeProspectCustomFields(
	result *common.ListObjectMetadataResult, customFields []pardotCustomField,
) {
	if result == nil || len(customFields) == 0 {
		return
	}

	meta, ok := result.Result[prospectsObjectName]
	if !ok {
		return
	}

	if meta.Fields == nil {
		meta.Fields = common.FieldsMetadata{}
	}

	for _, cf := range customFields {
		key := cf.FieldID
		if key == "" {
			continue
		}

		display := cf.Name
		if display == "" {
			display = cf.FieldID
		}

		meta.Fields[key] = common.FieldMetadata{
			DisplayName:  display,
			ValueType:    pardotCustomFieldValueType(cf.Type),
			ProviderType: cf.Type,
			IsCustom:     goutils.Pointer(true),
		}
	}

	result.Result[prospectsObjectName] = meta
}

// pardotCustomFieldValueType maps the Pardot CustomField "type" string to
// Ampersand's common.ValueType. Pardot types come from a fixed vocabulary:
// Text, TextArea, Number, Date, CRMUser, Dropdown, Radio Button, Checkbox,
// Multi-Select, Hidden.
func pardotCustomFieldValueType(pardotType string) common.ValueType {
	switch strings.ToLower(strings.ReplaceAll(pardotType, " ", "")) {
	case "text", "textarea", "hidden", "crmuser":
		return common.ValueTypeString
	case "number":
		return common.ValueTypeFloat
	case "date":
		return common.ValueTypeDate
	case "dropdown", "radiobutton":
		return common.ValueTypeSingleSelect
	case "checkbox":
		return common.ValueTypeMultiSelect
	case "multi-select", "multiselect":
		return common.ValueTypeMultiSelect
	default:
		return common.ValueTypeOther
	}
}

func containsFold(list []string, target string) bool {
	for _, item := range list {
		if strings.EqualFold(item, target) {
			return true
		}
	}

	return false
}
