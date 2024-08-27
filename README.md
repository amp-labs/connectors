<br/>
<div align="center">
    <a href="https://www.buildwithfern.com/?utm_source=github&utm_medium=readme&utm_campaign=docs-starter-openapi&utm_content=logo">
    <img src="https://res.cloudinary.com/dycvts6vp/image/upload/v1723671980/ampersand-logo-black_fwzpfw.svg" height="50" align="center" alt="Ampersand logo" >
    </a>
<br/>
<br/><br/>


<div align="center">

[![Discord](https://img.shields.io/badge/Join%20The%20Community-black?logo=discord)](https://discord.gg/BWP4BpKHvf) [![Documentation](https://img.shields.io/badge/Read%20our%20Documentation-black?logo=book)](https://docs.withampersand.com)
</div>

</div>

<br/>


Add enterprise-grade
integrations to your SaaS this week. Bi-directional and customer-configurable. 

<div align="center">
    <img src="https://mintlify.s3-us-west-1.amazonaws.com/ampersand-24eb5c1a/images/overview.png" alt="Ampersand Overview" width="100%">
</div>


Ampersand is a declarative platform for SaaS builders who are creating product integrations. It allow you to:

* Read data from your customer’s SaaS
* Write data to your customer’s SaaS
* Coming soon: subscribe to events (creates, deletes, and field changes) in your customer’s SaaS


# Ampersand Connectors

This is a Go library that makes it easier to make API calls to SaaS products such as Salesforce and Hubspot. It handles constructing the correct API requests given desired objects and fields.

It can be either be used as a standalone library, or as a part of the [Ampersand platform](https://docs.withampersand.com/), which offers additional benefits such as:
- Handling auth flows
- Orchestration of scheduled reads, real-time writes, or bulk writes
- Handling API quotas from SaaS APIs
- A dashboard for observability and troubleshooting

## Examples

See the [examples directory](https://github.com/amp-labs/connectors/tree/main/examples) for examples of how to use the library.

| Provider      | Auth Connector                                                                                  | Deep Connector                                                                                  | AuthN                 | Notes |
|---------------|-------------------------------------------------------------------------------------------------|------------------------------------------------------------------|-----------------------|------|
| **Salesforce** | [example](https://github.com/amp-labs/connectors/tree/main/examples/auth_connectors/salesforce) | [example](https://github.com/amp-labs/connectors/tree/main/examples/deep_connectors/salesforce) | OAuth2 + Auth Code    | |
| **Adobe** | [example](https://github.com/amp-labs/connectors/tree/main/examples/auth_connectors/adobe)      | | OAuth2 + Client Creds | |
| **Anthropic** | [example](https://github.com/amp-labs/connectors/tree/main/examples/auth_connectors/antrhopic)  | | API Key               | |
| **Blueshift** | [example](https://github.com/amp-labs/connectors/tree/main/examples/auth_connectors/blueshift) | | Basic Auth            | |

## Supported connectors

Browse [the catalog](https://github.com/amp-labs/connectors/tree/main/providers) to see a list of all the connectors that Ampersand supports, and which features are supported for connector.

## How to initialize a Connector

```go
// Example for Salesforce
client, err := salesforce.NewConnector(
    salesforce.WithClient(context.Background(), http.DefaultClient, cfg, tok),
    salesforce.WithWorkspace(Workspace))
```

## Auth connectors

Auth connectors allow you to proxy through requests to a SaaS provider via Ampersand. 

#### Adding a new provider

To add a new basic connector that allows proxying through the ampersand platform, you need to add a new file to the `providers` package.

### Initialization

**Note**: If your provider requires variables to be replaced in the catalog (providers.yaml), there is a defined list of options that will replace placeholders with actual values. 
Error message will indicate what options are missing. 

For example, a provider may use `{{.workspace}}` in the base URL which needs to be replaced with an actual customer instance name. In that case, you would initialize the connector like this:

```go
conn, err := connector.NewConnector(
    providers.SomeProvider,
    connector.WithClient(context.Background(), http.DefaultClient, cfg, tok),
    connector.WithWorkspace(Workspace), // {{.workspace}}
)
```
This will **automatically** replace workspace catalog variable with an actual value that you have specified in the option.
