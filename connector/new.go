// nolint
package connector

import (
	"fmt"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/apollo"
	"github.com/amp-labs/connectors/providers/atlassian"
	"github.com/amp-labs/connectors/providers/attio"
	"github.com/amp-labs/connectors/providers/closecrm"
	"github.com/amp-labs/connectors/providers/constantcontact"
	"github.com/amp-labs/connectors/providers/customerapp"
	"github.com/amp-labs/connectors/providers/docusign"
	"github.com/amp-labs/connectors/providers/dynamicscrm"
	"github.com/amp-labs/connectors/providers/generic"
	"github.com/amp-labs/connectors/providers/gong"
	"github.com/amp-labs/connectors/providers/hubspot"
	"github.com/amp-labs/connectors/providers/instantly"
	"github.com/amp-labs/connectors/providers/intercom"
	"github.com/amp-labs/connectors/providers/keap"
	"github.com/amp-labs/connectors/providers/klaviyo"
	"github.com/amp-labs/connectors/providers/marketo"
	"github.com/amp-labs/connectors/providers/outreach"
	"github.com/amp-labs/connectors/providers/pipedrive"
	"github.com/amp-labs/connectors/providers/pipeliner"
	"github.com/amp-labs/connectors/providers/salesforce"
	"github.com/amp-labs/connectors/providers/salesloft"
	"github.com/amp-labs/connectors/providers/smartlead"
	"github.com/amp-labs/connectors/providers/zendesksupport"
	"github.com/amp-labs/connectors/providers/zohocrm"
)

func NewConnector(provider providers.Provider, parameters Parameters) (connectors.Connector, error) {
	var (
		conn connectors.Connector
		err  error
	)

	switch provider {
	case providers.Hubspot:
		conn, err = buildConnector(newHubspotConnector, parameters, mustAuthenticatedClient)
	case providers.Salesforce:
		conn, err = buildConnector(newSalesforceConnector, parameters, mustAuthenticatedClient, mustWorkspace)
	case providers.Docusign:
		conn, err = buildConnector(newDocusignConnector, parameters, mustAuthenticatedClient, mustMetadata)
	case providers.Intercom:
		conn, err = buildConnector(newIntercomConnector, parameters, mustAuthenticatedClient)
	case providers.Salesloft:
		conn, err = buildConnector(newSalesloftConnector, parameters, mustAuthenticatedClient)
	case providers.DynamicsCRM:
		conn, err = buildConnector(newDynamicsCRMConnector, parameters, mustAuthenticatedClient, mustWorkspace)
	case providers.ZendeskSupport:
		conn, err = buildConnector(newZendeskSupportConnector, parameters, mustAuthenticatedClient, mustWorkspace)
	case providers.Outreach:
		conn, err = buildConnector(newOutreachConnector, parameters, mustAuthenticatedClient)
	case providers.Atlassian:
		conn, err = buildConnector(newAtlassianConnector, parameters, mustAuthenticatedClient, mustWorkspace)
	case providers.Pipeliner:
		conn, err = buildConnector(newPipelinerConnector, parameters, mustAuthenticatedClient, mustWorkspace)
	case providers.Smartlead:
		conn, err = buildConnector(newSmartleadConnector, parameters, mustAuthenticatedClient)
	case providers.Marketo:
		conn, err = buildConnector(newMarketoConnector, parameters, mustAuthenticatedClient, mustWorkspace)
	case providers.Instantly:
		conn, err = buildConnector(newInstantlyConnector, parameters, mustAuthenticatedClient)
	case providers.Apollo:
		conn, err = buildConnector(newApolloConnector, parameters, mustAuthenticatedClient)
	case providers.Gong:
		conn, err = buildConnector(newGongConnector, parameters, mustAuthenticatedClient)
	case providers.Attio:
		conn, err = buildConnector(newAttioConnector, parameters, mustAuthenticatedClient)
	case providers.Pipedrive:
		conn, err = buildConnector(newPipedriveConnector, parameters, mustAuthenticatedClient)
	case providers.Zoho:
		conn, err = buildConnector(newZohoConnector, parameters, mustAuthenticatedClient)
	case providers.Close:
		conn, err = buildConnector(newCloseConnector, parameters, mustAuthenticatedClient)
	case providers.Klaviyo:
		conn, err = buildConnector(newKlaviyoConnector, parameters, mustAuthenticatedClient)
	case providers.CustomerJourneysApp:
		conn, err = buildConnector(newCustomerJourneysAppConnector, parameters, mustAuthenticatedClient)
	case providers.ConstantContact:
		conn, err = buildConnector(newConstantContactConnector, parameters, mustAuthenticatedClient)
	case providers.Keap:
		conn, err = buildConnector(newKeapConnector, parameters, mustAuthenticatedClient)
	default:
		conn, err = newGenericConnector(provider, parameters, mustAuthenticatedClient)
	}

	if err != nil {
		return nil, fmt.Errorf("error creating %s connector: %w", provider, err)
	}

	return conn, nil
}

