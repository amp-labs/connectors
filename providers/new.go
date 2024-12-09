package providers

import (
	"fmt"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/connector"
	"github.com/amp-labs/connectors/providers/apollo"
	"github.com/amp-labs/connectors/providers/atlassian"
	"github.com/amp-labs/connectors/providers/attio"
	"github.com/amp-labs/connectors/providers/closecrm"
	"github.com/amp-labs/connectors/providers/constantcontact"
	"github.com/amp-labs/connectors/providers/customerapp"
	"github.com/amp-labs/connectors/providers/docusign"
	"github.com/amp-labs/connectors/providers/dynamicscrm"
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

func NewConnector(provider Provider, parameters Parameters) (*connectors.Connector, error) {
	var connectorBuilder func(Parameters) (connectors.Connector, error)

	switch provider {
	case Hubspot:
		connectorBuilder = newHubspotConnector
	case Salesforce:
		connectorBuilder = newSalesforceConnector
	case Docusign:
		connectorBuilder = newDocusignConnector
	case Intercom:
		connectorBuilder = newIntercomConnector
	case Salesloft:
		connectorBuilder = newSalesloftConnector
	case DynamicsCRM:
		connectorBuilder = newDynamicsCRMConnector
	case ZendeskSupport:
		connectorBuilder = newZendeskSupportConnector
	case Outreach:
		connectorBuilder = newOutreachConnector
	case Atlassian:
		connectorBuilder = newAtlassianConnector
	case Pipeliner:
		connectorBuilder = newPipelinerConnector
	case Smartlead:
		connectorBuilder = newSmartleadConnector
	case Marketo:
		connectorBuilder = newMarketoConnector
	case Instantly:
		connectorBuilder = newInstantlyConnector
	case Apollo:
		connectorBuilder = newApolloConnector
	case Gong:
		connectorBuilder = newGongConnector
	case Attio:
		connectorBuilder = newAttioConnector
	case Pipedrive:
		connectorBuilder = newPipedriveConnector
	case Zoho:
		connectorBuilder = newZohoConnector
	case Close:
		connectorBuilder = newCloseConnector
	case Klaviyo:
		connectorBuilder = newKlaviyoConnector
	case CustomerJourneysApp:
		connectorBuilder = newCustomerJourneysAppConnector
	case ConstantContact:
		connectorBuilder = newConstantContactConnector
	case Keap:
		connectorBuilder = newKeapConnector
	default:
		conn, err := newGenericConnector(provider, parameters)
		if err != nil {
			return nil, fmt.Errorf("error creating generic connector: %w", err)
		}

		return &conn, nil
	}

	conn, err := connectorBuilder(parameters)
	if err != nil {
		return nil, fmt.Errorf("error creating %s connector: %w", provider, err)
	}

	return &conn, nil
}

func newGenericConnector(provider Provider, parameters Parameters) (connectors.Connector, error) {
	// Try building a generic connector
	values, err := ParseParams(parameters, ParameterAuthenticatedClient)
	if err != nil {
		return nil, fmt.Errorf("error building generic connector: %w", err)
	}

	return connector.NewConnector(
		provider,
		connector.WithAuthenticatedClient(values.AuthenticatedClient),
		connector.WithWorkspace(GetConnectorMetadata(values.Metadata, MetadataKeyWorkspace)),
	)
}

func newSalesforceConnector(parameters Parameters) (connectors.Connector, error) {
	values, err := ParseParams(parameters, ParameterAuthenticatedClient, ParameterMetadata)
	if err != nil {
		return nil, err
	}

	workspace, err := MustGetConnectorMetadata(values.Metadata, MetadataKeyWorkspace)
	if err != nil {
		return nil, err
	}

	return salesforce.NewConnector(
		salesforce.WithAuthenticatedClient(values.AuthenticatedClient),
		salesforce.WithWorkspace(workspace))
}

func newHubspotConnector(parameters Parameters) (connectors.Connector, error) {
	values, err := ParseParams(parameters, ParameterAuthenticatedClient)
	if err != nil {
		return nil, err
	}

	if values.Module == "" {
		values.Module = hubspot.ModuleCRM
	}

	return hubspot.NewConnector(
		hubspot.WithAuthenticatedClient(values.AuthenticatedClient),
		hubspot.WithModule(values.Module))
}

func newDocusignConnector(parameters Parameters) (connectors.Connector, error) {
	values, err := ParseParams(parameters, ParameterAuthenticatedClient, ParameterMetadata)
	if err != nil {
		return nil, err
	}

	server, err := MustGetConnectorMetadata(values.Metadata, MetadataKeyServer)
	if err != nil {
		return nil, err
	}

	return docusign.NewConnector(
		docusign.WithAuthenticatedClient(values.AuthenticatedClient),

		// TODO: Add a WithServer option to the Docusign connector, or let the connector validate the server
		// and pass in the metadata directly
		docusign.WithMetadata(map[string]string{
			string(MetadataKeyServer): server,
		}),
	)
}

func newIntercomConnector(parameters Parameters) (connectors.Connector, error) {
	values, err := ParseParams(parameters, ParameterAuthenticatedClient)
	if err != nil {
		return nil, err
	}

	return intercom.NewConnector(
		intercom.WithAuthenticatedClient(values.AuthenticatedClient),
	)
}

func newSalesloftConnector(parameters Parameters) (connectors.Connector, error) {
	values, err := ParseParams(parameters, ParameterAuthenticatedClient)
	if err != nil {
		return nil, err
	}

	return salesloft.NewConnector(
		salesloft.WithAuthenticatedClient(values.AuthenticatedClient),
	)
}

func newDynamicsCRMConnector(parameters Parameters) (connectors.Connector, error) {
	values, err := ParseParams(parameters, ParameterAuthenticatedClient)
	if err != nil {
		return nil, err
	}

	workspace, err := MustGetConnectorMetadata(values.Metadata, MetadataKeyWorkspace)
	if err != nil {
		return nil, err
	}

	return dynamicscrm.NewConnector(
		dynamicscrm.WithWorkspace(workspace),
		dynamicscrm.WithAuthenticatedClient(values.AuthenticatedClient),
	)
}

func newOutreachConnector(parameters Parameters) (connectors.Connector, error) {
	values, err := ParseParams(parameters, ParameterAuthenticatedClient)
	if err != nil {
		return nil, err
	}

	return outreach.NewConnector(
		outreach.WithAuthenticatedClient(values.AuthenticatedClient),
	)
}

func newZendeskSupportConnector(parameters Parameters) (connectors.Connector, error) {
	values, err := ParseParams(parameters, ParameterAuthenticatedClient)
	if err != nil {
		return nil, err
	}

	workspace, err := MustGetConnectorMetadata(values.Metadata, MetadataKeyWorkspace)
	if err != nil {
		return nil, err
	}

	return zendesksupport.NewConnector(
		zendesksupport.WithWorkspace(workspace),
		zendesksupport.WithAuthenticatedClient(values.AuthenticatedClient),
	)
}

func newAtlassianConnector(parameters Parameters) (connectors.Connector, error) {
	values, err := ParseParams(parameters, ParameterAuthenticatedClient, ParameterMetadata)
	if err != nil {
		return nil, err
	}

	workspace, err := MustGetConnectorMetadata(values.Metadata, MetadataKeyWorkspace)
	if err != nil {
		return nil, err
	}

	cloudId, err := MustGetConnectorMetadata(values.Metadata, MetadataKeyCloudId)
	if err != nil {
		return nil, err
	}

	opts := []atlassian.Option{
		atlassian.WithAuthenticatedClient(values.AuthenticatedClient),
		atlassian.WithModule(values.Module),
		atlassian.WithWorkspace(workspace),
	}

	if values.Module == atlassian.ModuleJira {
		// TODO: Add a WithCloudId option to the Atlassian connector, or let the connector validate the cloud ID
		// and pass in the metadata directly
		opts = append(opts, atlassian.WithMetadata(map[string]string{
			string(MetadataKeyCloudId): cloudId,
		}))
	}

	return atlassian.NewConnector(opts...)
}

func newPipelinerConnector(parameters Parameters) (connectors.Connector, error) {
	values, err := ParseParams(parameters, ParameterAuthenticatedClient)
	if err != nil {
		return nil, err
	}

	workspace, err := MustGetConnectorMetadata(values.Metadata, MetadataKeyWorkspace)
	if err != nil {
		return nil, err
	}

	return pipeliner.NewConnector(
		pipeliner.WithWorkspace(workspace),
		pipeliner.WithAuthenticatedClient(values.AuthenticatedClient),
	)
}

func newSmartleadConnector(parameters Parameters) (connectors.Connector, error) {
	values, err := ParseParams(parameters, ParameterAuthenticatedClient)
	if err != nil {
		return nil, err
	}

	return smartlead.NewConnector(
		smartlead.WithAuthenticatedClient(values.AuthenticatedClient),
	)
}

func newMarketoConnector(parameters Parameters) (connectors.Connector, error) {
	values, err := ParseParams(parameters, ParameterAuthenticatedClient)
	if err != nil {
		return nil, err
	}

	workspace, err := MustGetConnectorMetadata(values.Metadata, MetadataKeyWorkspace)
	if err != nil {
		return nil, err
	}

	if values.Module == "" {
		values.Module = marketo.ModuleEmpty
	}

	return marketo.NewConnector(
		marketo.WithWorkspace(workspace),
		marketo.WithAuthenticatedClient(values.AuthenticatedClient),
		marketo.WithModule(values.Module),
	)
}

func newInstantlyConnector(parameters Parameters) (connectors.Connector, error) {
	values, err := ParseParams(parameters, ParameterAuthenticatedClient)
	if err != nil {
		return nil, err
	}

	return instantly.NewConnector(
		instantly.WithAuthenticatedClient(values.AuthenticatedClient),
	)
}

func newApolloConnector(parameters Parameters) (connectors.Connector, error) {
	values, err := ParseParams(parameters, ParameterAuthenticatedClient)
	if err != nil {
		return nil, err
	}

	return apollo.NewConnector(
		apollo.WithAuthenticatedClient(values.AuthenticatedClient),
	)
}

func newGongConnector(parameters Parameters) (connectors.Connector, error) {
	values, err := ParseParams(parameters, ParameterAuthenticatedClient)
	if err != nil {
		return nil, err
	}

	return gong.NewConnector(
		gong.WithAuthenticatedClient(values.AuthenticatedClient),
	)
}

func newAttioConnector(parameters Parameters) (connectors.Connector, error) {
	values, err := ParseParams(parameters, ParameterAuthenticatedClient)
	if err != nil {
		return nil, err
	}

	return attio.NewConnector(
		attio.WithAuthenticatedClient(values.AuthenticatedClient),
	)
}

func newPipedriveConnector(parameters Parameters) (connectors.Connector, error) {
	values, err := ParseParams(parameters, ParameterAuthenticatedClient)
	if err != nil {
		return nil, err
	}

	return pipedrive.NewConnector(
		pipedrive.WithAuthenticatedClient(values.AuthenticatedClient),
	)
}

func newZohoConnector(parameters Parameters) (connectors.Connector, error) {
	values, err := ParseParams(parameters, ParameterAuthenticatedClient)
	if err != nil {
		return nil, err
	}

	return zohocrm.NewConnector(
		zohocrm.WithAuthenticatedClient(values.AuthenticatedClient),
	)
}

func newCloseConnector(parameters Parameters) (connectors.Connector, error) {
	values, err := ParseParams(parameters, ParameterAuthenticatedClient)
	if err != nil {
		return nil, err
	}

	return closecrm.NewConnector(
		closecrm.WithAuthenticatedClient(values.AuthenticatedClient),
	)
}

func newKlaviyoConnector(parameters Parameters) (connectors.Connector, error) {
	values, err := ParseParams(parameters, ParameterAuthenticatedClient)
	if err != nil {
		return nil, err
	}

	return klaviyo.NewConnector(
		klaviyo.WithAuthenticatedClient(values.AuthenticatedClient),
	)
}

func newCustomerJourneysAppConnector(parameters Parameters) (connectors.Connector, error) {
	values, err := ParseParams(parameters, ParameterAuthenticatedClient)
	if err != nil {
		return nil, err
	}

	return customerapp.NewConnector(
		customerapp.WithAuthenticatedClient(values.AuthenticatedClient),
	)
}

func newConstantContactConnector(parameters Parameters) (connectors.Connector, error) {
	values, err := ParseParams(parameters, ParameterAuthenticatedClient)
	if err != nil {
		return nil, err
	}

	return constantcontact.NewConnector(
		constantcontact.WithAuthenticatedClient(values.AuthenticatedClient),
	)
}

func newKeapConnector(parameters Parameters) (connectors.Connector, error) {
	values, err := ParseParams(parameters, ParameterAuthenticatedClient)
	if err != nil {
		return nil, err
	}

	return keap.NewConnector(
		keap.WithAuthenticatedClient(values.AuthenticatedClient),
	)
}
