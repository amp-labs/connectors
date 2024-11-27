package klaviyo

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/klaviyo/metadata"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals

var prioritySinceFieldsForRead = []string{ //nolint:gochecknoglobals
	"updated_at", // most desired field for incremental readin
	"updated",
	"datetime",
	"completed_at", // for Bulk Jobs
	"created_at",
	"created", // least preferred field
}

var objectsNameToSinceFieldName = make(map[common.ModuleID]map[string]string) //nolint:gochecknoglobals

func init() {
	// Every object should be associated with a fieldName which will be used for iterative reading via Since parameter.
	// Instead of recalculating this over and over on package start we can "find" fields of interest.
	// There is a preferred choice when it comes to time filtering represented by `prioritySinceFieldsForRead`.
	for moduleID, module := range metadata.Schemas.Modules {
		objectsNameToSinceFieldName[moduleID] = make(map[string]string)

		for objectName, object := range module.Objects {
		Search:
			for _, preferredSinceField := range prioritySinceFieldsForRead {
				for currentField := range object.FieldsMap {
					if preferredSinceField == currentField {
						objectsNameToSinceFieldName[moduleID][objectName] = preferredSinceField

						// break search for this object
						break Search
					}
				}
			}
		}
	}
}

// Reporting APIs is not added to the write method, this should be done via proxy.
// https://developers.klaviyo.com/en/reference/reporting_api_overview
// "tracking-settings" is ignored as a standalone object.
// https://developers.klaviyo.com/en/reference/update_tracking_setting
const (
	objectNameCampaigns                     = "campaigns"
	objectNameCampaignSendJobs              = "campaign-send-jobs"
	objectNameCampaignMessageAssignTemplate = "campaign-message-assign-template"
	objectNameCampaignMessages              = "campaign-messages"
	objectNameBackInStockSubscriptions      = "back-in-stock-subscriptions"
	objectNameDataPrivacyDeletionJobs       = "data-privacy-deletion-jobs"
	objectNameFlows                         = "flows"
	objectNameLists                         = "lists"
	objectNameMetricAggregates              = "metric-aggregates"
	objectNamePushTokens                    = "push-tokens"
	objectNameSegments                      = "segments"
	objectNameTags                          = "tags"
	objectNameTagGroups                     = "tag-groups"
	objectNameTemplateUniversalContent      = "template-universal-content"
	objectNameTemplate                      = "templates"
	objectNameWebhooks                      = "webhooks"
	// Catalog Variants have bulk jobs for each operation.
	objectNameCatalogVariants              = "catalog-variants"
	objectNameCatalogVariantBulkCreateJobs = "catalog-variant-bulk-create-jobs"
	objectNameCatalogVariantBulkUpdateJobs = "catalog-variant-bulk-update-jobs"
	objectNameCatalogVariantBulkDeleteJobs = "catalog-variant-bulk-delete-jobs"
	// Catalog Items have bulk jobs for each operation.
	objectNameCatalogItems              = "catalog-items"
	objectNameCatalogItemBulkCreateJobs = "catalog-item-bulk-create-jobs"
	objectNameCatalogItemBulkUpdateJobs = "catalog-item-bulk-update-jobs"
	objectNameCatalogItemBulkDeleteJobs = "catalog-item-bulk-delete-jobs"
	// Catalog Categories have bulk jobs for each operation.
	objectNameCatalogCategories             = "catalog-categories"
	objectNameCatalogCategoryBulkCreateJobs = "catalog-category-bulk-create-jobs"
	objectNameCatalogCategoryBulkUpdateJobs = "catalog-category-bulk-update-jobs"
	objectNameCatalogCategoryBulkDeleteJobs = "catalog-category-bulk-delete-jobs"
	// Client APIs have only POST actions, object names are artificially created and mapped to URLs.
	objectNameClientSubscriptions            = "client-subscriptions"
	objectNameClientPushTokens               = "client-push-tokens"
	objectNameClientPushTokensUnregister     = "client-push-token-unregister"
	objectNameClientEvents                   = "client-events"
	objectNameClientProfiles                 = "client-profiles"
	objectNameClientEventBulkCreate          = "client-event-bulk-create"
	objectNameClientBackInStockSubscriptions = "client-back-in-stock-subscriptions"
	// Coupons.
	objectNameCoupons                   = "coupons"
	objectNameCouponCodes               = "coupon-codes"
	objectNameCouponCodesBulkCreateJobs = "coupon-code-bulk-create-jobs"
	// Events.
	objectNameEvents              = "events"
	objectNameEventBulkCreateJobs = "event-bulk-create-jobs"
	// Images.
	objectNameImages      = "images"
	objectNameImageUpload = "image-upload"
	// Profile Suppression/Subscription jobs.
	objectNameSuppressionBulkCreateJobs  = "profile-suppression-bulk-create-jobs"
	objectNameSuppressionBulkDeleteJobs  = "profile-suppression-bulk-delete-jobs"
	objectNameSubscriptionBulkCreateJobs = "profile-subscription-bulk-create-jobs"
	objectNameSubscriptionBulkDeleteJobs = "profile-subscription-bulk-delete-jobs"
	// Profiles.
	objectNameProfiles              = "profiles"
	objectNameProfileBulkImportJobs = "profile-bulk-import-jobs"
)