func buildConnector(
	builder func(Parameters) (connectors.Connector, error),
	parameters Parameters,
	opts ...Option,
) (connectors.Connector, error) {
	for _, opt := range opts {
		opt(&parameters)
	}

	if parameters.validity.invalid {
		return nil, parameters.validity.error
	}

	return builder(parameters)
}

func newGenericConnector(provider providers.Provider, parameters Parameters, opts ...Option) (connectors.Connector, error) {
	for _, opt := range opts {
		opt(&parameters)
	}

	if parameters.validity.invalid {
		return nil, parameters.validity.error
	}

	return generic.NewConnector(
		provider,
		generic.WithAuthenticatedClient(parameters.AuthenticatedClient),
		generic.WithWorkspace(parameters.Workspace),
	)
}

func newSalesforceConnector(parameters Parameters) (connectors.Connector, error) {
	return salesforce.NewConnector(
		salesforce.WithAuthenticatedClient(parameters.AuthenticatedClient),
		salesforce.WithWorkspace(parameters.Workspace))
}

func newHubspotConnector(parameters Parameters) (connectors.Connector, error) {
	if parameters.Module == "" {
		parameters.Module = hubspot.ModuleCRM
	}

	return hubspot.NewConnector(
		hubspot.WithAuthenticatedClient(parameters.AuthenticatedClient),
		hubspot.WithModule(parameters.Module))
}

func newDocusignConnector(parameters Parameters) (connectors.Connector, error) {
	// Server needs to be available in the metadata
	if _, err := MustGetConnectorMetadata(parameters.Metadata, MetadataKeyServer); err != nil {
		return nil, err
	}

	return docusign.NewConnector(
		docusign.WithAuthenticatedClient(parameters.AuthenticatedClient),
		docusign.WithMetadata(parameters.Metadata),
	)
}

func newIntercomConnector(parameters Parameters) (connectors.Connector, error) {
	return intercom.NewConnector(
		intercom.WithAuthenticatedClient(parameters.AuthenticatedClient),
	)
}

func newSalesloftConnector(parameters Parameters) (connectors.Connector, error) {
	return salesloft.NewConnector(
		salesloft.WithAuthenticatedClient(parameters.AuthenticatedClient),
	)
}

func newDynamicsCRMConnector(parameters Parameters) (connectors.Connector, error) {
	return dynamicscrm.NewConnector(
		dynamicscrm.WithWorkspace(parameters.Workspace),
		dynamicscrm.WithAuthenticatedClient(parameters.AuthenticatedClient),
	)
}

func newOutreachConnector(parameters Parameters) (connectors.Connector, error) {
	return outreach.NewConnector(
		outreach.WithAuthenticatedClient(parameters.AuthenticatedClient),
	)
}

func newZendeskSupportConnector(parameters Parameters) (connectors.Connector, error) {
	return zendesksupport.NewConnector(
		zendesksupport.WithWorkspace(parameters.Workspace),
		zendesksupport.WithAuthenticatedClient(parameters.AuthenticatedClient),
	)
}

