package hubspot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/simultaneously"
	"github.com/amp-labs/connectors/providers/hubspot/internal/associations"
	"github.com/amp-labs/connectors/providers/hubspot/internal/batch"
	"github.com/amp-labs/connectors/providers/hubspot/internal/core"
)

// Read reads data from Hubspot. If Since is set, it will use the
// ReadUsingSearchAPI endpoint instead to filter records, but it will be
// limited to a maximum of 10,000 records. This is a limit of the
// search endpoint. If Since is not set, it will use the read endpoint.
// In case Deleted objects won’t appear in any search results.
// Deleted objects can only be read by using this endpoint.
func (c *Connector) Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) { //nolint:funlen
	ctx = logging.With(ctx, "connector", "hubspot")

	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	//
	// Read using regular GET endpoints.
	//
	switch {
	case core.CRMObjectsWithoutPropertiesAPISupport.Has(params.ObjectName):
		// Object is part of CRM namespace but outside ObjectAPI.
		// For instance object "lists" is returned only via CRM Search endpoint.
		return c.searchCRM(ctx, searchCRMParams{
			SearchParams: SearchParams{
				ObjectName: params.ObjectName,
				NextPage:   params.NextPage,
				Fields:     params.Fields,
			},
		})
	case core.MarketingObjects.Has(params.ObjectName):
		// Object is part of Hubspot Marketing API.
		return c.readMarketing(ctx, params, core.MarketingObjects[params.ObjectName])
	case core.CommunicationObjects.Has(params.ObjectName):
		return c.readCommunications(ctx, params, core.CommunicationObjects[params.ObjectName])
	case core.MiscellaneousObjects.Has(params.ObjectName):
		return c.readMiscAPI(ctx, params, core.MiscellaneousObjects[params.ObjectName])
	default:
		// Otherwise object belongs to Hubspot Objects API (sub-category of CRM namespace).
		return c.readCRMObjectsAPI(ctx, params)
	}
}

// CRM objects can be read using two ways.
//   - If there are Since/Until parameters it will use:
//     https://api.hubapi.com/crm/objects/2026-03/{objectType}/search
//   - Otherwise, the Objects API endpoint is used:
//     https://api.hubapi.com/crm/objects/2026-03/{objectType}
func (c *Connector) readCRMObjectsAPI(
	ctx context.Context, params common.ReadParams,
) (*common.ReadResult, error) { //nolint:funlen
	// If filtering is required, then we have to use the search endpoint.
	// The Search endpoint has a 10K record limit. In case this limit is reached,
	// the sorting allows the caller to continue in another call by offsetting
	// until the ID of the last record that was successfully fetched.
	filters := make(Filters, 0)
	if !params.Since.IsZero() {
		filters = append(filters, BuildLastModifiedFilterGroup(&params))
	}

	if !params.Until.IsZero() {
		filters = append(filters, BuildUntilTimestampFilterGroup(&params))
	}

	filters = append(filters, BuildBuilderFilters(params.BuilderFilter)...)

	if len(filters) != 0 {
		searchParams := SearchParams{
			ObjectName: params.ObjectName,
			FilterGroups: []FilterGroup{{
				Filters: filters,
				// Add more filter groups to OR them together
			}},
			SortBy: []SortBy{
				BuildSort(ObjectFieldHsObjectId, SortDirectionAsc),
			},
			NextPage:          params.NextPage,
			Fields:            params.Fields,
			AssociatedObjects: params.AssociatedObjects,
		}

		return c.ReadUsingSearchAPI(ctx, searchParams)
	}

	url, err := c.buildCRMReadURL(params)
	if err != nil {
		return nil, err
	}

	rsp, err := c.JSONHTTPClient().Get(ctx, url)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		rsp,
		core.GetRecords,
		core.GetNextRecordsURL,
		associations.CreateDataMarshallerWithAssociations(
			ctx, c.associationsFiller, params.ObjectName, params.AssociatedObjects),
		params.Fields,
	)
}