var supportedObjectsByCreate = map[common.ModuleID]datautils.StringSet{ //nolint:gochecknoglobals
	Module2024Oct15: datautils.NewSet(
		// https://developers.klaviyo.com/en/reference/create_campaign
		objectNameCampaigns,
		// https://developers.klaviyo.com/en/reference/send_campaign
		objectNameCampaignSendJobs,
		// https://developers.klaviyo.com/en/reference/assign_template_to_campaign_message
		objectNameCampaignMessageAssignTemplate,
		// https://developers.klaviyo.com/en/reference/create_catalog_variant
		objectNameCatalogVariants,
		// https://developers.klaviyo.com/en/reference/bulk_create_catalog_variants
		objectNameCatalogVariantBulkCreateJobs,
		// https://developers.klaviyo.com/en/reference/bulk_update_catalog_variants
		objectNameCatalogVariantBulkUpdateJobs,
		// https://developers.klaviyo.com/en/reference/bulk_delete_catalog_variants
		objectNameCatalogVariantBulkDeleteJobs,
		// https://developers.klaviyo.com/en/reference/create_catalog_item
		objectNameCatalogItems,
		// https://developers.klaviyo.com/en/reference/bulk_create_catalog_items
		objectNameCatalogItemBulkCreateJobs,
		// https://developers.klaviyo.com/en/reference/bulk_update_catalog_items
		objectNameCatalogItemBulkUpdateJobs,
		// https://developers.klaviyo.com/en/reference/bulk_delete_catalog_items
		objectNameCatalogItemBulkDeleteJobs,
		// https://developers.klaviyo.com/en/reference/create_back_in_stock_subscription
		objectNameBackInStockSubscriptions,
		// https://developers.klaviyo.com/en/reference/create_catalog_category
		objectNameCatalogCategories,
		// https://developers.klaviyo.com/en/reference/bulk_create_catalog_categories
		objectNameCatalogCategoryBulkCreateJobs,
		// https://developers.klaviyo.com/en/reference/bulk_update_catalog_categories
		objectNameCatalogCategoryBulkUpdateJobs,
		// https://developers.klaviyo.com/en/reference/bulk_delete_catalog_categories
		objectNameCatalogCategoryBulkDeleteJobs,
		// https://developers.klaviyo.com/en/reference/create_coupon
		objectNameCoupons,
		// https://developers.klaviyo.com/en/reference/delete_coupon
		objectNameCouponCodes,
		// https://developers.klaviyo.com/en/reference/bulk_create_coupon_codes
		objectNameCouponCodesBulkCreateJobs,
		// https://developers.klaviyo.com/en/reference/request_profile_deletion
		objectNameDataPrivacyDeletionJobs,
		// https://developers.klaviyo.com/en/reference/create_event
		objectNameEvents,
		// https://developers.klaviyo.com/en/reference/bulk_create_events
		objectNameEventBulkCreateJobs,
		// https://developers.klaviyo.com/en/reference/upload_image_from_url
		objectNameImages,
		// https://developers.klaviyo.com/en/reference/upload_image_from_file
		objectNameImageUpload,
		// https://developers.klaviyo.com/en/reference/create_list
		objectNameLists,
		// https://developers.klaviyo.com/en/reference/query_metric_aggregates
		objectNameMetricAggregates,
		// https://developers.klaviyo.com/en/reference/bulk_suppress_profiles
		objectNameSuppressionBulkCreateJobs,
		// https://developers.klaviyo.com/en/reference/bulk_unsuppress_profiles
		objectNameSuppressionBulkDeleteJobs,
		// https://developers.klaviyo.com/en/reference/bulk_subscribe_profiles
		objectNameSubscriptionBulkCreateJobs,
		// https://developers.klaviyo.com/en/reference/bulk_unsubscribe_profiles
		objectNameSubscriptionBulkDeleteJobs,
		// https://developers.klaviyo.com/en/reference/create_profile
		objectNameProfiles,
		// https://developers.klaviyo.com/en/reference/create_push_token
		objectNamePushTokens,
		// https://developers.klaviyo.com/en/reference/spawn_bulk_profile_import_job
		objectNameProfileBulkImportJobs,
		// https://developers.klaviyo.com/en/reference/create_segment
		objectNameSegments,
		// https://developers.klaviyo.com/en/reference/create_tag
		objectNameTags,
		// https://developers.klaviyo.com/en/reference/create_tag_group
		objectNameTagGroups,
		// https://developers.klaviyo.com/en/reference/create_universal_content
		objectNameTemplateUniversalContent,
		// https://developers.klaviyo.com/en/reference/create_template
		objectNameTemplate,
		// https://developers.klaviyo.com/en/reference/create_webhook
		objectNameWebhooks,
		//
		// Client APIs have multiple POST actions.
		//
		// https://developers.klaviyo.com/en/reference/create_client_subscription
		objectNameClientSubscriptions,
		// https://developers.klaviyo.com/en/reference/create_client_push_token
		objectNameClientPushTokens,
		// https://developers.klaviyo.com/en/reference/unregister_client_push_token
		objectNameClientPushTokensUnregister,
		// https://developers.klaviyo.com/en/reference/create_client_event
		objectNameClientEvents,
		// https://developers.klaviyo.com/en/reference/create_client_profile
		objectNameClientProfiles,
		// https://developers.klaviyo.com/en/reference/bulk_create_client_events
		objectNameClientEventBulkCreate,
		// https://developers.klaviyo.com/en/reference/create_client_back_in_stock_subscription
		objectNameClientBackInStockSubscriptions,
	),
}