func newAtlassianConnector(parameters Parameters) (connectors.Connector, error) {
	atlassianOpts := []atlassian.Option{
		atlassian.WithAuthenticatedClient(parameters.AuthenticatedClient),
		atlassian.WithWorkspace(parameters.Workspace),
	}

	if parameters.Module == "" {
		parameters.Module = atlassian.ModuleJira
	}

	if parameters.Module == atlassian.ModuleJira {
		// TODO: Validate the cloudId inside the connector
		atlassianOpts = append(atlassianOpts, atlassian.WithMetadata(parameters.Metadata))
	}

	atlassianOpts = append(atlassianOpts, atlassian.WithModule(parameters.Module))

	return atlassian.NewConnector(atlassianOpts...)
}

func newPipelinerConnector(parameters Parameters) (connectors.Connector, error) {
	return pipeliner.NewConnector(
		pipeliner.WithWorkspace(parameters.Workspace),
		pipeliner.WithAuthenticatedClient(parameters.AuthenticatedClient),
	)
}

func newSmartleadConnector(parameters Parameters) (connectors.Connector, error) {
	return smartlead.NewConnector(
		smartlead.WithAuthenticatedClient(parameters.AuthenticatedClient),
	)
}

func newMarketoConnector(parameters Parameters) (connectors.Connector, error) {
	if parameters.Module == "" {
		parameters.Module = marketo.ModuleLeads
	}

	return marketo.NewConnector(
		marketo.WithWorkspace(parameters.Workspace),
		marketo.WithAuthenticatedClient(parameters.AuthenticatedClient),
		marketo.WithModule(parameters.Module),
	)
}

func newInstantlyConnector(parameters Parameters) (connectors.Connector, error) {
	return instantly.NewConnector(
		instantly.WithAuthenticatedClient(parameters.AuthenticatedClient),
	)
}

func newApolloConnector(parameters Parameters) (connectors.Connector, error) {
	return apollo.NewConnector(
		apollo.WithAuthenticatedClient(parameters.AuthenticatedClient),
	)
}

func newGongConnector(parameters Parameters) (connectors.Connector, error) {
	return gong.NewConnector(
		gong.WithAuthenticatedClient(parameters.AuthenticatedClient),
	)
}

func newAttioConnector(parameters Parameters) (connectors.Connector, error) {
	return attio.NewConnector(
		attio.WithAuthenticatedClient(parameters.AuthenticatedClient),
	)
}

func newPipedriveConnector(parameters Parameters) (connectors.Connector, error) {
	return pipedrive.NewConnector(
		pipedrive.WithAuthenticatedClient(parameters.AuthenticatedClient),
	)
}

func newZohoConnector(parameters Parameters) (connectors.Connector, error) {
	return zohocrm.NewConnector(
		zohocrm.WithAuthenticatedClient(parameters.AuthenticatedClient),
	)
}

func newCloseConnector(parameters Parameters) (connectors.Connector, error) {
	return closecrm.NewConnector(
		closecrm.WithAuthenticatedClient(parameters.AuthenticatedClient),
	)
}

func newKlaviyoConnector(parameters Parameters) (connectors.Connector, error) {
	return klaviyo.NewConnector(
		klaviyo.WithAuthenticatedClient(parameters.AuthenticatedClient),
	)
}

func newCustomerJourneysAppConnector(parameters Parameters) (connectors.Connector, error) {
	return customerapp.NewConnector(
		customerapp.WithAuthenticatedClient(parameters.AuthenticatedClient),
	)
}

func newConstantContactConnector(parameters Parameters) (connectors.Connector, error) {
	return constantcontact.NewConnector(
		constantcontact.WithAuthenticatedClient(parameters.AuthenticatedClient),
	)
}

func newKeapConnector(parameters Parameters) (connectors.Connector, error) {
	return keap.NewConnector(
		keap.WithAuthenticatedClient(parameters.AuthenticatedClient),
	)
}
