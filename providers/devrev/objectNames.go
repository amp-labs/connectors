package devrev

import (
	"github.com/amp-labs/connectors/internal/datautils"
)

// objectsWithModifiedDateFilter lists object names whose list endpoints support
// modified_date.after and modified_date.before query parameter.
var objectsWithModifiedDateFilter = datautils.NewSet( //nolint:gochecknoglobals
	"accounts",
	"artifacts",
	"brands",
	"code-changes",
	"conversations",
	"engagements",
	"groups",
	"incidents",
	"jobs",
	"links",
	"meetings",
	"parts",
	"rev-orgs",
	"rev-users",
	"timeline-entries",
	"works",
)