var supportedObjectsByUpdate = map[common.ModuleID]datautils.StringSet{ //nolint:gochecknoglobals
	Module2024Oct15: datautils.NewSet(
		// https://developers.klaviyo.com/en/reference/update_campaign
		objectNameCampaigns,
		// https://developers.klaviyo.com/en/reference/cancel_campaign_send
		objectNameCampaignSendJobs,
		// https://developers.klaviyo.com/en/reference/update_campaign_message
		objectNameCampaignMessages,
		// https://developers.klaviyo.com/en/reference/update_catalog_variant
		objectNameCatalogVariants,
		// https://developers.klaviyo.com/en/reference/update_catalog_item
		objectNameCatalogItems,
		// https://developers.klaviyo.com/en/reference/update_catalog_category
		objectNameCatalogCategories,
		// https://developers.klaviyo.com/en/reference/update_coupon
		objectNameCoupons,
		// https://developers.klaviyo.com/en/reference/update_coupon_code
		objectNameCouponCodes,
		// https://developers.klaviyo.com/en/reference/update_flow
		objectNameFlows,
		// https://developers.klaviyo.com/en/reference/update_image
		objectNameImages,
		// https://developers.klaviyo.com/en/reference/update_list
		objectNameLists,
		// https://developers.klaviyo.com/en/reference/update_profile
		objectNameProfiles,
		// https://developers.klaviyo.com/en/reference/update_segment
		objectNameSegments,
		// https://developers.klaviyo.com/en/reference/update_tag
		objectNameTags,
		// https://developers.klaviyo.com/en/reference/update_tag_group
		objectNameTagGroups,
		// https://developers.klaviyo.com/en/reference/update_universal_content
		objectNameTemplateUniversalContent,
		// https://developers.klaviyo.com/en/reference/update_template
		objectNameTemplate,
		// https://developers.klaviyo.com/en/reference/update_webhook
		objectNameWebhooks,
	),
}

