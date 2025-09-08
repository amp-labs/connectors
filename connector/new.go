// nolint:ireturn
package connector

import (
	"errors"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/aha"
	"github.com/amp-labs/connectors/providers/apollo"
	"github.com/amp-labs/connectors/providers/asana"
	"github.com/amp-labs/connectors/providers/ashby"
	"github.com/amp-labs/connectors/providers/atlassian"
	"github.com/amp-labs/connectors/providers/attio"
	"github.com/amp-labs/connectors/providers/avoma"
	"github.com/amp-labs/connectors/providers/aws"
	"github.com/amp-labs/connectors/providers/blueshift"
	"github.com/amp-labs/connectors/providers/braze"
	"github.com/amp-labs/connectors/providers/breakcold"
	"github.com/amp-labs/connectors/providers/brevo"
	"github.com/amp-labs/connectors/providers/campaignmonitor"
	"github.com/amp-labs/connectors/providers/capsule"
	"github.com/amp-labs/connectors/providers/chilipiper"
	"github.com/amp-labs/connectors/providers/claricopilot"
	"github.com/amp-labs/connectors/providers/clickup"
	"github.com/amp-labs/connectors/providers/closecrm"
	"github.com/amp-labs/connectors/providers/constantcontact"
	"github.com/amp-labs/connectors/providers/copper"
	"github.com/amp-labs/connectors/providers/customerapp"
	"github.com/amp-labs/connectors/providers/dixa"
	"github.com/amp-labs/connectors/providers/docusign"
	"github.com/amp-labs/connectors/providers/drift"
	"github.com/amp-labs/connectors/providers/dynamicsbusiness"
	"github.com/amp-labs/connectors/providers/dynamicscrm"
	"github.com/amp-labs/connectors/providers/fathom"
	"github.com/amp-labs/connectors/providers/fireflies"
	"github.com/amp-labs/connectors/providers/flatfile"
	"github.com/amp-labs/connectors/providers/freshdesk"
	"github.com/amp-labs/connectors/providers/front"
	"github.com/amp-labs/connectors/providers/github"
	"github.com/amp-labs/connectors/providers/gitlab"
	"github.com/amp-labs/connectors/providers/gong"
	"github.com/amp-labs/connectors/providers/google"
	"github.com/amp-labs/connectors/providers/gorgias"
	"github.com/amp-labs/connectors/providers/groove"
	"github.com/amp-labs/connectors/providers/helpscoutmailbox"
	"github.com/amp-labs/connectors/providers/heyreach"
	"github.com/amp-labs/connectors/providers/highlevelstandard"
	"github.com/amp-labs/connectors/providers/highlevelwhitelabel"
	"github.com/amp-labs/connectors/providers/hubspot"
	"github.com/amp-labs/connectors/providers/hunter"
	"github.com/amp-labs/connectors/providers/insightly"
	"github.com/amp-labs/connectors/providers/instantly"
	"github.com/amp-labs/connectors/providers/instantlyai"
	"github.com/amp-labs/connectors/providers/intercom"
	"github.com/amp-labs/connectors/providers/iterable"
	"github.com/amp-labs/connectors/providers/keap"
	"github.com/amp-labs/connectors/providers/kit"
	"github.com/amp-labs/connectors/providers/klaviyo"
	"github.com/amp-labs/connectors/providers/lemlist"
	"github.com/amp-labs/connectors/providers/lever"
	"github.com/amp-labs/connectors/providers/marketo"
	"github.com/amp-labs/connectors/providers/mixmax"
	"github.com/amp-labs/connectors/providers/monday"
	"github.com/amp-labs/connectors/providers/netsuite"
	"github.com/amp-labs/connectors/providers/outreach"
	"github.com/amp-labs/connectors/providers/pinterest"
	"github.com/amp-labs/connectors/providers/pipedrive"
	"github.com/amp-labs/connectors/providers/pipeliner"
	"github.com/amp-labs/connectors/providers/podium"
	"github.com/amp-labs/connectors/providers/pylon"
	"github.com/amp-labs/connectors/providers/salesforce"
	"github.com/amp-labs/connectors/providers/salesloft"
	"github.com/amp-labs/connectors/providers/seismic"
	"github.com/amp-labs/connectors/providers/servicenow"
	"github.com/amp-labs/connectors/providers/smartlead"
	"github.com/amp-labs/connectors/providers/stripe"
	"github.com/amp-labs/connectors/providers/teamleader"
	"github.com/amp-labs/connectors/providers/xero"
	"github.com/amp-labs/connectors/providers/zendeskchat"
	"github.com/amp-labs/connectors/providers/zendesksupport"
	"github.com/amp-labs/connectors/providers/zohocrm"
	"github.com/amp-labs/connectors/providers/zoom"
)

