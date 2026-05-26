package core

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

const (
	// DefaultPageSize is the default page size for paginated requests.
	// HubSpot's read endpoints support max 100 records per page.
	DefaultPageSize    = "100"
	DefaultPageSizeInt = int64(100)
)

const (
	ObjectMarketingCampaigns = "marketing-campaigns"
	ObjectMarketingEmails    = "marketing-emails"
	ObjectMarketingForms     = "marketing-forms"
	ObjectMarketingEvents    = "marketing-events"
	ObjectMeetingLinks       = "meeting-links"
	ObjectCustomChannels     = "custom-channels"
	ObjectChannelAccounts    = "channel-accounts"
	ObjectChannels           = "channels"
	ObjectInboxes            = "inboxes"
	ObjectThreads            = "threads"
)

//nolint:gochecknoglobals,lll
var (

	// CRMObjectsWithoutPropertiesAPISupport contains HubSpot CRM object names that
	// belong to the CRM API namespace but are not supported by the Object Properties API.
	//
	// These objects are not accessible through endpoints under either:
	//   /crm/v3/objects/{objectTypeId}/
	//	 /crm/objects/2026-03/{objectType}/{objectId}
	//
	// Objects that do support the Object Properties API are listed here:
	// https://developers.hubspot.com/docs/guides/api/crm/understanding-the-crm#object-type-ids
	CRMObjectsWithoutPropertiesAPISupport = datautils.NewSet( //nolint:gochecknoglobals
		"lists",
	)

	// MarketingObjects contains object names that belong to the HubSpot Marketing API.
	//
	// The Marketing API is separate from the CRM API and is not related to the Objects API.
	MarketingObjects = datautils.Map[string, ObjectDescription]{
		// https://developers.hubspot.com/docs/api-reference/latest/marketing/campaigns/get-campaigns
		ObjectMarketingCampaigns: {
			Path:              "campaigns",
			RecordTransformer: common.FlattenNestedFields("properties"),
			Version:           APIVersion2026March,
			PageSize:          DefaultPageSize,
		},
		// "marketing/emails" refers to HubSpot marketing emails, which are distinct
		// from the CRM email activity resource.
		//
		// The object name preserves the marketing-prefixed endpoint form to avoid a
		// naming collision with CRM emails.
		//
		// Path is relative to the Marketing API base path.
		//
		// Marketing emails:
		// https://developers.hubspot.com/docs/api-reference/latest/marketing/marketing-emails/get-emails
		// CRM emails:
		// https://developers.hubspot.com/docs/api-reference/latest/crm/activities/emails/guide
		ObjectMarketingEmails: {
			Path:              "emails",
			RecordTransformer: nil, // None. Fields and Raw are the same.
			Version:           APIVersion2026March,
			PageSize:          DefaultPageSize,
		},
		// https://developers.hubspot.com/docs/api-reference/2026-09-beta/marketing/forms/get-forms
		ObjectMarketingForms: {
			Path:              "forms",
			RecordTransformer: nil,
			Version:           APIVersion2026Sep,
			PageSize:          DefaultPageSize,
		},
		// https://developers.hubspot.com/docs/api-reference/latest/marketing/marketing-events/get-marketing-events
		ObjectMarketingEvents: {
			Path:              "marketing-events",
			RecordTransformer: nil,
			Version:           APIVersion2026March,
			PageSize:          DefaultPageSize,
		},
	}

	CommunicationObjects = datautils.Map[string, ObjectDescription]{
		// https://developers.hubspot.com/docs/api-reference/latest/conversations/custom-channels/channels/get-custom-channels
		// Requires scope: "conversations.custom_channels.read".
		// This is available only for Hubspot with "Sales Hub Professional" license.
		ObjectCustomChannels: {
			Path:              "custom-channels",
			RecordTransformer: nil,
			Version:           APIVersion2026March,
			PageSize:          DefaultPageSize,
		},
		// https://developers.hubspot.com/docs/api-reference/2026-09-beta/conversations/conversations/channel-accounts/get-channel-accounts
		ObjectChannelAccounts: {
			Path:              "channel-accounts",
			RecordTransformer: nil,
			Version:           APIVersion2026Sep,
			PageSize:          DefaultPageSize,
		},
		// https://developers.hubspot.com/docs/api-reference/2026-09-beta/conversations/conversations/channels/get-channels
		ObjectChannels: {
			Path:              "channels",
			RecordTransformer: nil,
			Version:           APIVersion2026Sep,
			PageSize:          DefaultPageSize,
		},
		// https://developers.hubspot.com/docs/api-reference/2026-09-beta/conversations/conversations/inbox/get-inboxes
		// Note: provider responds to request "/crm/v3/properties/inboxes" for ListObjectMetadata,
		// but that does not work for the read in ObjectsAPI.
		// The fields are not matching so it should be a different object.
		ObjectInboxes: {
			Path:              "inboxes",
			RecordTransformer: nil,
			Version:           APIVersion2026Sep,
			PageSize:          DefaultPageSize,
		},
		// https://developers.hubspot.com/docs/api-reference/2026-09-beta/conversations/conversations/threads/get-threads
		ObjectThreads: {
			Path:              "threads",
			RecordTransformer: nil,
			Version:           APIVersion2026Sep,
			PageSize:          DefaultPageSize,
		},
	}

	// MiscellaneousObjects are objects outside the CRM and Marketing APIs.
	MiscellaneousObjects = datautils.Map[string, ObjectDescription]{
		// https://api.hubapi.com/scheduler/2026-03/meetings/meeting-links
		ObjectMeetingLinks: {
			Path:              "/scheduler/2026-03/meetings/meeting-links",
			RecordTransformer: nil,
			Version:           "", // part of path.
			PageSize:          DefaultPageSize,
		},
		// https://developers.hubspot.com/docs/api-reference/latest/communication-preferences/guide?search=communication-preferences#get-all-subscription-types
		// Incremental read is not supported and even the page limit has no impact.
		// This endpoint is not well documented.
		"communication-preferences": {
			Path:              "/communication-preferences/2026-03/definitions",
			RecordTransformer: nil,
			Version:           "", // part of path.
			PageSize:          DefaultPageSize,
		},
	}
)

type ObjectDescription struct {
	// Path is URL path segment.
	Path string
	// RecordTransformer describes how to convert raw response and then extract selected fields by read operation.
	RecordTransformer common.RecordTransformer
	// Hubspot API Version.
	Version string
	// PageSize is maximum possible page limit for an object.
	PageSize string
}