func (c *Connector) buildCRMReadURL(params common.ReadParams) (string, error) {
	if len(params.NextPage) != 0 {
		// If NextPage is set, then we're reading the next page of results.
		// All that matters is the NextPage URL, the fields are ignored.
		return params.NextPage.String(), nil
	}

	// If NextPage is not set, then we're reading the first page of results.
	// We need to construct the query and then make the request.
	url, err := c.getCRMObjectsURL(params.ObjectName)
	if err != nil {
		return "", err
	}

	fields := params.Fields.List()
	if len(fields) != 0 {
		url.WithQueryParam("properties", strings.Join(fields, ","))
	}

	if params.Deleted {
		url.WithQueryParam("archived", "true")
	}

	url.WithQueryParam("limit", core.DefaultPageSize)

	return url.String(), nil
}

func (c *Connector) readMarketing(ctx context.Context,
	params common.ReadParams, object core.ObjectDescription,
) (*common.ReadResult, error) {
	requestedAssociations := datautils.NewSetFromList(params.AssociatedObjects)
	unsupportedAssociations := requestedAssociations.Subtract(object.Associations)

	if len(unsupportedAssociations) != 0 {
		return nil, fmt.Errorf("%w: associations %v",
			readhelper.ErrAssociationsUnsupported, strings.Join(unsupportedAssociations, ","),
		)
	}

	url, err := c.buildMarketingReadURL(params, &object)
	if err != nil {
		return nil, err
	}

	resp, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	identifier := "id"
	if params.ObjectName == core.ObjectMarketingEvents {
		identifier = "objectId"
	}

	marshaler := readhelper.MakeMarshaledDataFuncWithId(
		object.RecordTransformer,
		readhelper.IdFieldQuery{Field: identifier},
	)

	if params.ObjectName == core.ObjectMarketingCampaigns {
		marshaler = readhelper.ChainedMarshaller(
			// Process campaigns normally.
			readhelper.MakeMarshaledDataFuncWithId(
				object.RecordTransformer,
				readhelper.IdFieldQuery{Field: identifier},
			),
			// Enhance marketing campaigns with associations.
			readhelper.HydrateAssociations(ctx,
				core.ObjectMarketingCampaigns, params.AssociatedObjects,
				c.lookupMarketingCampaignAssociations,
			),
		)
	}

	return common.ParseResultFiltered(
		params,
		resp,
		common.MakeRecordsFunc("results"),
		makeIncrementalFilterFunc(params),
		marshaler,
		params.Fields,
	)
}

// When reading objects in Hubspot you must explicitly request the fields.
// https://developers.hubspot.com/docs/api-reference/latest/marketing/campaigns/guide#campaign-properties
//
// Reading campaigns object:
// https://developers.hubspot.com/docs/api-reference/latest/marketing/campaigns/get-campaigns
//   - Incremental reading is not available.
//   - Sorting is applied using "updatedAt" field from newest to oldest.
func (c *Connector) buildMarketingReadURL(
	params common.ReadParams, object *core.ObjectDescription,
) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		// Next page
		return urlbuilder.New(params.NextPage.String())
	}

	// First page
	url, err := c.getMarketingURL(object)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("limit", readhelper.PageSizeWithDefaultStr(params, core.DefaultPageSize))

	if params.ObjectName == core.ObjectMarketingForms || params.ObjectName == core.ObjectMeetingLinks {
		// This object does not have such query params. For consistency, it is reflected here.
		// Sending non-existent query params is not considered an error by provider.
	} else {
		url.WithQueryParam("properties", strings.Join(params.Fields.List(), ","))
		url.WithQueryParam("sort", "-updatedAt") // newest first
	}

	return url, nil
}