var ErrInvalidProvider = errors.New("invalid provider")

func New(provider providers.Provider, params common.ConnectorParams) (connectors.Connector, error) {
	constructor, ok := connectorConstructors[provider]
	if !ok {
		return nil, ErrInvalidProvider
	}

	return constructor(params)
}

var connectorConstructors = map[providers.Provider]outputConstructorFunc{ // nolint:gochecknoglobals
	providers.AWS:                     wrapper(newAWSConnector),
	providers.Aha:                     wrapper(newAhaConnector),
	providers.Apollo:                  wrapper(newApolloConnector),
	providers.Asana:                   wrapper(newAsanaConnector),
	providers.Ashby:                   wrapper(newAshbyConnector),
	providers.Atlassian:               wrapper(newAtlassianConnector),
	providers.Attio:                   wrapper(newAttioConnector),
	providers.Avoma:                   wrapper(newAvomaConnector),
	providers.Blueshift:               wrapper(newBlueshiftConnector),
	providers.Braze:                   wrapper(newBrazeConnector),
	providers.Breakcold:               wrapper(newBreakcoldConnector),
	providers.Brevo:                   wrapper(newBrevoConnector),
	providers.CampaignMonitor:         wrapper(newCampaignMonitorConnector),
	providers.Capsule:                 wrapper(newCapsuleConnector),
	providers.ChiliPiper:              wrapper(newChiliPiperConnector),
	providers.ClariCopilot:            wrapper(newClariCopilotConnector),
	providers.ClickUp:                 wrapper(newClickUpConnector),
	providers.Close:                   wrapper(newCloseConnector),
	providers.ConstantContact:         wrapper(newConstantContactConnector),
	providers.Copper:                  wrapper(newCopperConnector),
	providers.CustomerJourneysApp:     wrapper(newCustomerJourneysAppConnector),
	providers.Dixa:                    wrapper(newDixaConnector),
	providers.Docusign:                wrapper(newDocusignConnector),
	providers.Drift:                   wrapper(newDriftConnector),
	providers.DynamicsBusinessCentral: wrapper(newDynamicsBusinessCentral),
	providers.DynamicsCRM:             wrapper(newDynamicsCRMConnector),
	providers.Fathom:                  wrapper(newFathomConnector),
	providers.Fireflies:               wrapper(newFirefliesConnector),
	providers.Flatfile:                wrapper(newFlatfileConnector),
	providers.Freshdesk:               wrapper(newFreshdeskConnector),
	providers.Front:                   wrapper(newFrontConnector),
	providers.GitLab:                  wrapper(newGitLabConnector),
	providers.Github:                  wrapper(newGithubConnector),
	providers.Gong:                    wrapper(newGongConnector),
	providers.Google:                  wrapper(newGoogleConnector),
	providers.Gorgias:                 wrapper(newGorgiasConnector),
	providers.Groove:                  wrapper(newGrooveConnector),
	providers.HelpScoutMailbox:        wrapper(newHelpScoutMailboxConnector),
	providers.HeyReach:                wrapper(newHeyReachConnector),
	providers.HighLevelStandard:       wrapper(newHighLevelStandardConnector),
	providers.HighLevelWhiteLabel:     wrapper(newHighLevelWhiteLabelConnector),
	providers.Hubspot:                 wrapper(newHubspotConnector),
	providers.Hunter:                  wrapper(newHunterConnector),
	providers.Insightly:               wrapper(newInsightlyConnector),
	providers.Instantly:               wrapper(newInstantlyConnector),
	providers.InstantlyAI:             wrapper(newInstantlyAIConnector),
	providers.Intercom:                wrapper(newIntercomConnector),
	providers.Iterable:                wrapper(newIterableConnector),
	providers.Keap:                    wrapper(newKeapConnector),
	providers.Kit:                     wrapper(newKitConnector),
	providers.Klaviyo:                 wrapper(newKlaviyoConnector),
	providers.Lemlist:                 wrapper(newLemlistConnector),
	providers.Lever:                   wrapper(newLeverConnector),
	providers.Marketo:                 wrapper(newMarketoConnector),
	providers.Mixmax:                  wrapper(newMixmaxConnector),
	providers.Monday:                  wrapper(newMondayConnector),
	providers.Netsuite:                wrapper(newNetsuiteConnector),
	providers.Outreach:                wrapper(newOutreachConnector),
	providers.Pinterest:               wrapper(newPinterestConnector),
	providers.Pipedrive:               wrapper(newPipedriveConnector),
	providers.Pipeliner:               wrapper(newPipelinerConnector),
	providers.Podium:                  wrapper(newPodiumConnector),
	providers.Pylon:                   wrapper(newPylonConnector),
	providers.Salesforce:              wrapper(newSalesforceConnector),
	providers.Salesloft:               wrapper(newSalesloftConnector),
	providers.ServiceNow:              wrapper(newServiceNowConnector),
	providers.Seismic:                 wrapper(newSeismicConnector),
	providers.Smartlead:               wrapper(newSmartleadConnector),
	providers.Stripe:                  wrapper(newStripeConnector),
	providers.Teamleader:              wrapper(newTeamleaderConnector),
	providers.Xero:                    wrapper(newXeroConnector),
	providers.ZendeskChat:             wrapper(newZendeskChatConnector),
	providers.ZendeskSupport:          wrapper(newZendeskSupportConnector),
	providers.Zoho:                    wrapper(newZohoConnector),
	providers.Zoom:                    wrapper(newZoomConnector),
}

