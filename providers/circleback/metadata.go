package circleback

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/circleback/internal/metadata"
)

// listObjectMetadata serves the catalog generated from Circleback's OpenAPI spec, plus the
// meetings `transcript` field, which Read fetches from GET /meeting/{meetingId}/transcript.
func (c *Connector) listObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	selection, err := metadata.Catalog.Select(objectNames)
	if err != nil {
		return nil, err
	}

	// No-op when meetings was not requested.
	selection.AddField(objectNameMeetings, fieldTranscript, common.FieldMetadata{
		DisplayName:  "Transcript",
		ValueType:    common.ValueTypeOther,
		ProviderType: "array",
	})

	return selection.Result(), nil
}
