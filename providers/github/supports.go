package github

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/providers/github/metadata"
)

var (
	supportPagination = datautils.NewSet( //nolint: gochecknoglobals
		"advisories", "blocks", "classrooms", "codespaces",
		"deliveries", "emails", "events", "followers", "following", "gists",
		"gists/starred", "gpg_keys", "installation-requests", "installation/repositories",
		"issues", "keys", "licenses", "marketplace_listing/plans",
		"marketplace_listing/stubbed/plans", "marketplace_purchases", "migrations",
		"notifications", "orgs", "packages", "gists/public", "public_emails", "repos",
		"repository_invitations", "secrets", "social_accounts", "ssh_signing_keys",
		"stubbed", "subscriptions", "teams", "user/installations", "user/issues",
		"user/memberships/orgs", "user/starred",
	)

	supportSince = datautils.NewSet( //nolint: gochecknoglobals
		"gists", "gists/starred", "issues",
		"gists/public", "repos", "user/issues",
	)
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := metadata.Schemas.ObjectNames().GetList(staticschema.RootModuleID)

	return components.EndpointRegistryInput{
		staticschema.RootModuleID: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}