type outputConstructorFunc func(p common.ConnectorParams) (connectors.Connector, error)

type inputConstructorFunc[T connectors.Connector] func(p common.ConnectorParams) (T, error)

func wrapper[T connectors.Connector](input inputConstructorFunc[T]) outputConstructorFunc {
	return func(p common.ConnectorParams) (connectors.Connector, error) {
		return input(p)
	}
}

func newSalesforceConnector(params common.ConnectorParams) (*salesforce.Connector, error) {
	return salesforce.NewConnector(
		salesforce.WithAuthenticatedClient(params.AuthenticatedClient),
		salesforce.WithWorkspace(params.Workspace),
	)
}

func newHubspotConnector(params common.ConnectorParams) (*hubspot.Connector, error) {
	return hubspot.NewConnector(
		hubspot.WithAuthenticatedClient(params.AuthenticatedClient),
		hubspot.WithModule(params.Module),
	)
}

func newDocusignConnector(
	params common.ConnectorParams,
) (*docusign.Connector, error) {
	return docusign.NewConnector(
		docusign.WithAuthenticatedClient(params.AuthenticatedClient),
		docusign.WithMetadata(params.Metadata),
	)
}

func newIntercomConnector(
	params common.ConnectorParams,
) (*intercom.Connector, error) {
	return intercom.NewConnector(
		intercom.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newSalesloftConnector(
	params common.ConnectorParams,
) (*salesloft.Connector, error) {
	return salesloft.NewConnector(
		salesloft.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newDynamicsCRMConnector(
	params common.ConnectorParams,
) (*dynamicscrm.Connector, error) {
	return dynamicscrm.NewConnector(
		dynamicscrm.WithWorkspace(params.Workspace),
		dynamicscrm.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newOutreachConnector(
	params common.ConnectorParams,
) (*outreach.Connector, error) {
	return outreach.NewConnector(
		outreach.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newZendeskSupportConnector(
	params common.ConnectorParams,
) (*zendesksupport.Connector, error) {
	return zendesksupport.NewConnector(
		zendesksupport.WithWorkspace(params.Workspace),
		zendesksupport.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newAtlassianConnector(
	params common.ConnectorParams,
) (*atlassian.Connector, error) {
	return atlassian.NewConnector(
		atlassian.WithAuthenticatedClient(params.AuthenticatedClient),
		atlassian.WithModule(params.Module),
		atlassian.WithWorkspace(params.Workspace),
		atlassian.WithMetadata(params.Metadata),
	)
}

func newPipelinerConnector(
	params common.ConnectorParams,
) (*pipeliner.Connector, error) {
	return pipeliner.NewConnector(
		pipeliner.WithWorkspace(params.Workspace),
		pipeliner.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newSmartleadConnector(
	params common.ConnectorParams,
) (*smartlead.Connector, error) {
	return smartlead.NewConnector(
		smartlead.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newMarketoConnector(
	params common.ConnectorParams,
) (*marketo.Connector, error) {
	return marketo.NewConnector(
		marketo.WithWorkspace(params.Workspace),
		marketo.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newInstantlyConnector(
	params common.ConnectorParams,
) (*instantly.Connector, error) {
	return instantly.NewConnector(
		instantly.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newApolloConnector(
	params common.ConnectorParams,
) (*apollo.Connector, error) {
	return apollo.NewConnector(
		apollo.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newGongConnector(
	params common.ConnectorParams,
) (*gong.Connector, error) {
	return gong.NewConnector(
		gong.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newAttioConnector(
	params common.ConnectorParams,
) (*attio.Connector, error) {
	return attio.NewConnector(
		attio.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newPipedriveConnector(
	params common.ConnectorParams,
) (*pipedrive.Connector, error) {
	return pipedrive.NewConnector(
		pipedrive.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newZohoConnector(
	params common.ConnectorParams,
) (*zohocrm.Connector, error) {
	return zohocrm.NewConnector(
		zohocrm.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newClariCopilotConnector(
	params common.ConnectorParams,
) (*claricopilot.Connector, error) {
	return claricopilot.NewConnector(params)
}

func newCloseConnector(
	params common.ConnectorParams,
) (*closecrm.Connector, error) {
	return closecrm.NewConnector(
		closecrm.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newKlaviyoConnector(
	params common.ConnectorParams,
) (*klaviyo.Connector, error) {
	return klaviyo.NewConnector(
		klaviyo.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newCustomerJourneysAppConnector(
	params common.ConnectorParams,
) (*customerapp.Connector, error) {
	return customerapp.NewConnector(
		customerapp.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newConstantContactConnector(
	params common.ConnectorParams,
) (*constantcontact.Connector, error) {
	return constantcontact.NewConnector(
		constantcontact.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newCopperConnector(
	params common.ConnectorParams,
) (*copper.Connector, error) {
	return copper.NewConnector(params)
}

func newKeapConnector(
	params common.ConnectorParams,
) (*keap.Connector, error) {
	return keap.NewConnector(
		keap.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newKitConnector(
	params common.ConnectorParams,
) (*kit.Connector, error) {
	return kit.NewConnector(
		kit.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newIterableConnector(
	params common.ConnectorParams,
) (*iterable.Connector, error) {
	return iterable.NewConnector(
		iterable.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newAsanaConnector(
	params common.ConnectorParams,
) (*asana.Connector, error) {
	return asana.NewConnector(
		asana.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newStripeConnector(
	params common.ConnectorParams,
) (*stripe.Connector, error) {
	return stripe.NewConnector(
		stripe.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newZoomConnector(
	params common.ConnectorParams,
) (*zoom.Connector, error) {
	return zoom.NewConnector(
		zoom.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newBrevoConnector(
	params common.ConnectorParams,
) (*brevo.Connector, error) {
	return brevo.NewConnector(params)
}

func newBlueshiftConnector(
	params common.ConnectorParams,
) (*blueshift.Connector, error) {
	return blueshift.NewConnector(params)
}

func newAshbyConnector(
	params common.ConnectorParams,
) (*ashby.Connector, error) {
	return ashby.NewConnector(params)
}

func newGithubConnector(
	params common.ConnectorParams,
) (*github.Connector, error) {
	return github.NewConnector(params)
}

func newAhaConnector(
	params common.ConnectorParams,
) (*aha.Connector, error) {
	return aha.NewConnector(params)
}

func newClickUpConnector(
	params common.ConnectorParams,
) (*clickup.Connector, error) {
	return clickup.NewConnector(params)
}

func newMondayConnector(
	params common.ConnectorParams,
) (*monday.Connector, error) {
	return monday.NewConnector(params)
}

func newHeyReachConnector(
	params common.ConnectorParams,
) (*heyreach.Connector, error) {
	return heyreach.NewConnector(params)
}

func newAWSConnector(
	params common.ConnectorParams,
) (*aws.Connector, error) {
	return aws.NewConnector(params)
}

func newDriftConnector(
	params common.ConnectorParams,
) (*drift.Connector, error) {
	return drift.NewConnector(params)
}

func newDynamicsBusinessCentral(
	params common.ConnectorParams,
) (*dynamicsbusiness.Connector, error) {
	return dynamicsbusiness.NewConnector(params)
}

func newMixmaxConnector(
	params common.ConnectorParams,
) (*mixmax.Connector, error) {
	return mixmax.NewConnector(params)
}

func newDixaConnector(
	params common.ConnectorParams,
) (*dixa.Connector, error) {
	return dixa.NewConnector(params)
}

func newFlatfileConnector(
	params common.ConnectorParams,
) (*flatfile.Connector, error) {
	return flatfile.NewConnector(params)
}

func newFrontConnector(
	params common.ConnectorParams,
) (*front.Connector, error) {
	return front.NewConnector(params)
}

func newFreshdeskConnector(
	params common.ConnectorParams,
) (*freshdesk.Connector, error) {
	return freshdesk.NewConnector(
		freshdesk.WithAuthenticatedClient(params.AuthenticatedClient),
		freshdesk.WithWorkspace(params.Workspace),
	)
}

func newServiceNowConnector(
	params common.ConnectorParams,
) (*servicenow.Connector, error) {
	return servicenow.NewConnector(params)
}

func newChiliPiperConnector(
	params common.ConnectorParams,
) (*chilipiper.Connector, error) {
	return chilipiper.NewConnector(
		chilipiper.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newHunterConnector(
	params common.ConnectorParams,
) (*hunter.Connector, error) {
	return hunter.NewConnector(params)
}

func newInsightlyConnector(
	params common.ConnectorParams,
) (*insightly.Connector, error) {
	return insightly.NewConnector(params)
}

func newPodiumConnector(
	params common.ConnectorParams,
) (*podium.Connector, error) {
	return podium.NewConnector(params)
}

func newLemlistConnector(
	params common.ConnectorParams,
) (*lemlist.Connector, error) {
	return lemlist.NewConnector(params)
}

func newGorgiasConnector(
	params common.ConnectorParams,
) (*gorgias.Connector, error) {
	return gorgias.NewConnector(params)
}

func newZendeskChatConnector(
	params common.ConnectorParams,
) (*zendeskchat.Connector, error) {
	return zendeskchat.NewConnector(params)
}

func newCapsuleConnector(
	params common.ConnectorParams,
) (*capsule.Connector, error) {
	return capsule.NewConnector(params)
}

func newInstantlyAIConnector(
	params common.ConnectorParams,
) (*instantlyai.Connector, error) {
	return instantlyai.NewConnector(params)
}

func newGitLabConnector(
	params common.ConnectorParams,
) (*gitlab.Connector, error) {
	return gitlab.NewConnector(params)
}

func newHelpScoutMailboxConnector(
	params common.ConnectorParams,
) (*helpscoutmailbox.Connector, error) {
	return helpscoutmailbox.NewConnector(params)
}

func newGrooveConnector(
	params common.ConnectorParams,
) (*groove.Connector, error) {
	return groove.NewConnector(params)
}

func newPinterestConnector(
	params common.ConnectorParams,
) (*pinterest.Connector, error) {
	return pinterest.NewConnector(params)
}

func newAvomaConnector(
	params common.ConnectorParams,
) (*avoma.Connector, error) {
	return avoma.NewConnector(params)
}

func newFirefliesConnector(
	params common.ConnectorParams,
) (*fireflies.Connector, error) {
	return fireflies.NewConnector(params)
}

func newGoogleConnector(
	params common.ConnectorParams,
) (*google.Connector, error) {
	return google.NewConnector(params)
}

func newLeverConnector(
	params common.ConnectorParams,
) (*lever.Connector, error) {
	return lever.NewConnector(params)
}

func newBrazeConnector(
	params common.ConnectorParams,
) (*braze.Connector, error) {
	return braze.NewConnector(params)
}

func newFathomConnector(
	params common.ConnectorParams,
) (*fathom.Connector, error) {
	return fathom.NewConnector(params)
}

func newTeamleaderConnector(
	params common.ConnectorParams,
) (*teamleader.Connector, error) {
	return teamleader.NewConnector(params)
}

func newCampaignMonitorConnector(
	params common.ConnectorParams,
) (*campaignmonitor.Connector, error) {
	return campaignmonitor.NewConnector(params)
}

func newNetsuiteConnector(
	params common.ConnectorParams,
) (*netsuite.Connector, error) {
	return netsuite.NewConnector(params)
}

func newSeismicConnector(
	params common.ConnectorParams,
) (*seismic.Connector, error) {
	return seismic.NewConnector(params)
}

func newXeroConnector(
	params common.ConnectorParams,
) (*xero.Connector, error) {
	return xero.NewConnector(params)
}

func newBreakcoldConnector(
	params common.ConnectorParams,
) (*breakcold.Connector, error) {
	return breakcold.NewConnector(params)
}

func newPylonConnector(
	params common.ConnectorParams,
) (*pylon.Connector, error) {
	return pylon.NewConnector(params)
}

func newHighLevelStandardConnector(
	params common.ConnectorParams,
) (*highlevelstandard.Connector, error) {
	return highlevelstandard.NewConnector(params)
}

func newHighLevelWhiteLabelConnector(
	params common.ConnectorParams,
) (*highlevelwhitelabel.Connector, error) {
	return highlevelwhitelabel.NewConnector(params)
}
