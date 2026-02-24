package associations

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/core"
)

func CreateDataMarshallerWithAssociations(
	ctx context.Context, filler Filler,
	objectName string, associatedObjects []string,
) common.MarshalFromNodeFunc {
	return readhelper.ChainedMarshaller(
		core.GetDataMarshaller(),
		// Enhance records with associations by fetching these relationships.
		func(rows []common.ReadResultRow) error {
			return filler.FillAssociations(ctx, objectName, &rows, associatedObjects)
		},
	)
}
