package constantcontact

import (
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/constantcontact/metadata"
)

const (
	objectNameAccountEmails           = "account_emails"
	objectNameAccounts                = "accounts"
	objectNameContactCustomFields     = "contact_custom_fields"
	objectNameContactLists            = "contact_lists"
	objectNameContactTags             = "contact_tags"
	objectNameContacts                = "contacts"
	objectNameEmailCampaignActivities = "email_campaign_activities"
	objectNameEmailCampaigns          = "email_campaigns"
	objectNameSegments                = "segments"
	objectNameSubscriptions           = "subscriptions"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals

// nolint: lll,gochecknoglobals
var supportedObjectsByCreate = datautils.NewSet(
	// https://v3.developer.constantcontact.com/api_reference/index.html#!/Account_Services/addAccountEmailAddress
	objectNameAccountEmails,
	// https://v3.developer.constantcontact.com/api_reference/index.html#!/Technology_Partners/provision
	objectNameAccounts,
	// https://v3.developer.constantcontact.com/api_reference/index.html#!/Contacts_Custom_Fields/postCustomFields
	objectNameContactCustomFields,
	// https://v3.developer.constantcontact.com/api_reference/index.html#!/Contact_Lists/createList
	objectNameContactLists,
	// https://v3.developer.constantcontact.com/api_reference/index.html#!/Contact_Tags/postTag
	objectNameContactTags,
	// https://v3.developer.constantcontact.com/api_reference/index.html#!/Contacts/createContact
	objectNameContacts,
	// https://v3.developer.constantcontact.com/api_reference/index.html#!/Email_Campaigns/createEmailCampaignUsingPOST
	objectNameEmailCampaigns,
	// https://v3.developer.constantcontact.com/api_reference/index.html#!/Segments/createSegment
	objectNameSegments,
)

// nolint: lll,gochecknoglobals
var supportedObjectsByUpdate = datautils.NewSet(
	// https://v3.developer.constantcontact.com/api_reference/index.html#!/Contacts_Custom_Fields/putCustomField
	objectNameContactCustomFields,
	// https://v3.developer.constantcontact.com/api_reference/index.html#!/Contact_Lists/putList
	objectNameContactLists,
	// https://v3.developer.constantcontact.com/api_reference/index.html#!/Contact_Tags/putTag
	objectNameContactTags,
	// https://v3.developer.constantcontact.com/api_reference/index.html#!/Contacts/putContact
	objectNameContacts,
	// https://v3.developer.constantcontact.com/api_reference/index.html#!/Email_Campaigns/updateEmailCampaignActivityUsingPUT
	objectNameEmailCampaignActivities,
	// https://v3.developer.constantcontact.com/api_reference/index.html#!/Email_Campaigns/renameEmailCampaignUsingPATCH
	objectNameEmailCampaigns,
	// https://v3.developer.constantcontact.com/api_reference/index.html#!/Segments/updateSegment
	objectNameSegments,
	// https://v3.developer.constantcontact.com/api_reference/index.html#!/Technology_Partners_Webhooks/putWebhooksTopic
	objectNameSubscriptions,
)

// nolint: lll,gochecknoglobals
var supportedObjectsByDelete = datautils.NewStringSet(
	// https://v3.developer.constantcontact.com/api_reference/index.html#!/Contacts_Custom_Fields/deleteCustomField
	objectNameContactCustomFields,
	// https://v3.developer.constantcontact.com/api_reference/index.html#!/Contact_Lists/deleteListActivity
	objectNameContactLists,
	// https://v3.developer.constantcontact.com/api_reference/index.html#!/Contact_Tags/deleteTag
	objectNameContactTags,
	// https://v3.developer.constantcontact.com/api_reference/index.html#!/Contacts/deleteContact
	objectNameContacts,
	// https://v3.developer.constantcontact.com/api_reference/index.html#!/Email_Campaigns/removeEmailCampaignUsingDELETE
	objectNameEmailCampaigns,
	// https://v3.developer.constantcontact.com/api_reference/index.html#!/Segments/deleteSegment
	objectNameSegments,
	// https://v3.developer.constantcontact.com/api_reference/index.html#!/Technology_Partners_Webhooks/deleteWebhooksSubscriptions
	objectNameSubscriptions,
)

var objectNameToWritePath = map[string]string{ //nolint:gochecknoglobals
	objectNameAccountEmails:           "/account/emails",
	objectNameEmailCampaigns:          "/emails",
	objectNameEmailCampaignActivities: "/emails/activities",
}

var objectNameToWriteResponseIdentifier = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	objectNameAccounts:                "encoded_account_id",
	objectNameAccountEmails:           "email_id",
	objectNameContactLists:            "list_id",
	objectNameContactTags:             "tag_id",
	objectNameContactCustomFields:     "custom_field_id",
	objectNameEmailCampaigns:          "campaign_id",
	objectNameEmailCampaignActivities: "campaign_activity_id",
	objectNameSubscriptions:           "topic_id",
},
	func(objectName string) (id string) {
		return naming.NewSingularString(objectName).String() + "_id"
	},
)
