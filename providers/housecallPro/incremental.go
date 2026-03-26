package housecallpro

import (
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

type objectReadSpec struct {
	supportsIncremental bool
	// timeKey is the record field used as API sort_by and for connector-side Since/Until filtering.
	timeKey string
}

func incrementalUpdatedAt() objectReadSpec {
	return objectReadSpec{
		supportsIncremental: true,
		timeKey:             "updated_at",
	}
}

//nolint:gochecknoglobals
var objectReadSpecs = datautils.NewDefaultMap(map[string]objectReadSpec{
	"customers":                      incrementalUpdatedAt(),
	"estimates":                      incrementalUpdatedAt(),
	"jobs":                           incrementalUpdatedAt(),
	"price_book/material_categories": incrementalUpdatedAt(),
	"events":                         incrementalUpdatedAt(),
}, func(string) objectReadSpec {
	// Objects not listed here omit sort_by / sort_direction and do not apply connector-side Since/Until.
	return objectReadSpec{}
})

func makeFilterFunc(params common.ReadParams, reqURL *urlbuilder.URL) common.RecordsFilterFunc {
	nextPage := makeNextRecordsURL(reqURL)

	spec := objectReadSpecs.Get(params.ObjectName)
	if !spec.supportsIncremental {
		return readhelper.MakeIdentityFilterFunc(nextPage)
	}

	return readhelper.MakeTimeFilterFunc(
		readhelper.ReverseOrder,
		readhelper.NewTimeBoundary(),
		spec.timeKey,
		time.RFC3339,
		nextPage,
	)
}
