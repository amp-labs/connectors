package freshdesk

import (
	"context"
	"log/slog"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
)

func (conn *Connector) ListObjectMetadata(ctx context.Context,
	objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	// Ensure that objectNames is not empty
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	metadataResults := common.ListObjectMetadataResult{
		Result: make(map[string]common.ObjectMetadata),
		Errors: make(map[string]error),
	}

	for _, object := range objectNames {
		if !readSupportedObjects.Has(object) {
			metadataResults.Errors[object] = common.ErrObjectNotSupported

			continue
		}

		// A failure is recorded against the object rather than aborting, so that one
		// unreachable endpoint doesn't discard the metadata of every other object.
		fields, err := conn.describeObject(ctx, object)
		if err != nil {
			metadataResults.Errors[object] = err

			continue
		}

		metadataResults.Result[object] = *common.NewObjectMetadata(
			naming.CapitalizeFirstLetterEveryWord(object), fields,
		)
	}

	return &metadataResults, nil
}

// describeObject derives an object's fields from the records it holds, then applies what
// Freshdesk declares about them for the few objects it introspects. The records are what name
// the fields, since they are the ones a read returns, while a declaration tells their type,
// label and choices, which a value alone cannot reveal.
func (conn *Connector) describeObject(ctx context.Context, object string) (common.FieldsMetadata, error) {
	fields, err := conn.sampleFields(ctx, object)
	if err != nil {
		return nil, err
	}

	// Declarations are an enrichment, so the sampled metadata stands if they cannot be read.
	if err := conn.overlayDeclaredFields(ctx, object, fields); err != nil {
		slog.Warn("cannot read field declarations, describing from data alone",
			"object", object, "error", err)
	}

	return fields, nil
}

// sampleFields reads a page of records and derives the fields from their values.
func (conn *Connector) sampleFields(ctx context.Context, object string) (common.FieldsMetadata, error) {
	url, err := conn.getAPIURL(object)
	if err != nil {
		return nil, err
	}

	withPageSize(url, object, metadataPageSize)

	response, err := conn.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return buildMetadataFields(response)
}

func buildMetadataFields(response *common.JSONHTTPResponse) (common.FieldsMetadata, error) {
	sample, err := extractSampleRecords(response)
	if err != nil {
		return nil, err
	}

	if len(sample) == 0 {
		return nil, common.ErrMissingExpectedValues
	}

	return makeFieldsMetadata(sample), nil
}

// extractSampleRecords returns the records to infer field types from,
// reading the response body the same way that Read does.
func extractSampleRecords(response *common.JSONHTTPResponse) ([]map[string]any, error) {
	body, ok := response.Body()
	if !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	sample, err := extractRecords(body)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	return sample, nil
}

// makeFieldsMetadata derives field types from the sampled records.
// Most objects are described by their payload alone, Freshdesk declaring the fields of only a few.
// Every record of the page is inspected, which allows a field that is null in the earlier
// records to be typed from a later one.
func makeFieldsMetadata(records []map[string]any) common.FieldsMetadata {
	fields := make(common.FieldsMetadata)

	for _, record := range records {
		for fld, value := range record {
			if known, seen := fields[fld]; seen && known.ValueType != common.ValueTypeOther {
				continue // already typed by an earlier record.
			}

			fields[fld] = common.FieldMetadata{
				DisplayName:  fld,
				ValueType:    inferValueType(value),
				ProviderType: "", // not available.
				Values:       nil,
			}
		}
	}

	return fields
}

// inferValueType maps a sample value onto an Ampersand type.
// Timestamps are carried as ISO-8601 strings, which would otherwise be reported as plain text.
func inferValueType(value any) common.ValueType {
	text, ok := value.(string)
	if !ok {
		return common.InferValueTypeFromData(value)
	}

	if _, err := time.Parse(time.RFC3339, text); err == nil {
		return common.ValueTypeDateTime
	}

	if _, err := time.Parse(time.DateOnly, text); err == nil {
		return common.ValueTypeDate
	}

	return common.ValueTypeString
}
