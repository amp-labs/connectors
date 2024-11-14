package klaviyo

import (
	"github.com/amp-labs/connectors/common"
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
		objectNameCampaigns,
		objectNameCampaignSendJobs,
		objectNameCampaignMessageAssignTemplate,
		objectNameCatalogVariants,
		objectNameCatalogVariantBulkCreateJobs,
		objectNameCatalogVariantBulkUpdateJobs,
		objectNameCatalogVariantBulkDeleteJobs,
		objectNameCatalogItems,
		objectNameCatalogItemBulkCreateJobs,
		objectNameCatalogItemBulkUpdateJobs,
		objectNameCatalogItemBulkDeleteJobs,
		objectNameBackInStockSubscriptions,
		objectNameCatalogCategories,
		objectNameCatalogCategoryBulkCreateJobs,
		objectNameCatalogCategoryBulkUpdateJobs,
		objectNameCatalogCategoryBulkDeleteJobs,
		objectNameCoupons,
		objectNameCouponCodes,
		objectNameCouponCodesBulkCreateJobs,
		objectNameDataPrivacyDeletionJobs,
		objectNameEvents,
		objectNameEventBulkCreateJobs,
		objectNameImages,
		objectNameImageUpload,
		objectNameLists,
		objectNameMetricAggregates,
		objectNameSuppressionBulkCreateJobs,
		objectNameSuppressionBulkDeleteJobs,
		objectNameSubscriptionBulkCreateJobs,
		objectNameSubscriptionBulkDeleteJobs,
		objectNameProfiles,
		objectNamePushTokens,
		objectNameProfileBulkImportJobs,
		objectNameSegments,
		objectNameTags,
		objectNameTagGroups,
		objectNameTemplateUniversalContent,
		objectNameTemplate,
		objectNameWebhooks,
	),
}

var supportedObjectsByUpdate = map[common.ModuleID]datautils.StringSet{ //nolint:gochecknoglobals
	Module2024Oct15: datautils.NewSet(
		objectNameCampaigns,
		objectNameCampaignSendJobs,
		objectNameCampaignMessages,
		objectNameCatalogVariants,
		objectNameCatalogItems,
		objectNameCatalogCategories,
		objectNameCoupons,
		objectNameCouponCodes,
		objectNameFlows,
		objectNameImages,
		objectNameLists,
		objectNameProfiles,
		objectNameSegments,
		objectNameTags,
		objectNameTagGroups,
		objectNameTemplateUniversalContent,
		objectNameTemplate,
		objectNameWebhooks,
	),
}

var supportedObjectsByDelete = map[common.ModuleID]datautils.StringSet{ //nolint:gochecknoglobals
	Module2024Oct15: datautils.NewSet(
		objectNameCampaigns,
		objectNameCatalogVariants,
		objectNameCatalogItems,
		objectNameCatalogCategories,
		objectNameCoupons,
		objectNameCouponCodes,
		objectNameFlows,
		objectNameLists,
		objectNameSegments,
		objectNameTags,
		objectNameTagGroups,
		objectNameTemplateUniversalContent,
		objectNameTemplate,
		objectNameWebhooks,
	),
}

var ObjectNameToWritePath = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
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
