package github

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/github/metadata"
)

//nolint:lll,gochecknoglobals
var (
	// https://docs.github.com/en/rest/gists/gists?apiVersion=2022-11-28#create-a-gist
	// https://docs.github.com/en/rest/gists/gists?apiVersion=2022-11-28#delete-a-gist
	objectNameGist = "gists"

	// https://docs.github.com/en/rest/users/emails?apiVersion=2022-11-28#add-an-email-address-for-the-authenticated-user
	objectNameUserEmails = "user/emails"

	// https://docs.github.com/en/rest/users/gpg-keys?apiVersion=2022-11-28#create-a-gpg-key-for-the-authenticated-user
	// https://docs.github.com/en/rest/users/gpg-keys?apiVersion=2022-11-28#delete-a-gpg-key-for-the-authenticated-user
	objectNameUserGpgKeys = "user/gpg_keys"

	// https://docs.github.com/en/rest/users/keys?apiVersion=2022-11-28#create-a-public-ssh-key-for-the-authenticated-user
	// https://docs.github.com/en/rest/users/keys?apiVersion=2022-11-28#delete-a-public-ssh-key-for-the-authenticated-user
	objectNameUserKeys = "user/keys"

	// https://docs.github.com/en/rest/users/ssh-signing-keys?apiVersion=2022-11-28#create-a-ssh-signing-key-for-the-authenticated-user
	// https://docs.github.com/en/rest/users/ssh-signing-keys?apiVersion=2022-11-28#delete-an-ssh-signing-key-for-the-authenticated-user
	objectNameUserSSHSigningKeys = "user/ssh_signing_keys"

	// https://docs.github.com/en/rest/users/social-accounts?apiVersion=2022-11-28#add-social-accounts-for-the-authenticated-user
	objectNameUserSocialAccounts = "user/social_accounts"

	// https://docs.github.com/en/rest/users/followers?apiVersion=2022-11-28#unfollow-a-user
	objectNameUserFollowing = "user/following"

	// https://docs.github.com/en/rest/users/blocking?apiVersion=2022-11-28#unblock-a-user
	objectNameUserBlocks = "user/blocks"

	// https://docs.github.com/en/rest/orgs/orgs?apiVersion=2022-11-28#delete-an-organization
	objectNameOrgs = "orgs"

	// https://docs.github.com/en/rest/codespaces/codespaces?apiVersion=2022-11-28#create-a-codespace-for-the-authenticated-user
	// https://docs.github.com/en/rest/codespaces/codespaces?apiVersion=2022-11-28#update-a-codespace-for-the-authenticated-user
	// https://docs.github.com/en/rest/codespaces/codespaces?apiVersion=2022-11-28#delete-a-codespace-for-the-authenticated-user
	objectNameUserCodespaces = "user/codespaces"
)

var (
	supportPagination = datautils.NewSet( //nolint: gochecknoglobals
		"advisories", "blocks", "classrooms", "user/codespaces",
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

	supportByUpdate = datautils.NewSet( //nolint: gochecknoglobals
		objectNameGist, objectNameUserCodespaces,
	)
)

func supportedOperations() components.EndpointRegistryInput {
	readSupport := metadata.Schemas.ObjectNames().GetList(common.ModuleRoot)

	// nolint:lll
	writeSupport := []string{
		objectNameGist, objectNameUserEmails, objectNameUserGpgKeys, objectNameUserKeys, objectNameUserSSHSigningKeys, objectNameUserSocialAccounts,
		objectNameUserCodespaces,
	}

	// nolint:lll
	deleteSupport := []string{
		objectNameUserGpgKeys, objectNameUserKeys, objectNameUserSSHSigningKeys, objectNameUserFollowing, objectNameUserEmails, objectNameUserBlocks,
		objectNameOrgs, objectNameGist, objectNameUserCodespaces,
	}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupport, ",")),
				Support:  components.WriteSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(deleteSupport, ",")),
				Support:  components.DeleteSupport,
			},
		},
	}
}
