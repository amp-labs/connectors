package gitlab

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/staticschema"
)

var objectResponders = datautils.NewSet( //nolint:gochecknoglobals
	"application/appearance", "application/plan_limits", "application/settings",
	"application/statistics", "license", "metadata", "sidekiq/compound_metrics", "sidekiq/job_stats",
	"sidekiq/process_metrics", "sidekiq/queue_metrics", "usage_data/queries", "user/emails",
	"user/preferences", "user/status", "user/support_pin", "user_counts",
	"version", "web_commits/public_key",
)

// nolint: funlen
func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		"admin/ci/variables",
		"admin/search/migrations",
		"application/appearance",
		"application/plan_limits",
		"application/settings",
		"applications",
		"audit_events",
		"broadcast_messages",
		"bulk_imports",
		"bulk_imports/entities",
		"deploy_keys",
		"deploy_tokens",
		"events",
		"experiments",
		"geo_nodes/status",
		"geo_sites",
		"geo_sites/status",
		"groups",
		"hooks",
		"issues",
		"issues_statistics",
		"license",
		"licenses",
		"member_roles",
		"merge_requests",
		"metadata",
		"namespaces",
		"pages/domains",
		"personal_access_tokens",
		"personal_access_tokens/self/associations",
		"projects",
		"project_aliases",
		"project_repository_storage_moves",
		"runners",
		"runners/all",
		"service_accounts",
		"sidekiq/compound_metrics",
		"sidekiq/job_stats",
		"sidekiq/process_metrics",
		"sidekiq/queue_metrics",
		"snippet_repository_storage_moves",
		"snippets",
		"snippets/all",
		"snippets/public",
		"templates/dockerfiles",
		"templates/gitignores",
		"templates/gitlab_ci_ymls",
		"templates/licenses",
		"todos",
		"topics",
		"user/activities",
		"user/emails",
		"user/gpg_keys",
		"user/keys",
		"users",
		"web_commits/public_key",
	}

	writeSupport := []string{
		"admin/ci/variables",
		"applications",
		"broadcast_messages",
		"bulk_imports",
		"chat/completions",
		"code_suggestions/completions",
		"deploy_keys",
		"geo_nodes",
		"geo_sites",
		"group_repository_storage_moves",
		"groups",
		"hooks",
		"import/bitbucket",
		"import/bitbucket_server",
		"import/github",
		"import/github/cancel",
		"import/github/gists",
		"member_roles",
		"organizations",
		"project_aliases",
		"project_repository_storage_moves",
		"projects",
		"runners",
		"security/vulnerability_exports",
		"service_accounts",
		"snippets",
		"todos/mark_as_done",
		"topics",
		"user/emails",
		"user/gpg_keys",
		"user/keys",
		"user/personal_access_tokens",
		"user/runners",
		"user/support_pin",
		"users",
	}

	return components.EndpointRegistryInput{
		staticschema.RootModuleID: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupport, ",")),
				Support:  components.WriteSupport,
			},
		},
	}
}
