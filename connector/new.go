package connector

import (
	"errors"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/aha"
	"github.com/amp-labs/connectors/providers/apollo"
	"github.com/amp-labs/connectors/providers/asana"
	"github.com/amp-labs/connectors/providers/ashby"
	"github.com/amp-labs/connectors/providers/atlassian"
	"github.com/amp-labs/connectors/providers/attio"
	"github.com/amp-labs/connectors/providers/aws"
	"github.com/amp-labs/connectors/providers/blueshift"
	"github.com/amp-labs/connectors/providers/brevo"
	"github.com/amp-labs/connectors/providers/capsule"
	"github.com/amp-labs/connectors/providers/chilipiper"
	"github.com/amp-labs/connectors/providers/clickup"
	"github.com/amp-labs/connectors/providers/closecrm"
	"github.com/amp-labs/connectors/providers/constantcontact"
	"github.com/amp-labs/connectors/providers/customerapp"
	"github.com/amp-labs/connectors/providers/dixa"
	"github.com/amp-labs/connectors/providers/docusign"
	"github.com/amp-labs/connectors/providers/drift"
	"github.com/amp-labs/connectors/providers/dynamicscrm"
	"github.com/amp-labs/connectors/providers/freshdesk"
	"github.com/amp-labs/connectors/providers/front"
	"github.com/amp-labs/connectors/providers/github"
	"github.com/amp-labs/connectors/providers/gong"
	"github.com/amp-labs/connectors/providers/gorgias"
	"github.com/amp-labs/connectors/providers/heyreach"
	"github.com/amp-labs/connectors/providers/hubspot"
	"github.com/amp-labs/connectors/providers/hunter"
	"github.com/amp-labs/connectors/providers/instantly"
	"github.com/amp-labs/connectors/providers/instantlyai"
	"github.com/amp-labs/connectors/providers/intercom"
	"github.com/amp-labs/connectors/providers/iterable"
	"github.com/amp-labs/connectors/providers/keap"
	"github.com/amp-labs/connectors/providers/kit"
	"github.com/amp-labs/connectors/providers/klaviyo"
	"github.com/amp-labs/connectors/providers/lemlist"
	"github.com/amp-labs/connectors/providers/marketo"
	"github.com/amp-labs/connectors/providers/mixmax"
	"github.com/amp-labs/connectors/providers/monday"
	"github.com/amp-labs/connectors/providers/outreach"
	"github.com/amp-labs/connectors/providers/pipedrive"
	"github.com/amp-labs/connectors/providers/pipeliner"
	"github.com/amp-labs/connectors/providers/podium"
	"github.com/amp-labs/connectors/providers/salesforce"
	"github.com/amp-labs/connectors/providers/salesloft"
	"github.com/amp-labs/connectors/providers/servicenow"
	"github.com/amp-labs/connectors/providers/smartlead"
	"github.com/amp-labs/connectors/providers/stripe"
	"github.com/amp-labs/connectors/providers/zendeskchat"
	"github.com/amp-labs/connectors/providers/zendesksupport"
	"github.com/amp-labs/connectors/providers/zohocrm"
	"github.com/amp-labs/connectors/providers/zoom"
)

var ErrInvalidProvider = errors.New("invalid provider")

