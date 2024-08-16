package providers

import "strings"

// Label keys.
const (
	LabelPrimaryCategory     = "primary-category"
	LabelSecondaryCategories = "secondary-categories"
)

// Label values (categories).
const (
	CategoryAdNetworks                  = "Ad Networks"
	CategoryAI                          = "AI"
	CategoryAnalytics                   = "Analytics"
	CategoryAnalyticsBI                 = "Analytics & BI"
	CategoryAppointmentScheduling       = "Appointment Scheduling"
	CategoryBilling                     = "Billing"
	CategoryCloudContentCollaboration   = "Cloud Content Collaboration"
	CategoryCodeManagement              = "Code Management"
	CategoryCommunication               = "Communication"
	CategoryContactCenter               = "Contact Center"
	CategoryContractLifecycleManagement = "Contract Lifecycle Management"
	CategoryConversationalPlatform      = "Conversational Platform"
	CategoryCPQ                         = "CPQ"
	CategoryCRM                         = "CRM"
	CategoryCustomerDataPlatform        = "Customer Data Platform"
	CategoryDigitalAssetManagement      = "Digital Asset Management"
	CategoryDigitalExperiencePlatform   = "Digital Experience Platform"
	CategoryECommerce                   = "E-Commerce"
	CategoryEmailMarketing              = "Email Marketing"
	CategoryEnrichment                  = "Enrichment"
	CategoryEnterpriseContentManagement = "Enterprise Content Management"
	CategoryESignature                  = "E-Signature"
	CategoryFieldService                = "Field Service"
	CategoryFinance                     = "Finance"
	CategoryFormBuilder                 = "Form Builder"
	CategoryHelpDesk                    = "Help Desk"
	CategoryITSM                        = "ITSM"
	CategoryKnowledgeBase               = "Knowledge Base"
	CategoryLeadCapture                 = "Lead Capture"
	CategoryLiveChat                    = "Live Chat"
	CategoryMarketingAutomation         = "Marketing Automation"
	CategoryMarketIntelligence          = "Market Intelligence"
	CategoryMiscellaneous               = "Miscellaneous"
	CategoryMobileMarketing             = "Mobile Marketing"
	CategoryOnlineReputationManagement  = "Online Reputation Management"
	CategoryOperations                  = "Operations"
	CategoryProductivityCollaboration   = "Productivity & Collaboration"
	CategoryProductManagement           = "Product Management"
	CategoryProjectManagement           = "Project Management"
	CategoryRevOps                      = "RevOps"
	CategorySalesEnablement             = "Sales Enablement"
	CategorySalesEngagement             = "Sales Engagement"
	CategorySalesIntelligence           = "Sales Intelligence"
	CategorySalesMiscellaneous          = "Sales Miscellaneous"
	CategoryScheduling                  = "Scheduling"
	CategoryService                     = "Service"
	CategorySharedInbox                 = "Shared Inbox"
	CategorySMSMarketing                = "SMS Marketing"
	CategorySocialMediaManagement       = "Social Media Management"
	CategorySoftwareDesignPlatform      = "Software Design Platform"
	CategorySubscriptionManagement      = "Subscription Management"
	CategoryTaskManagement              = "Task Management"
	CategoryTimeTracking                = "Time Tracking"
	CategoryUserGeneratedContent        = "User-Generated Content"
	CategoryVisualCollaboration         = "Visual Collaboration"
	CategoryWebsiteBuilder              = "Website Builder"
	CategoryWorkflowAutomation          = "Workflow Automation"
)

func list(labelValues ...string) string {
	return strings.Join(labelValues, "|")
}

func unlist(listString string) []string { // nolint:unused
	return strings.Split(listString, "|")
}
