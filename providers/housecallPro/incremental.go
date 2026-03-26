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

// incr is a list object that supports connector-side incremental read on timeKey.
func incr(timeKey string) objectReadSpec {
	return objectReadSpec{
		supportsIncremental: true,
		timeKey:             timeKey,
	}
}

//nolint:gochecknoglobals
var objectReadSpecs = datautils.NewDefaultMap(map[string]objectReadSpec{
	"customers":                      incr("updated_at"),
	"estimates":                      incr("updated_at"),
	"jobs":                           incr("updated_at"),
	"price_book/material_categories": incr("updated_at"),
	"events":                         incr("updated_at"),
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
