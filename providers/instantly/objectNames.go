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

var (
	supportedObjectsByRead   = readObjects.KeySet()                                 //nolint:gochecknoglobals
	supportedObjectsByWrite  = handy.MergeSets(createObjects.KeySet(), updateObjects.KeySet()) //nolint:gochecknoglobals
	supportedObjectsByDelete = deleteObjects.KeySet()                               //nolint:gochecknoglobals
)

var readObjects = handy.Map[string, deep.ObjectData]{
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

var createObjects = handy.Map[string, string]{
	// Add lead to campaign.
	// https://developer.instantly.ai/campaign/add-leads-to-a-campaign
	objectNameLeads: "lead/add",
	// Add blocklist entry.
	// https://developer.instantly.ai/blocklist/add-entries-to-blocklist
	objectNameBlocklistEntries: "blocklist/add/entries",
	// Create message - unibox reply.
	// https://developer.instantly.ai/unibox/send-reply
	objectNameUniboxReplies: "unibox/emails/reply",
	// Create tag.
	// https://developer.instantly.ai/tags/create-a-new-tag
	objectNameTags: "custom-tag",
}

var updateObjects = handy.Map[string, string]{
	// Update tag.
	// https://developer.instantly.ai/tags/update-tag
	objectNameTags: "custom-tag",
}

var deleteObjects = handy.Map[string, string]{
	// Delete tag.
	// https://developer.instantly.ai/tags/delete-a-tag
	objectNameTags: "custom-tag",
}

var writeResponseRecordIdPaths = map[string]*string{ // nolint:gochecknoglobals
	objectNameLeads:            nil, // ID is not returned for Leads.
	objectNameBlocklistEntries: handy.Pointers.Str("blocklist_id"),
	objectNameUniboxReplies:    handy.Pointers.Str("message_id"),
	objectNameTags:             handy.Pointers.Str("id"),
}
