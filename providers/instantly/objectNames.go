package instantly

import (
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/internal/deep"
)

const (
	objectNameCampaigns        = "campaigns"
	objectNameAccounts         = "accounts"
	objectNameEmails           = "emails"
	objectNameTags             = "tags"
	objectNameLeads            = "leads"
	objectNameBlocklistEntries = "blocklist-entries"
	objectNameUniboxReplies    = "unibox-replies"
)

var supportedObjectsByRead = handy.NewSet( //nolint:gochecknoglobals
	// Object Name	----------	API endpoint path
	objectNameCampaigns, // campaign/list
	objectNameAccounts,  // account/list
	objectNameEmails,    // unibox/emails
	objectNameTags,      // custom-tag
)

var supportedObjectsByWrite = handy.NewSet( //nolint:gochecknoglobals
	// Object Name	----------	API endpoint path
	objectNameTags,             // custom-tag
	objectNameLeads,            // lead/add
	objectNameBlocklistEntries, // blocklist/add/entries
	objectNameUniboxReplies,    // unibox/emails/reply
)

var supportedObjectsByDelete = handy.NewSet( //nolint:gochecknoglobals
	// Delete tag.
	// https://developer.instantly.ai/tags/delete-a-tag
	objectNameTags,
)

var objectResolver = handy.Map[string, deep.ObjectData]{
	// https://developer.instantly.ai/campaign-1/list-campaigns
	// Empty string of data location means the response is an array itself holding what we need.
	objectNameCampaigns: {
		URLPath:  "campaign/list",
		NodePath: "",
	},
	// https://developer.instantly.ai/account/list-accounts
	objectNameAccounts: {
		URLPath:  "account/list",
		NodePath: "accounts",
	},
	// https://developer.instantly.ai/unibox/emails-or-list
	objectNameEmails: {
		URLPath:  "unibox/emails",
		NodePath: "data",
	},
	// https://developer.instantly.ai/tags/list-tags
	objectNameTags: {
		URLPath:  "custom-tag",
		NodePath: "data",
	},
}
