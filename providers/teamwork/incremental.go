package teamwork

import "github.com/amp-labs/connectors/internal/datautils"

var objectsWithSinceQuery = datautils.NewSet( // nolint:gochecknoglobals
	"comments",
	"companies",
	"dashboards",
	"jobroles",
	"latestactivity",
	"me/timers",
	"messages",
	"milestones",
	"notebooks",
	"people",
	"projects",
	"risks",
	"skills",
	"starred",
	"statuses",
	"tags",
	"tasklists",
	"tasks",
	"templates",
	"time",
	"timers",
	"updates",
)

var objectsWithUntilQuery = datautils.NewSet( // nolint:gochecknoglobals
	"jobroles",
	"skills",
	"tasks",
)
