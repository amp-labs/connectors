package instantly

import "github.com/amp-labs/connectors/internal/datautils"

const (
	objectNameCampaigns        = "campaigns"
	objectNameAccounts         = "accounts"
	objectNameEmails           = "emails"
	objectNameTags             = "tags"
	objectNameLeads            = "leads"
	objectNameBlocklistEntries = "blocklist-entries"
	objectNameUniboxReplies    = "unibox-replies"
)

var supportedObjectsByRead = datautils.NewSet( //nolint:gochecknoglobals
	// Object Name	----------	API endpoint path
	objectNameCampaigns, // campaign/list
	objectNameAccounts,  // account/list
	objectNameEmails,    // unibox/emails
	objectNameTags,      // custom-tag
)

var supportedObjectsByWrite = datautils.NewSet( //nolint:gochecknoglobals
	// Object Name	----------	API endpoint path
	objectNameTags,             // custom-tag
	objectNameLeads,            // lead/add
	objectNameBlocklistEntries, // blocklist/add/entries
	objectNameUniboxReplies,    // unibox/emails/reply
)

var supportedObjectsByDelete = datautils.NewSet( //nolint:gochecknoglobals
	// Delete tag.
	// https://developer.instantly.ai/tags/delete-a-tag
	objectNameTags,
)
