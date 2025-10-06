package facebook

import (
	"fmt"

	"github.com/amp-labs/connectors/internal/datautils"
)

func (c *Adapter) constructURL(objName string) string {
	if EndpointWithAdAccountIdInURLPath.Has(objName) {
		return fmt.Sprintf("act_%s/%s", c.adAccountId, objName)
	}

	if EndpointWithBusineesIdInURLPath.Has(objName) {
		return fmt.Sprintf("%s/%s", c.businessId, objName)
	}

	return ""
}

var EndpointWithAdAccountIdInURLPath = datautils.NewSet( //nolint:gochecknoglobals
	"users",
	"ad_place_page_sets",
	"adrules_library",
	"adplayables",
	"adlabels",
	"adimages",
	"account_controls",
	"ads",
	"adsets",
	"advertisable_applications",
	"advideos",
	"agencies",
	"applications",
	"broadtargetingcategories",
	"customaudiencestos",
	"customconversions",
	"deprecatedtargetingadsets",
	"dsa_recommendations",
	"impacting_ad_studies",
	"minimum_budgets",
	"promote_pages",
	"publisher_block_lists",
	"reachfrequencypredictions",
	"saved_audiences",
	"subscribed_apps",
	"targetingbrowse",
	"tracking",
	"adcreatives",
	"campaigns",
	"customaudiences",
)

var EndpointWithBusineesIdInURLPath = datautils.NewSet( //nolint:gochecknoglobals
	"ad_studies",
	"adnetworkanalytics",
	"adnetworkanalytics_results",
	"adspixels",
	"agencies",
	"business_invoices",
	"business_users",
	"client_pages",
	"client_pixels",
	"client_product_catalogs",
	"client_whatsapp_business_accounts",
	"clients",
	"collaborative_ads_collaboration_requests",
	"collaborative_ads_suggested_partners",
	"event_source_groups",
	"extendedcredits",
	"initiated_audience_sharing_requests",
	"instagram_accounts",
	"managed_partner_ads_funding_source_details",
	"owned_apps",
	"owned_businesses",
	"owned_pages",
	"owned_pixels",
	"owned_product_catalogs",
	"owned_whatsapp_business_accounts",
	"pending_client_ad_accounts",
	"pending_client_apps",
	"pending_client_pages",
	"pending_owned_ad_accounts",
	"pending_owned_pages",
	"pending_users",
	"received_audience_sharing_requests",
	"system_users",
)
