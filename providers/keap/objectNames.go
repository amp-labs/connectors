package keap

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/keap/metadata"
)

const (
	objectNameAffiliates           = "affiliates"
	objectNameAppointments         = "appointments"
	objectNameAutomationCategories = "automationCategories"
	objectNameCompanies            = "companies"
	objectNameContacts             = "contacts"
	objectNameOrders               = "orders"
	objectNameSubscriptions        = "subscriptions"
	objectNameEmails               = "emails"
	objectNameFiles                = "files"
	objectNameNotes                = "notes"
	objectNameOpportunities        = "opportunities"
	objectNamePaymentMethodConfigs = "paymentMethodConfigs"
	objectNameProducts             = "products"
	objectNameHooks                = "hooks"
	objectNameTags                 = "tags"
	objectNameTagCategories        = "tag_categories"
	objectNameTasks                = "tasks"
	objectNameUsers                = "users"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals

var supportedObjectsByCreate = map[common.ModuleID]datautils.StringSet{ //nolint:gochecknoglobals
	ModuleV1: datautils.NewSet(
		// https://developer.infusionsoft.com/docs/rest/#tag/Affiliate/operation/createAffiliateUsingPOST
		objectNameAffiliates,
		// https://developer.infusionsoft.com/docs/rest/#tag/Appointment/operation/createAppointmentUsingPOST
		objectNameAppointments,
		// https://developer.infusionsoft.com/docs/rest/#tag/Company/operation/createCompanyUsingPOST
		objectNameCompanies,
		// https://developer.infusionsoft.com/docs/rest/#tag/Contact/operation/createContactUsingPOST
		objectNameContacts,
		// https://developer.infusionsoft.com/docs/rest/#tag/E-Commerce/operation/createOrderUsingPOST
		objectNameOrders,
		// https://developer.infusionsoft.com/docs/rest/#tag/E-Commerce/operation/createSubscriptionUsingPOST
		objectNameSubscriptions,
		// https://developer.infusionsoft.com/docs/rest/#tag/Email/operation/createEmailUsingPOST
		objectNameEmails,
		// https://developer.infusionsoft.com/docs/rest/#tag/File/operation/createFileUsingPOST
		objectNameFiles,
		// https://developer.infusionsoft.com/docs/rest/#tag/Note/operation/createNoteUsingPOST
		objectNameNotes,
		// https://developer.infusionsoft.com/docs/rest/#tag/Opportunity/operation/createOpportunityUsingPOST
		objectNameOpportunities,
		// https://developer.infusionsoft.com/docs/rest/#tag/Product/operation/createProductUsingPOST
		objectNameProducts,
		// https://developer.infusionsoft.com/docs/rest/#tag/REST-Hooks/operation/create_a_hook_subscription
		objectNameHooks,
		// https://developer.infusionsoft.com/docs/rest/#tag/Tags/operation/createTagUsingPOST
		objectNameTags,
		// https://developer.infusionsoft.com/docs/rest/#tag/Tags/operation/createTagCategoryUsingPOST
		objectNameTagCategories,
		// https://developer.infusionsoft.com/docs/rest/#tag/Task/operation/createTaskUsingPOST
		objectNameTasks,
		// https://developer.infusionsoft.com/docs/rest/#tag/Users/operation/createUserUsingPOST
		objectNameUsers,
	),
	ModuleV2: datautils.NewSet(
		// https://developer.keap.com/docs/restv2/#tag/Affiliate/operation/addAffiliateUsingPOST
		objectNameAffiliates,
		// https://developer.keap.com/docs/restv2/#tag/AutomationCategory/operation/createCategoryUsingPOST
		objectNameAutomationCategories,
		// https://developer.keap.com/docs/restv2/#tag/Company/operation/createCompanyUsingPOST_1
		objectNameCompanies,
		// https://developer.keap.com/docs/restv2/#tag/Contact/operation/createContactUsingPOST_1
		objectNameContacts,
		// https://developer.keap.com/docs/restv2/#tag/Email/operation/createEmailUsingPOST_1
		objectNameEmails,
		// https://developer.keap.com/docs/restv2/#tag/PaymentMethodConfig/operation/createPaymentMethodConfigUsingPOST
		objectNamePaymentMethodConfigs,
		// https://developer.keap.com/docs/restv2/#tag/Subscription-Plans/operation/createSubscriptionV2UsingPOST
		objectNameSubscriptions,
		// https://developer.keap.com/docs/restv2/#tag/Tags/operation/createTagUsingPOST_1
		objectNameTags,
		// https://developer.keap.com/docs/restv2/#tag/Tags/operation/createTagCategoryUsingPOST_1
		objectNameTagCategories,
	),
}

// Evert update performed using PATCH is not present in the PUT set.
var supportedObjectsByUpdatePATCH = map[common.ModuleID]datautils.StringSet{ //nolint:gochecknoglobals
	ModuleV1: datautils.NewSet(
		// https://developer.infusionsoft.com/docs/rest/#tag/Appointment/operation/updatePropertiesOnAppointmentUsingPATCH
		objectNameAppointments,
		// https://developer.infusionsoft.com/docs/rest/#tag/Company/operation/updateCompanyUsingPATCH
		objectNameCompanies,
		// https://developer.infusionsoft.com/docs/rest/#tag/Contact/operation/updatePropertiesOnContactUsingPATCH
		objectNameContacts,
		// https://developer.infusionsoft.com/docs/rest/#tag/Note/operation/updatePropertiesOnNoteUsingPATCH
		objectNameNotes,
		// https://developer.infusionsoft.com/docs/rest/#tag/Opportunity/operation/updatePropertiesOnOpportunityUsingPATCH
		objectNameOpportunities,
		// https://developer.infusionsoft.com/docs/rest/#tag/Product/operation/updateProductUsingPATCH
		objectNameProducts,
		// https://developer.infusionsoft.com/docs/rest/#tag/Task/operation/updatePropertiesOnTaskUsingPATCH
		objectNameTasks,
	),
	ModuleV2: datautils.NewSet(
		// https://developer.keap.com/docs/restv2/#tag/Affiliate/operation/updateAffiliateUsingPATCH
		objectNameAffiliates,
		// https://developer.keap.com/docs/restv2/#tag/Company/operation/patchCompanyUsingPATCH
		objectNameCompanies,
		// https://developer.keap.com/docs/restv2/#tag/Contact/operation/patchContactUsingPATCH
		objectNameContacts,
		// https://developer.keap.com/docs/restv2/#tag/Tags/operation/patchTagUsingPATCH
		objectNameTags,
		// https://developer.keap.com/docs/restv2/#tag/Tags/operation/patchTagCategoryUsingPATCH
		objectNameTagCategories,
	),
}

// Evert update performed using PUT is not present in the PATCH set.
var supportedObjectsByUpdatePUT = map[common.ModuleID]datautils.StringSet{ //nolint:gochecknoglobals
	ModuleV1: datautils.NewSet(
		// https://developer.infusionsoft.com/docs/rest/#tag/File/operation/updateFileUsingPUT
		objectNameFiles,
		// https://developer.infusionsoft.com/docs/rest/#tag/REST-Hooks/operation/update_a_hook_subscription
		objectNameHooks,
	),
	ModuleV2: datautils.NewSet(
		// https://developer.keap.com/docs/restv2/#tag/AutomationCategory/operation/saveCategoryUsingPUT
		objectNameAutomationCategories,
	),
}

var supportedObjectsByDelete = map[common.ModuleID]datautils.StringSet{ //nolint:gochecknoglobals
	ModuleV1: datautils.NewSet(
		// https://developer.infusionsoft.com/docs/rest/#tag/Appointment/operation/deleteAppointmentUsingDELETE
		objectNameAppointments,
		// https://developer.keap.com/docs/rest/#tag/Contact/operation/deleteContactUsingDELETE
		objectNameContacts,
		// https://developer.infusionsoft.com/docs/rest/#tag/E-Commerce/operation/deleteOrderUsingDELETE
		objectNameOrders,
		// https://developer.infusionsoft.com/docs/rest/#tag/Email/operation/deleteEmailUsingDELETE
		objectNameEmails,
		// https://developer.infusionsoft.com/docs/rest/#tag/File/operation/deleteFileUsingDELETE
		objectNameFiles,
		// https://developer.infusionsoft.com/docs/rest/#tag/Note/operation/deleteNoteUsingDELETE
		objectNameNotes,
		// https://developer.infusionsoft.com/docs/rest/#tag/Product/operation/deleteProductUsingDELETE
		objectNameProducts,
		// https://developer.infusionsoft.com/docs/rest/#tag/REST-Hooks/operation/delete_a_hook_subscription
		objectNameHooks,
		// https://developer.infusionsoft.com/docs/rest/#tag/Task/operation/deleteTaskUsingDELETE
		objectNameTasks,
	),
	ModuleV2: datautils.NewSet(
		// https://developer.keap.com/docs/restv2/#tag/AutomationCategory/operation/deleteCategoriesUsingDELETE
		objectNameAutomationCategories,
		// https://developer.keap.com/docs/restv2/#tag/Company/operation/deleteCompanyUsingDELETE
		objectNameCompanies,
		// https://developer.keap.com/docs/restv2/#tag/Contact/operation/deleteContactUsingDELETE_1
		objectNameContacts,
		// https://developer.keap.com/docs/restv2/#tag/Email/operation/deleteEmailUsingDELETE_1
		objectNameEmails,
		// https://developer.keap.com/docs/restv2/#tag/Tags/operation/deleteTagCategoryUsingDELETE
		objectNameTagCategories,
	),
}

// objectNameToWritePath maps ObjectName to URL path used for Write operation.
//
// Some of the ignored endpoints:
// "/v1/account/profile" -- update single profile
// "/v1/emails/queue" -- send email to list of contacts
// "/v1/emails/unsync" -- un-sync a batch of email records
// "/v1/emails/sync" -- create a set of email records
// "/v2/businessProfile" -- update single profile.
var objectNameToWritePath = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	objectNameTagCategories:        "tags/categories",
	objectNameAutomationCategories: "automationCategory", // API uses singular form. Others are consistently plural.
},
	func(objectName string) (jsonPath string) {
		return objectName
	},
)

var objectNameToWriteResponseIdentifier = common.ModuleObjectNameToFieldName{ //nolint:gochecknoglobals
	ModuleV1: datautils.NewDefaultMap(map[string]string{
		objectNameFiles: "id",

		objectNameHooks: "key",
	},
		func(objectName string) (id string) {
			return "id"
		},
	),
	ModuleV2: datautils.NewDefaultMap(map[string]string{
		objectNamePaymentMethodConfigs: "session_key",
	},
		func(objectName string) (id string) {
			return "id"
		},
	),
}

var objectsWithCustomFields = map[common.ModuleID]datautils.StringSet{ // nolint:gochecknoglobals
	ModuleV1: datautils.NewStringSet(
		objectNameAffiliates,
		objectNameAppointments,
		objectNameCompanies,
		objectNameContacts,
		objectNameNotes,
		objectNameOpportunities,
		objectNameOrders,
		objectNameSubscriptions,
		objectNameTasks,
	),
	ModuleV2: datautils.NewStringSet(
		objectNameAffiliates,
		objectNameContacts,
		objectNameNotes,
		objectNameOrders,
		objectNameSubscriptions,
		objectNameTasks,
	),
}
