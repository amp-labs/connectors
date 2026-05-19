package associations

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/hubspot/internal/core"
)

// CampaignAssetsFiller populates associations on `campaigns` rows by calling
// HubSpot's marketing campaigns assets endpoint:
//
//	GET /marketing/campaigns/2026-03/{campaignId}/assets/{assetType}
//
// Unlike CRM associations (which support a batch read API), HubSpot's
// campaign assets endpoint is per-campaign per-asset-type, so this filler
// fans out N campaigns × M asset types calls.
//
// https://developers.hubspot.com/docs/api-reference/latest/marketing/campaigns/assets/list-campaign-assets
type CampaignAssetsFiller struct {
	clientCRM    *common.JSONHTTPClient
	providerInfo *providers.ProviderInfo
}

// NewCampaignAssetsFiller constructs a filler for campaign asset associations.
func NewCampaignAssetsFiller(
	client *common.JSONHTTPClient, providerInfo *providers.ProviderInfo,
) *CampaignAssetsFiller {
	return &CampaignAssetsFiller{
		clientCRM:    client,
		providerInfo: providerInfo,
	}
}

// campaignAssetTypeMap translates Ampersand object names (the ones used in amp.yaml,
// e.g. "forms", "marketing-emails", "marketing-events") to the asset-type path
// segment expected by HubSpot's campaigns assets endpoint.
//
// HubSpot accepts uppercase singular asset type identifiers in the path.
//
//nolint:gochecknoglobals
var campaignAssetTypeMap = map[string]string{
	"forms":            "FORM",
	"marketing-emails": "MARKETING_EMAIL",
	"marketing-events": "MARKETING_EVENT",
	"ads":              "AD",
	"blog-posts":       "BLOG_POST",
	"landing-pages":    "LANDING_PAGE",
	"website-pages":    "WEBSITE_PAGE",
	"social-posts":     "SOCIAL_POST",
	"workflows":        "WORKFLOW",
	"sequences":        "SEQUENCE",
	"static-lists":     "STATIC_LIST",
	"ctas":             "CTA",
	"sms":              "SMS",
	"meetings":         "MEETING",
	"calls":            "CALL",
	"sales-emails":     "SALES_EMAIL",
}

// campaignAssetsResponse mirrors the shape of HubSpot's list-campaign-assets endpoint.
type campaignAssetsResponse struct {
	Results []campaignAssetItem `json:"results"`
}

type campaignAssetItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// FillAssociations implements Filler. fromObjName is expected to be "campaigns".
// Each entry in toAssociatedObjects is an Ampersand-style asset name (see
// campaignAssetTypeMap). Unknown asset names are skipped with a warning.
func (f *CampaignAssetsFiller) FillAssociations(
	ctx context.Context, fromObjName string, toAssociatedObjects []string,
	data []common.ReadResultRow,
) error {
	if fromObjName != "campaigns" {
		// Defensive: this filler is only intended for the campaigns object.
		return fmt.Errorf("%w: campaign assets filler invoked for object %q",
			common.ErrObjectNotSupported, fromObjName)
	}

	for _, assetName := range toAssociatedObjects {
		hubspotAssetType, ok := campaignAssetTypeMap[assetName]
		if !ok {
			logging.Logger(ctx).Warn(
				"skipping unsupported campaign asset type",
				"assetName", assetName,
			)

			continue
		}

		for i, row := range data {
			assets, err := f.fetchCampaignAssets(ctx, row.Id, hubspotAssetType)
			if err != nil {
				return err
			}

			if len(assets) == 0 {
				continue
			}

			if data[i].Associations == nil {
				data[i].Associations = make(map[string][]common.Association)
			}

			data[i].Associations[assetName] = assets
		}
	}

	return nil
}

// fetchCampaignAssets calls /marketing/campaigns/2026-03/{campaignId}/assets/{assetType}
// and returns one common.Association per asset returned.
func (f *CampaignAssetsFiller) fetchCampaignAssets(
	ctx context.Context, campaignID, assetType string,
) ([]common.Association, error) {
	url, err := urlbuilder.New(
		f.providerInfo.BaseURL,
		"marketing", "campaigns", core.APIVersion2026March,
		campaignID, "assets", assetType,
	)
	if err != nil {
		return nil, err
	}

	rsp, err := f.clientCRM.Get(ctx, url.String())
	if err != nil {
		var httpErr *common.HTTPError
		if errors.As(err, &httpErr) && httpErr.Status == http.StatusNotFound {
			// No assets of this type on this campaign; not an error.
			return nil, nil
		}

		return nil, fmt.Errorf("error fetching HubSpot campaign assets: %w", err)
	}

	parsed, err := common.UnmarshalJSON[campaignAssetsResponse](rsp)
	if err != nil {
		return nil, err
	}

	out := make([]common.Association, 0, len(parsed.Results))
	for _, item := range parsed.Results {
		assoc := common.Association{
			ObjectId: item.ID,
		}

		if item.Name != "" {
			assoc.ProviderAssociationMetadata = map[string]any{
				"name": strings.TrimSpace(item.Name),
			}
		}

		out = append(out, assoc)
	}

	return out, nil
}
