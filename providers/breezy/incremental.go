package breezy

import (
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/datautils"
)

// objectTimeField maps read objects to the timestamp field used for connector-side
// incremental filtering when ReadParams.Since/Until are set.
//
// Breezy list endpoints do not expose updated_since-style query params on these
// objects; we fetch the full list and filter in the connector (Salesfinity pattern).
var objectTimeField = datautils.NewDefaultMap( //nolint:gochecknoglobals
	datautils.Map[string, string]{
		objectPositions: "updated_date",
	},
	func(string) string { return "" },
)

func makeFilterFunc(params common.ReadParams) common.RecordsFilterFunc {
	timeField := objectTimeField.Get(params.ObjectName)
	if timeField == "" {
		return readhelper.MakeIdentityFilterFunc(noNextPage)
	}

	// List endpoints do not document a stable sort order (unlike position candidates).
	return readhelper.MakeTimeFilterFunc(
		readhelper.Unordered,
		readhelper.NewTimeBoundary(),
		timeField,
		time.RFC3339,
		noNextPage,
	)
}
