package jump

import (
	"context"
	"net/http"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
)

// optionalMetadataFields lists nested relations that read queries only include when requested.
var optionalMetadataFields = map[string][]string{ //nolint:gochecknoglobals
	"contacts":     {"contactInfo", "integrationReferences"},
	"documents":    {"parsedResult"},
	"meetingPreps": {"answers", "prepContents"},
	"meetings": {
		"host", "meetingEvents", "notes",
		"owner", "participants", "pulseResults",
		"recordSuggestions", "scorecardResults", "signalResults",
		"tasks", "transcript",
	},
	"scorecards": {"criteria"},
	"tasks":      {"assignee"},
}

var objectDisplayNames = map[string]string{ //nolint:gochecknoglobals
	"contacts":          "Contacts",
	"documents":         "Documents",
	"integrations":      "Integrations",
	"meetingPreps":      "Meeting Preps",
	"meetings":          "Meetings",
	"notes":             "Notes",
	"pulses":            "Pulses",
	"scorecards":        "Scorecards",
	"signalDefinitions": "Signal Definitions",
	"tasks":             "Tasks",
	"users":             "Users",
}

func metadataReadParams(objectName string) common.ReadParams {
	return common.ReadParams{
		ObjectName: objectName,
		Fields:     connectors.Fields(optionalMetadataFields[objectName]...),
		PageSize:   1,
	}
}

// GraphQL introspection is not available on the Jump API, so object metadata is
// inferred by sampling the first record returned by a read query.
func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	return c.buildReadRequest(ctx, metadataReadParams(objectName))
}

func (c *Connector) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	result, err := c.parseReadResponse(ctx, metadataReadParams(objectName), request, response)
	if err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	objectMetadata := common.ObjectMetadata{
		Fields:      make(map[string]common.FieldMetadata),
		FieldsMap:   make(map[string]string),
		DisplayName: objectDisplayNames[objectName],
	}

	for field, value := range result.Data[0].Raw {
		objectMetadata.AddFieldMetadata(field, common.FieldMetadata{
			DisplayName:  field,
			ValueType:    common.InferValueTypeFromData(value),
			ProviderType: "",
			Values:       nil,
		})
	}

	return &objectMetadata, nil
}
