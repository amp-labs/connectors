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
		"applications",
		"broadcast_messages",
		"code_suggestions/completions",
		"deploy_keys",
		"geo_nodes",
		"geo_sites",
		"chat/completions",
		"groups",
		"bulk_imports",
		"group_repository_storage_moves",
		"import/github",
		"import/github/cancel",
		"import/github/gists",
		"import/bitbucket_server",
		"import/bitbucket",
		"admin/ci/variables",
		"member_roles",
		"organizations",
		"projects",
		"project_aliases",
		"project_repository_storage_moves",
		"runners",
		"snippets",
		"hooks",
		"todos/mark_as_done",
		"topics",
		"users",
		"user/runners",
		"user/support_pin",
		"service_accounts",
		"user/keys",
		"user/gpg_keys",
		"user/emails",
		"user/personal_access_tokens",
		"security/vulnerability_exports",
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