var supportedObjectsByDelete = map[common.ModuleID]datautils.StringSet{ //nolint:gochecknoglobals
	Module2024Oct15: datautils.NewSet(
		// https://developers.klaviyo.com/en/reference/delete_campaign
		objectNameCampaigns,
		// https://developers.klaviyo.com/en/reference/delete_catalog_variant
		objectNameCatalogVariants,
		// https://developers.klaviyo.com/en/reference/delete_catalog_item
		objectNameCatalogItems,
		// https://developers.klaviyo.com/en/reference/delete_catalog_category
		objectNameCatalogCategories,
		// https://developers.klaviyo.com/en/reference/delete_coupon
		objectNameCoupons,
		// https://developers.klaviyo.com/en/reference/delete_coupon_code
		objectNameCouponCodes,
		// https://developers.klaviyo.com/en/reference/delete_flow
		objectNameFlows,
		// https://developers.klaviyo.com/en/reference/delete_list
		objectNameLists,
		// https://developers.klaviyo.com/en/reference/delete_segment
		objectNameSegments,
		// https://developers.klaviyo.com/en/reference/delete_tag
		objectNameTags,
		// https://developers.klaviyo.com/en/reference/delete_tag_group
		objectNameTagGroups,
		// https://developers.klaviyo.com/en/reference/delete_universal_content
		objectNameTemplateUniversalContent,
		// https://developers.klaviyo.com/en/reference/delete_template
		objectNameTemplate,
		// https://developers.klaviyo.com/en/reference/delete_webhook
		objectNameWebhooks,
	),
}

var objectNameToWritePath = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	objectNameClientSubscriptions:            "client/subscriptions",
	objectNameClientPushTokens:               "client/push-tokens",
	objectNameClientPushTokensUnregister:     "client/push-token-unregister",
	objectNameClientEvents:                   "client/events",
	objectNameClientProfiles:                 "client/profiles",
	objectNameClientEventBulkCreate:          "client/event-bulk-create",
	objectNameClientBackInStockSubscriptions: "client/back-in-stock-subscriptions",
},
	func(objectName string) (jsonPath string) {
		return "api/" + objectName
	},
)

// Write payload requires "type" property.
// This information can be inferred from ObjectName and client should not worry about it.
var objectNameToTypeWritePayload = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	objectNameCampaignMessageAssignTemplate:  "campaign-message",
	objectNameTemplateUniversalContent:       "template-universal-content",
	objectNameClientSubscriptions:            "subscription",
	objectNameClientPushTokens:               "push-token",
	objectNameClientPushTokensUnregister:     "push-token-unregister",
	objectNameClientEvents:                   "event",
	objectNameClientProfiles:                 "profile",
	objectNameClientEventBulkCreate:          "event-bulk-create",
	objectNameClientBackInStockSubscriptions: "back-in-stock-subscription",
},
	func(objectName string) (objectType string) {
		return naming.NewSingularString(objectName).String()
	},
)