// makeIncrementalFilterFunc embodies connector-side filtering.
// ReverseOrder is used because we request Campaigns sorted from newest to oldest.
func makeIncrementalFilterFunc(params common.ReadParams) common.RecordsFilterFunc {
	if params.Since.IsZero() && params.Until.IsZero() {
		return readhelper.MakeIdentityFilterFunc(core.GetNextRecordsURL)
	}

	order := readhelper.ReverseOrder
	if params.ObjectName == core.ObjectMarketingForms {
		order = readhelper.Unordered
	}

	return readhelper.MakeTimeFilterFunc(
		order,
		readhelper.NewTimeBoundary(),
		"updatedAt", time.RFC3339,
		core.GetNextRecordsURL,
	)
}

func (c *Connector) readCommunications(ctx context.Context, // nolint:funlen
	params common.ReadParams, object core.ObjectDescription,
) (*common.ReadResult, error) {
	var (
		url *urlbuilder.URL
		err error
	)

	if params.IsFirstPage() {
		url, err = c.getCommunicationURL(params.ObjectName, &object)
	} else {
		url, err = urlbuilder.New(params.NextPage.String())
	}

	if err != nil {
		return nil, err
	}

	url.WithQueryParam("limit", readhelper.PageSizeWithDefaultStr(params, object.PageSize))

	// Prepare the URL query params and pick provider side filtering method.
	var filter common.RecordsFilterFunc

	switch params.ObjectName {
	case core.ObjectCustomChannels:
		// This object cannot be tested. The way it works is an assumption.
		filter = readhelper.MakeTimeFilterFunc(
			readhelper.Unordered, readhelper.NewTimeBoundary(),
			"createdAt", time.RFC3339, core.GetNextRecordsURL,
		)
	case core.ObjectChannels:
		filter = readhelper.MakeIdentityFilterFunc(core.GetNextRecordsURL)
	case core.ObjectInboxes:
		url.WithQueryParam("sort", "-updatedAt") // newest first

		filter = readhelper.MakeTimeFilterFunc(
			readhelper.ReverseOrder, readhelper.NewTimeBoundary(),
			"updatedAt", time.RFC3339, core.GetNextRecordsURL,
		)
	case core.ObjectChannelAccounts:
		url.WithQueryParam("sort", "-createdAt") // newest first

		filter = readhelper.MakeTimeFilterFunc(
			readhelper.ReverseOrder, readhelper.NewTimeBoundary(),
			"createdAt", time.RFC3339, core.GetNextRecordsURL,
		)
	case core.ObjectThreads:
		filter = readhelper.MakeTimeFilterFunc(
			readhelper.Unordered, readhelper.NewTimeBoundary(),
			"createdAt", time.RFC3339, core.GetNextRecordsURL,
		)
	default:
		return nil, common.ErrObjectNotSupported
	}

	resp, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResultFiltered(
		params,
		resp,
		common.MakeRecordsFunc("results"),
		filter,
		readhelper.MakeMarshaledDataFuncWithId(
			object.RecordTransformer,
			readhelper.IdFieldQuery{Field: "id"},
		),
		params.Fields,
	)
}

func (c *Connector) readMiscAPI(ctx context.Context,
	params common.ReadParams, object core.ObjectDescription,
) (*common.ReadResult, error) {
	url, err := c.buildMiscURL(params, &object)
	if err != nil {
		return nil, err
	}

	resp, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResultFiltered(
		params,
		resp,
		common.MakeRecordsFunc("results"),
		readhelper.MakeTimeFilterFunc(
			readhelper.Unordered,
			readhelper.NewTimeBoundary(),
			"updatedAt", time.RFC3339,
			core.GetNextRecordsURL,
		),
		readhelper.MakeMarshaledDataFuncWithId(
			object.RecordTransformer,
			readhelper.IdFieldQuery{Field: "id"},
		),
		params.Fields,
	)
}

func (c *Connector) buildMiscURL(
	params common.ReadParams, object *core.ObjectDescription,
) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		// Next page
		return urlbuilder.New(params.NextPage.String())
	}

	// First page
	url, err := c.rootURL(object.Path)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("limit", readhelper.PageSizeWithDefaultStr(params, object.PageSize))

	return url, nil
}

