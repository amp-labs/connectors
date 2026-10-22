package files

import _ "embed"

// nolint:gochecknoglobals
var (
	WordPressMetadataFile         = "providers/wordpress/internal/metadata/objectsMetadata.json.gz"
	CombinedResponsesFileLocation = "scripts/scraper/wordpress/internal/files/combined-responses.json"

	//go:embed combined-responses.json
	CombinedResponses []byte

	// Endpoints is a list of endpoints that represent collections.
	// This list is manually created from the discovery endpoint http://localhost:4444/wp/v2/.
	// This assumes that WordPress requests are performed via the local proxy.
	Endpoints = []string{
		"/wp/v2/sites/{workspace_id}/activity-log-events",
		"/wp/v2/sites/{workspace_id}/block-patterns/patterns",
		"/wp/v2/sites/{workspace_id}/block-types",
		"/wp/v2/sites/{workspace_id}/blocks",
		"/wp/v2/sites/{workspace_id}/comments",
		"/wp/v2/sites/{workspace_id}/feedback",
		"/wp/v2/sites/{workspace_id}/fields",
		"/wp/v2/sites/{workspace_id}/font-collections",
		"/wp/v2/sites/{workspace_id}/font-families",
		"/wp/v2/sites/{workspace_id}/guidelines",
		"/wp/v2/sites/{workspace_id}/icon-collections",
		"/wp/v2/sites/{workspace_id}/icons",
		"/wp/v2/sites/{workspace_id}/jetpack-forms",
		"/wp/v2/sites/{workspace_id}/jp_pay_order",
		"/wp/v2/sites/{workspace_id}/jp_pay_product",
		"/wp/v2/sites/{workspace_id}/knowledge/guideline-scopes",
		"/wp/v2/sites/{workspace_id}/knowledge",
		"/wp/v2/sites/{workspace_id}/media",
		"/wp/v2/sites/{workspace_id}/menu-items",
		"/wp/v2/sites/{workspace_id}/menu-locations",
		"/wp/v2/sites/{workspace_id}/menus",
		"/wp/v2/sites/{workspace_id}/navigation",
		"/wp/v2/sites/{workspace_id}/pages",
		"/wp/v2/sites/{workspace_id}/pattern-directory/patterns",
		"/wp/v2/sites/{workspace_id}/plugins",
		"/wp/v2/sites/{workspace_id}/posts",
		"/wp/v2/sites/{workspace_id}/sidebars",
		"/wp/v2/sites/{workspace_id}/statuses",
		"/wp/v2/sites/{workspace_id}/tags",
		"/wp/v2/sites/{workspace_id}/taxonomies",
		"/wp/v2/sites/{workspace_id}/template-parts",
		"/wp/v2/sites/{workspace_id}/templates",
		"/wp/v2/sites/{workspace_id}/themes",
		"/wp/v2/sites/{workspace_id}/types",
		"/wp/v2/sites/{workspace_id}/users",
		"/wp/v2/sites/{workspace_id}/view-config",
		"/wp/v2/sites/{workspace_id}/widget-types",
		"/wp/v2/sites/{workspace_id}/widgets",
		"/wp/v2/sites/{workspace_id}/wp-abilities/v1/abilities",
		"/wp/v2/sites/{workspace_id}/wp_guideline_type",
		"/wp/v2/sites/{workspace_id}/wp_knowledge_type",
		"/wp/v2/sites/{workspace_id}/wp_pattern_category",
	}
)
