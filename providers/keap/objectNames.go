package keap

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/keap/metadata"
)

const (
	// Version 1
	// https://developer.keap.com/docs/rest/
	objectNameAffiliates    = "v1/affiliates"
	objectNameAppointments  = "v1/appointments"
	objectNameCompanies     = "v1/companies"
	objectNameContacts      = "v1/contacts"
	objectNameOrders        = "v1/orders"
	objectNameSubscriptions = "v1/subscriptions"
	objectNameEmails        = "v1/emails"
	objectNameFiles         = "v1/files"
	objectNameNotes         = "v1/notes"
	objectNameOpportunities = "v1/opportunities"
	objectNameProducts      = "v1/products"
	objectNameHooks         = "v1/hooks"
	objectNameTags          = "v1/tags"
	objectNameTagCategories = "v1/tags/categories"
	objectNameTasks         = "v1/tasks"
	objectNameUsers         = "v1/users"

	// Version 2
	// https://developer.keap.com/docs/restv2/
	objectNameAffiliatesV2           = "v2/affiliates"
	objectNameAutomationCategoriesV2 = "v2/automationCategory"
	objectNameCompaniesV2            = "v2/companies"
	objectNameContactsV2             = "v2/contacts"
	objectNameSubscriptionsV2        = "v2/subscriptions"
	objectNameEmailsV2               = "v2/emails"
	objectNamePaymentMethodConfigsV2 = "v2/paymentMethodConfigs"
	objectNameTagsV2                 = "v2/tags"
	objectNameTagCategoriesV2        = "v2/tags/categories"
	objectNameTasksV2                = "v2/tasks"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals

var supportedObjectsByCreate = map[common.ModuleID]datautils.StringSet{ //nolint:gochecknoglobals
	common.ModuleRoot: datautils.NewSet(
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
		// https://developer.keap.com/docs/restv2/#tag/Affiliate/operation/addAffiliateUsingPOST
		objectNameAffiliatesV2,
		// https://developer.keap.com/docs/restv2/#tag/AutomationCategory/operation/createCategoryUsingPOST
		objectNameAutomationCategoriesV2,
		// https://developer.keap.com/docs/restv2/#tag/Company/operation/createCompanyUsingPOST_1
		objectNameCompaniesV2,
		// https://developer.keap.com/docs/restv2/#tag/Contact/operation/createContactUsingPOST_1
		objectNameContactsV2,
		// https://developer.keap.com/docs/restv2/#tag/Email/operation/createEmailUsingPOST_1
		objectNameEmailsV2,
		// https://developer.keap.com/docs/restv2/#tag/PaymentMethodConfig/operation/createPaymentMethodConfigUsingPOST
		objectNamePaymentMethodConfigsV2,
		// https://developer.keap.com/docs/restv2/#tag/Subscription-Plans/operation/createSubscriptionV2UsingPOST
		objectNameSubscriptionsV2,
		// https://developer.keap.com/docs/restv2/#tag/Tags/operation/createTagUsingPOST_1
		objectNameTagsV2,
		// https://developer.keap.com/docs/restv2/#tag/Tags/operation/createTagCategoryUsingPOST_1
		objectNameTagCategoriesV2,
		// https://developer.keap.com/docs/restv2/#tag/Task/operation/createTaskUsingPOST_1
		objectNameTasksV2,
	),
}

// Every update performed using PATCH is not present in the PUT set.
var supportedObjectsByUpdatePATCH = map[common.ModuleID]datautils.StringSet{ //nolint:gochecknoglobals
	common.ModuleRoot: datautils.NewSet(
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
		// https://developer.keap.com/docs/restv2/#tag/Affiliate/operation/updateAffiliateUsingPATCH
		objectNameAffiliatesV2,
		// https://developer.keap.com/docs/restv2/#tag/Company/operation/patchCompanyUsingPATCH
		objectNameCompaniesV2,
		// https://developer.keap.com/docs/restv2/#tag/Contact/operation/patchContactUsingPATCH
		objectNameContactsV2,
		// https://developer.keap.com/docs/restv2/#tag/Tags/operation/patchTagUsingPATCH
		objectNameTagsV2,
		// https://developer.keap.com/docs/restv2/#tag/Tags/operation/patchTagCategoryUsingPATCH
		objectNameTagCategoriesV2,
		// https://developer.keap.com/docs/restv2/#tag/Task/operation/updateTaskUsingPATCH
		objectNameTasksV2,
	),
}

// Every update performed using PUT is not present in the PATCH set.
var supportedObjectsByUpdatePUT = map[common.ModuleID]datautils.StringSet{ //nolint:gochecknoglobals
	common.ModuleRoot: datautils.NewSet(
		// https://developer.infusionsoft.com/docs/rest/#tag/File/operation/updateFileUsingPUT
		objectNameFiles,
		// https://developer.infusionsoft.com/docs/rest/#tag/REST-Hooks/operation/update_a_hook_subscription
		objectNameHooks,
		// https://developer.keap.com/docs/restv2/#tag/AutomationCategory/operation/saveCategoryUsingPUT
		objectNameAutomationCategoriesV2,
	),
}

var supportedObjectsByDelete = map[common.ModuleID]datautils.StringSet{ //nolint:gochecknoglobals
	common.ModuleRoot: datautils.NewSet(
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
		// https://developer.keap.com/docs/restv2/#tag/AutomationCategory/operation/deleteCategoriesUsingDELETE
		objectNameAutomationCategoriesV2,
		// https://developer.keap.com/docs/restv2/#tag/Company/operation/deleteCompanyUsingDELETE
		objectNameCompaniesV2,
		// https://developer.keap.com/docs/restv2/#tag/Contact/operation/deleteContactUsingDELETE_1
		objectNameContactsV2,
		// https://developer.keap.com/docs/restv2/#tag/Email/operation/deleteEmailUsingDELETE_1
		objectNameEmailsV2,
		// https://developer.keap.com/docs/restv2/#tag/Tags/operation/deleteTagCategoryUsingDELETE
		objectNameTagCategoriesV2,
		// https://developer.keap.com/docs/restv2/#tag/Task/operation/deleteTaskUsingDELETE_1
		objectNameTasksV2,
	),
}

var objectNameToWriteResponseIdentifier = common.ModuleObjectNameToFieldName{ //nolint:gochecknoglobals
	common.ModuleRoot: datautils.NewDefaultMap(map[string]string{
		objectNameFiles:                  "id",
		objectNameHooks:                  "key",
		objectNamePaymentMethodConfigsV2: "session_key",
	},
		func(objectName string) (id string) {
			return "id"
		},
	),
}

var objectsWithCustomFields = map[common.ModuleID]datautils.StringSet{ // nolint:gochecknoglobals
	common.ModuleRoot: datautils.NewStringSet(
		objectNameAffiliates,
		objectNameAppointments,
		objectNameCompanies,
		objectNameContacts,
		objectNameNotes,
		objectNameOpportunities,
		objectNameOrders,
		objectNameSubscriptions,
		objectNameTasks,
		objectNameAffiliatesV2,
		objectNameContactsV2,
		objectNameSubscriptionsV2,
		objectNameTasksV2,
	),
}