// New returns a Connector instance for the given provider, initialized with the provided Parameters.
//
// This acts as a factory function that chooses appropriate connector constructor based on the provider argument.
// The returned connector will be configured according to the parameters supplied.
//
// If the provider is unrecognized, New returns ErrInvalidProvider.
func New( // nolint:gocyclo,cyclop,funlen,ireturn
	provider providers.Provider,
	params Parameters,
) (connectors.Connector, error) {
	var (
		connector    connectors.Connector
		connectorErr error
	)

	switch provider {
	case providers.Hubspot:
		connector, connectorErr = newHubspotConnector(params)
	case providers.Salesforce:
		connector, connectorErr = newSalesforceConnector(params)
	case providers.Docusign:
		connector, connectorErr = newDocusignConnector(params)
	case providers.Intercom:
		connector, connectorErr = newIntercomConnector(params)
	case providers.Salesloft:
		connector, connectorErr = newSalesloftConnector(params)
	case providers.DynamicsCRM:
		connector, connectorErr = newDynamicsCRMConnector(params)
	case providers.ZendeskSupport:
		connector, connectorErr = newZendeskSupportConnector(params)
	case providers.Outreach:
		connector, connectorErr = newOutreachConnector(params)
	case providers.Atlassian:
		connector, connectorErr = newAtlassianConnector(params)
	case providers.Pipeliner:
		connector, connectorErr = newPipelinerConnector(params)
	case providers.Smartlead:
		connector, connectorErr = newSmartleadConnector(params)
	case providers.Marketo:
		connector, connectorErr = newMarketoConnector(params)
	case providers.Instantly:
		connector, connectorErr = newInstantlyConnector(params)
	case providers.Apollo:
		connector, connectorErr = newApolloConnector(params)
	case providers.Gong:
		connector, connectorErr = newGongConnector(params)
	case providers.Attio:
		connector, connectorErr = newAttioConnector(params)
	case providers.Pipedrive:
		connector, connectorErr = newPipedriveConnector(params)
	case providers.Zoho:
		connector, connectorErr = newZohoConnector(params)
	case providers.Close:
		connector, connectorErr = newCloseConnector(params)
	case providers.Klaviyo:
		connector, connectorErr = newKlaviyoConnector(params)
	case providers.CustomerJourneysApp:
		connector, connectorErr = newCustomerJourneysAppConnector(params)
	case providers.ConstantContact:
		connector, connectorErr = newConstantContactConnector(params)
	case providers.Keap:
		connector, connectorErr = newKeapConnector(params)
	case providers.Kit:
		connector, connectorErr = newKitConnector(params)
	case providers.Iterable:
		connector, connectorErr = newIterableConnector(params)
	case providers.Asana:
		connector, connectorErr = newAsanaConnector(params)
	case providers.Stripe:
		connector, connectorErr = newStripeConnector(params)
	case providers.Zoom:
		connector, connectorErr = newZoomConnector(params)
	case providers.Brevo:
		connector, connectorErr = newBrevoConnector(params)
	case providers.Blueshift:
		connector, connectorErr = newBlueshiftConnector(params)
	case providers.Ashby:
		connector, connectorErr = newAshbyConnector(params)
	case providers.Github:
		connector, connectorErr = newGithubConnector(params)
	case providers.Aha:
		connector, connectorErr = newAhaConnector(params)
	case providers.ClickUp:
		connector, connectorErr = newClickUpConnector(params)
	case providers.Monday:
		connector, connectorErr = newMondayConnector(params)
	case providers.HeyReach:
		connector, connectorErr = newHeyReachConnector(params)
	case providers.AWS:
		connector, connectorErr = newAWSConnector(params)
	case providers.Drift:
		connector, connectorErr = newDriftConnector(params)
	case providers.Mixmax:
		connector, connectorErr = newMixmaxConnector(params)
	case providers.Dixa:
		connector, connectorErr = newDixaConnector(params)
	case providers.Front:
		connector, connectorErr = newFrontConnector(params)
	case providers.Freshdesk:
		connector, connectorErr = newFreshdeskConnector(params)
	case providers.ServiceNow:
		connector, connectorErr = newServiceNowConnector(params)
	case providers.ChiliPiper:
		connector, connectorErr = newChiliPiperConnector(params)
	case providers.Hunter:
		connector, connectorErr = newHunterConnector(params)
	case providers.Podium:
		connector, connectorErr = newPodiumConnector(params)
	case providers.Lemlist:
		connector, connectorErr = newLemlistConnector(params)
	case providers.Gorgias:
		connector, connectorErr = newGorgiasConnector(params)
	case providers.ZendeskChat:
		connector, connectorErr = newZendeskChatConnector(params)
	case providers.Capsule:
		connector, connectorErr = newCapsuleConnector(params)
	case providers.InstantlyAI:
		connector, connectorErr = newInstantlyAIConnector(params)
	default:
		return nil, ErrInvalidProvider
	}

	return connector, connectorErr
}