func (c *Connector) lookupMarketingCampaignAssociations(ctx context.Context,
	fromObject common.ObjectName,
	identifiers []readhelper.RowID,
	toObject string,
) (map[readhelper.RowID][]common.Association, error) {
	if fromObject != core.ObjectMarketingCampaigns {
		// Marketing campaigns is the only object supported by this method.
		return nil, readhelper.ErrAssociationLookupNotImplemented
	}

	if len(identifiers) == 0 {
		// Early exit.
		return make(map[readhelper.RowID][]common.Association), nil
	}

	switch toObject {
	case core.AssociationAssets:
		return c.lookupMarketingCampaignAssets(ctx, identifiers)
	case core.AssociationContacts:
		return c.lookupMarketingCampaignContacts(ctx, identifiers)
	default:
		return nil, readhelper.ErrAssociationsUnsupported
	}
}

// lookupMarketingCampaignAssets performs a batch read of marketing campaigns and
// resolves all assets associated with each campaign.
//
// The provider API returns campaign assets grouped by asset type
// (for example MARKETING_EMAIL, MARKETING_EVENT, or AD_CAMPAIGN).
// Each asset type is paginated independently and the batch endpoint only returns
// the first page, capped at 50 assets per type.
//
// This method:
//
//  1. Batch reads campaigns together with the first page of their assets.
//  2. Detects asset types that have additional pages.
//  3. Fetches all remaining pages for every paginated asset type.
//  4. Converts the final asset collection into Ampersand associations.
//
// Returned associations preserve the provider asset type in
// ProviderAssociationMetadata["assetType"].
func (c *Connector) lookupMarketingCampaignAssets(ctx context.Context,
	identifiers []string,
) (map[readhelper.RowID][]common.Association, error) {
	batchResult := batch.Read[marketingCampaignSchema](ctx, c.batchAdapter, batch.ReadParams{
		ObjectName:  core.ObjectMarketingCampaigns,
		Identifiers: identifiers,
	})

	if len(batchResult.Errors) != 0 {
		return nil, errors.Join(batchResult.Errors...)
	}

	assetBundles := convertBatchCampaignsToAssetBundles(batchResult.Records)

	// Fetch all remaining paginated assets and extend the asset registries stored inside the bundles.
	if err := c.fetchRemainingAssets(ctx, assetBundles); err != nil {
		return nil, err
	}

	// Convert collected assets into associations keyed by campaign ID.
	registry := map[readhelper.RowID][]common.Association{}

	for _, bundle := range assetBundles {
		campaignID := readhelper.RowID(bundle.ID)
		registry[campaignID] = make([]common.Association, bundle.Assets.CombinedLength())
		index := 0

		// Every asset becomes a distinct association.
		// The original provider asset type is preserved as metadata.
		for assetKind, assets := range bundle.Assets {
			for _, asset := range assets {
				registry[campaignID][index] = common.Association{
					ObjectId: asset.Identifier(),
					Raw:      asset,
					ProviderAssociationMetadata: map[string]any{
						"assetType": assetKind,
					},
				}
				index += 1
			}
		}
	}

	return registry, nil
}

// fetchRemainingAssets resolves all paginated asset collections for the provided campaigns.
//
// The initial campaign batch read only contains the first page of assets for each asset type.
// Pagination is independent per asset type, therefore each asset category must be traversed
// separately until exhaustion.
//
// The method mutates the provided bundles in place by appending newly fetched assets into bundle.Assets.
// Pagination continues until the provider stops returning a paging cursor.
func (c *Connector) fetchRemainingAssets(ctx context.Context, bundles []assetBundle) error {
	for _, bundle := range bundles {
		for assetKind, after := range bundle.NextPages {
			url, err := c.getMarketingCampaignAssetsURL(bundle.ID, assetKind)
			if err != nil {
				return err
			}

			// Traverse all pages for this campaign + asset type pair.
			for after != "" {
				// https://developers.hubspot.com/docs/api-reference/latest/marketing/campaigns/assets/get-assets
				url.WithQueryParam("after", after)
				delete(bundle.NextPages, assetKind) // symbolic bookkeeping

				resp, err := c.JSONHTTPClient().Get(ctx, url.String())
				if err != nil {
					return err
				}

				assets, err := common.UnmarshalJSON[assetsResponse](resp)
				if err != nil {
					return err
				}

				// Append fetched assets into the existing registry.
				for _, asset := range assets.Results {
					bundle.Assets.Add(assetKind, asset)
				}

				// Continue pagination if another page exists.
				if assets.Paging != nil {
					after = assets.Paging.Next.After

					continue
				}

				// No paging token means this was the final page.
				break
			}
		}
	}

	return nil
}

