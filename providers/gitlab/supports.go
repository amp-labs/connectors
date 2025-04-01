package gitlab

import "github.com/amp-labs/connectors/internal/datautils"

var objectResponders = datautils.NewSet( //nolint:gochecknoglobals
	"application/appearance", "application/plan_limits", "application/settings",
	"application/statistics", "license", "metadata", "sidekiq/compound_metrics", "sidekiq/job_stats",
	"sidekiq/process_metrics", "sidekiq/queue_metrics", "usage_data/queries", "user/emails",
	"user/preferences", "user/status", "user/support_pin", "user_counts",
	"version", "web_commits/public_key",
)