func newSalesforceConnector(params Parameters) (*salesforce.Connector, error) {
	return salesforce.NewConnector(
		salesforce.WithAuthenticatedClient(params.AuthenticatedClient),
		salesforce.WithWorkspace(params.Workspace),
	)
}

func newHubspotConnector(params Parameters) (*hubspot.Connector, error) {
	return hubspot.NewConnector(
		hubspot.WithAuthenticatedClient(params.AuthenticatedClient),
		hubspot.WithModule(params.Module),
	)
}

func newDocusignConnector(
	params Parameters,
) (*docusign.Connector, error) {
	return docusign.NewConnector(
		docusign.WithAuthenticatedClient(params.AuthenticatedClient),
		docusign.WithMetadata(params.Metadata),
	)
}

func newIntercomConnector(
	params Parameters,
) (*intercom.Connector, error) {
	return intercom.NewConnector(
		intercom.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newSalesloftConnector(
	params Parameters,
) (*salesloft.Connector, error) {
	return salesloft.NewConnector(
		salesloft.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newDynamicsCRMConnector(
	params Parameters,
) (*dynamicscrm.Connector, error) {
	return dynamicscrm.NewConnector(
		dynamicscrm.WithWorkspace(params.Workspace),
		dynamicscrm.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newOutreachConnector(
	params Parameters,
) (*outreach.Connector, error) {
	return outreach.NewConnector(
		outreach.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newZendeskSupportConnector(
	params Parameters,
) (*zendesksupport.Connector, error) {
	return zendesksupport.NewConnector(
		zendesksupport.WithWorkspace(params.Workspace),
		zendesksupport.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newAtlassianConnector(
	params Parameters,
) (*atlassian.Connector, error) {
	return atlassian.NewConnector(
		atlassian.WithAuthenticatedClient(params.AuthenticatedClient),
		atlassian.WithModule(params.Module),
		atlassian.WithWorkspace(params.Workspace),
		atlassian.WithMetadata(params.Metadata),
	)
}

func newPipelinerConnector(
	params Parameters,
) (*pipeliner.Connector, error) {
	return pipeliner.NewConnector(
		pipeliner.WithWorkspace(params.Workspace),
		pipeliner.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newSmartleadConnector(
	params Parameters,
) (*smartlead.Connector, error) {
	return smartlead.NewConnector(
		smartlead.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newMarketoConnector(
	params Parameters,
) (*marketo.Connector, error) {
	return marketo.NewConnector(
		marketo.WithWorkspace(params.Workspace),
		marketo.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newInstantlyConnector(
	params Parameters,
) (*instantly.Connector, error) {
	return instantly.NewConnector(
		instantly.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newApolloConnector(
	params Parameters,
) (*apollo.Connector, error) {
	return apollo.NewConnector(
		apollo.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newGongConnector(
	params Parameters,
) (*gong.Connector, error) {
	return gong.NewConnector(
		gong.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newAttioConnector(
	params Parameters,
) (*attio.Connector, error) {
	return attio.NewConnector(
		attio.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newPipedriveConnector(
	params Parameters,
) (*pipedrive.Connector, error) {
	return pipedrive.NewConnector(
		pipedrive.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newZohoConnector(
	params Parameters,
) (*zohocrm.Connector, error) {
	return zohocrm.NewConnector(
		zohocrm.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newCloseConnector(
	params Parameters,
) (*closecrm.Connector, error) {
	return closecrm.NewConnector(
		closecrm.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newKlaviyoConnector(
	params Parameters,
) (*klaviyo.Connector, error) {
	return klaviyo.NewConnector(
		klaviyo.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newCustomerJourneysAppConnector(
	params Parameters,
) (*customerapp.Connector, error) {
	return customerapp.NewConnector(
		customerapp.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newConstantContactConnector(
	params Parameters,
) (*constantcontact.Connector, error) {
	return constantcontact.NewConnector(
		constantcontact.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newKeapConnector(
	params Parameters,
) (*keap.Connector, error) {
	return keap.NewConnector(
		keap.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newKitConnector(
	params Parameters,
) (*kit.Connector, error) {
	return kit.NewConnector(
		kit.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newIterableConnector(
	params Parameters,
) (*iterable.Connector, error) {
	return iterable.NewConnector(
		iterable.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newAsanaConnector(
	params Parameters,
) (*asana.Connector, error) {
	return asana.NewConnector(
		asana.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newStripeConnector(
	params Parameters,
) (*stripe.Connector, error) {
	return stripe.NewConnector(
		stripe.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newZoomConnector(
	params Parameters,
) (*zoom.Connector, error) {
	return zoom.NewConnector(
		zoom.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newBrevoConnector(
	params Parameters,
) (*brevo.Connector, error) {
	return brevo.NewConnector(params)
}

func newBlueshiftConnector(
	params Parameters,
) (*blueshift.Connector, error) {
	return blueshift.NewConnector(params)
}

func newAshbyConnector(
	params Parameters,
) (*ashby.Connector, error) {
	return ashby.NewConnector(params)
}

func newGithubConnector(
	params Parameters,
) (*github.Connector, error) {
	return github.NewConnector(params)
}

func newAhaConnector(
	params Parameters,
) (*aha.Connector, error) {
	return aha.NewConnector(params)
}

func newClickUpConnector(
	params Parameters,
) (*clickup.Connector, error) {
	return clickup.NewConnector(params)
}

func newMondayConnector(
	params Parameters,
) (*monday.Connector, error) {
	return monday.NewConnector(params)
}

func newHeyReachConnector(
	params Parameters,
) (*heyreach.Connector, error) {
	return heyreach.NewConnector(params)
}

func newAWSConnector(
	params Parameters,
) (*aws.Connector, error) {
	return aws.NewConnector(params)
}

func newDriftConnector(
	params Parameters,
) (*drift.Connector, error) {
	return drift.NewConnector(params)
}

func newMixmaxConnector(
	params Parameters,
) (*mixmax.Connector, error) {
	return mixmax.NewConnector(params)
}

func newDixaConnector(
	params Parameters,
) (*dixa.Connector, error) {
	return dixa.NewConnector(params)
}

func newFrontConnector(
	params Parameters,
) (*front.Connector, error) {
	return front.NewConnector(params)
}

func newFreshdeskConnector(
	params Parameters,
) (*freshdesk.Connector, error) {
	return freshdesk.NewConnector(
		freshdesk.WithAuthenticatedClient(params.AuthenticatedClient),
		freshdesk.WithWorkspace(params.Workspace),
	)
}

func newServiceNowConnector(
	params Parameters,
) (*servicenow.Connector, error) {
	return servicenow.NewConnector(params)
}

func newChiliPiperConnector(
	params Parameters,
) (*chilipiper.Connector, error) {
	return chilipiper.NewConnector(
		chilipiper.WithAuthenticatedClient(params.AuthenticatedClient),
	)
}

func newHunterConnector(
	params Parameters,
) (*hunter.Connector, error) {
	return hunter.NewConnector(params)
}

func newPodiumConnector(
	params Parameters,
) (*podium.Connector, error) {
	return podium.NewConnector(params)
}

func newLemlistConnector(
	params Parameters,
) (*lemlist.Connector, error) {
	return lemlist.NewConnector(params)
}

func newGorgiasConnector(
	params Parameters,
) (*gorgias.Connector, error) {
	return gorgias.NewConnector(params)
}

func newZendeskChatConnector(
	params Parameters,
) (*zendeskchat.Connector, error) {
	return zendeskchat.NewConnector(params)
}

func newCapsuleConnector(
	params Parameters,
) (*capsule.Connector, error) {
	return capsule.NewConnector(params)
}

func newInstantlyAIConnector(
	params Parameters,
) (*instantlyai.Connector, error) {
	return instantlyai.NewConnector(params)
}