type marketingCampaignSchema struct {
	ID           CampaignID           `json:"id"`
	AssetMapping campaignAssetMapping `json:"assets"`
}

type (
	// campaignAssetMapping groups assets by provider asset type.
	//
	// Example keys:
	//   - MARKETING_EMAIL
	//   - MARKETING_EVENT
	//   - AD_CAMPAIGN
	campaignAssetMapping map[assetType]assetsResponse
	assetType            = string
	assetsResponse       struct {
		Results []assetDef `json:"results"`
		Paging  *struct {
			Next struct {
				After string `json:"after"`
				Link  string `json:"link"`
			} `json:"next"`
		} `json:"paging,omitempty"`
	}
	assetDef map[string]any
)

func (a assetDef) Identifier() string {
	id, _ := a["id"].(string)

	return id
}

// assetBundle is an intermediate aggregation structure used while resolving
// campaign assets across paginated provider responses.
//
// The batch campaign endpoint only returns the first page of assets for each
// asset type. assetBundle stores:
//
//   - all currently collected assets grouped by asset type
//   - pagination cursors for asset types that have more than one page
//
// The bundle acts as the bridge between the provider API representation and
// the final Ampersand association model.
type assetBundle struct {
	// Campaign identifier owning the assets.
	ID CampaignID
	// Fully or partially collected Assets grouped by asset type.
	Assets datautils.IndexedLists[assetType, assetDef]
	// NextPages holds "after" page token keyed by asset type.
	// Presence in this map indicates that additional pages must still be fetched.
	NextPages map[assetType]string
}

// convertBatchCampaignsToAssetBundles transforms batch campaign responses into
// mutable asset bundles suitable for pagination expansion.
func convertBatchCampaignsToAssetBundles(campaigns []marketingCampaignSchema) []assetBundle {
	bundles := make([]assetBundle, len(campaigns))

	for index, campaign := range campaigns {
		// Init bundle.
		bundle := assetBundle{
			ID:        campaign.ID,
			Assets:    make(datautils.IndexedLists[assetType, assetDef]),
			NextPages: make(map[assetType]string),
		}
		bundles[index] = bundle

		// Populate bundle.
		for assetKind, assetWrapper := range campaign.AssetMapping {
			// Store the first page of assets already returned by the batch API.
			for _, asset := range assetWrapper.Results {
				bundle.Assets.Add(assetKind, asset)
			}

			// Record pagination state for asset types that require additional fetches.
			if assetWrapper.Paging != nil {
				bundle.NextPages[assetKind] = assetWrapper.Paging.Next.After
			}
		}
	}

	return bundles
}

type crmObjectSchema[T ~string] struct {
	// ID which can be found inside the Data.
	ID T
	// Data is the raw JSON.
	Data map[string]any
}

func (c *crmObjectSchema[T]) UnmarshalJSON(bytes []byte) error {
	type essentials struct {
		ID T `json:"id"`
	}

	var essentialData essentials
	if err := json.Unmarshal(bytes, &essentialData); err != nil {
		return err
	}

	c.ID = essentialData.ID

	var everything map[string]any
	if err := json.Unmarshal(bytes, &everything); err != nil {
		return err
	}

	c.Data = everything

	return nil
}

