package dynamicsbusiness

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// LCID decimal: https://wiki.freepascal.org/Language_Codes
const languageCodeEnglishUSA = 1033

var ErrMetadataNotFound = errors.New("metadata for object was not found")

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	data, err := common.UnmarshalJSON[metadataResponse](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	var object *metadataValueResponse

	for _, entity := range data.Value {
		if strings.EqualFold(entity.EntitySetName, objectName) {
			object = &entity //nolint: exportloopref
		}
	}

	if object == nil {
		// This error is unlikely but for sanity check
		// ensure that response always has the object with the requested name.
		return nil, ErrMetadataNotFound
	}

	fields := object.Fields()

	displayName := naming.CapitalizeFirstLetterEveryWord(objectName)

	return common.NewObjectMetadata(displayName, fields), nil
}

// nolint:tagliatelle
type metadataResponse struct {
	OdataContext string                  `json:"@odata.context"`
	Value        []metadataValueResponse `json:"value"`
}

type metadataValueResponse struct {
	EntityName        string             `json:"entityName"`
	EntitySetName     string             `json:"entitySetName"`
	EntityCaptions    []any              `json:"entityCaptions"`
	EntitySetCaptions []any              `json:"entitySetCaptions"`
	Properties        []metadataProperty `json:"properties"`
	Actions           []any              `json:"actions"`
	EnumMembers       []any              `json:"enumMembers"`
}

func (r metadataValueResponse) Fields() map[string]common.FieldMetadata {
	fields := make(common.FieldsMetadata)

	for _, property := range r.Properties {
		displayName := property.Name

		// Find display name for English audience.
		for _, caption := range property.Captions {
			if caption.LanguageCode == languageCodeEnglishUSA {
				displayName = caption.Caption
			}
		}

		fields.AddFieldWithDisplayOnly(property.Name, displayName)
	}

	return fields
}

type metadataProperty struct {
	Name     string `json:"name"`
	Captions []struct {
		LanguageCode int    `json:"languageCode"`
		Caption      string `json:"caption"`
	} `json:"captions"`
}

func (c *Connector) parseReadResponse(
	ctx context.Context, params common.ReadParams,
	request *http.Request, response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	return common.ParseResult(
		response,
		common.ExtractOptionalRecordsFromPath("value"),
		func(node *ajson.Node) (string, error) {
			return jsonquery.New(node).StrWithDefault("@odata.nextLink", "")
		},
		common.GetMarshaledData,
		params.Fields,
	)
}