func (c *Connector) lookupMarketingCampaignContacts(ctx context.Context, // nolint:funlen
	campaignIdentifiers []string,
) (map[string][]common.Association, error) {
	var (
		contactTypes    = []string{"contactFirstTouch", "contactLastTouch", "influencedContacts"}
		numRequests     = len(contactTypes) * len(campaignIdentifiers)
		responseChannel = make(chan relationshipCampaignToContacts, numRequests)
		callbacks       = make([]simultaneously.Job, numRequests)
		index           = 0
	)

	// Collect contacts for each existing contact type aka relationship from campaign to contacts.
	for _, contactType := range contactTypes {
		// There is no batch endpoint, get contact ids for each campaign instance.
		for _, identifier := range campaignIdentifiers {
			callbacks[index] = func(ctx context.Context) error {
				return c.fetchMarketingCampaignContactIdentifiers(ctx, identifier, contactType, responseChannel)
			}
			index += 1
		}
	}

	// Wait for all routines.
	if err := simultaneously.DoCtx(ctx, -1, callbacks...); err != nil {
		close(responseChannel)

		return nil, err
	}

	close(responseChannel)

	// Collect all contact identifiers to fetch full contact information.
	// Create the lookup to match the contacts back to the campaigns.
	type contactRelationship struct {
		ContactType string
		CampaignID  CampaignID
	}

	campaignLookup := make(map[ContactID]contactRelationship)
	contactIDs := make(datautils.Set[string])

	for relationship := range responseChannel {
		for _, contactID := range relationship.ContactIDs {
			campaignLookup[contactID] = contactRelationship{
				ContactType: relationship.ContactType,
				CampaignID:  relationship.CampaignID,
			}
			contactIDs.AddOne(string(contactID))
		}
	}

	// Fetch contacts.
	contactBatchResult := batch.Read[crmObjectSchema[ContactID]](ctx, c.batchAdapter, batch.ReadParams{
		ObjectName:  core.ObjectContacts,
		Identifiers: contactIDs.List(),
	})

	if len(contactBatchResult.Errors) != 0 {
		return nil, errors.Join(contactBatchResult.Errors...)
	}

	// Create and fill in associations.
	registry := datautils.IndexedLists[string, common.Association]{}

	for _, contact := range contactBatchResult.Records {
		relationship := campaignLookup[contact.ID]
		registry.Add(string(relationship.CampaignID), common.Association{
			ObjectId: string(contact.ID),
			Raw:      contact.Data,
			ProviderAssociationMetadata: map[string]any{
				"associationType": relationship.ContactType,
			},
		})
	}

	return registry, nil
}

func (c *Connector) fetchMarketingCampaignContactIdentifiers(ctx context.Context,
	campaignIdentifier string,
	contactType string,
	outbox chan<- relationshipCampaignToContacts,
) error {
	url, err := c.getMarketingCampaignContactsURL(campaignIdentifier, contactType)
	if err != nil {
		return err
	}

	resp, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return err
	}

	collection, err := common.UnmarshalJSON[identifierCollection](resp)
	if err != nil {
		return err
	}

	outbox <- relationshipCampaignToContacts{
		CampaignID:  CampaignID(campaignIdentifier),
		ContactType: contactType,
		ContactIDs:  collection.IDs(),
	}

	return nil
}

// identifierCollection holds contact identifiers associated with marketing campaign.
// The contacts response for the following types is the same: contactFirstTouch, contactLastTouch, influencedContacts.
// https://developers.hubspot.com/docs/api-reference/latest/marketing/campaigns/reports/get-contact-ids
type identifierCollection struct {
	Results []struct {
		ID ContactID `json:"id"`
	} `json:"results"`
}

func (c identifierCollection) IDs() []ContactID {
	identifiers := make([]ContactID, len(c.Results))
	for index, result := range c.Results {
		identifiers[index] = result.ID
	}

	return identifiers
}

type (
	CampaignID string
	ContactID  string
)

type relationshipCampaignToContacts struct {
	CampaignID  CampaignID
	ContactType string
	ContactIDs  []ContactID
}
