<br/>
<div align="center">
    <a href="https://www.withampersand.com/?utm_source=github&utm_medium=readme&utm_campaign=connectors-repo&utm_content=logo">
    <img src="https://res.cloudinary.com/dycvts6vp/image/upload/v1723671980/ampersand-logo-black.svg" height="30" align="center" alt="Ampersand logo" >
    </a>
<br/>
<br/>

<div align="center">

[![Star us on GitHub](https://img.shields.io/github/stars/amp-labs/connectors?color=FFD700&label=Stars&logo=Github)](https://github.com/amp-labs/connectors) [![Discord](https://img.shields.io/badge/Join%20The%20Community-black?logo=discord)](https://discord.gg/BWP4BpKHvf) [![Documentation](https://img.shields.io/badge/Read%20our%20Documentation-black?logo=book)](https://docs.withampersand.com) ![PRs welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg) <img src="https://img.shields.io/static/v1?label=license&message=MIT&color=white" alt="License">
</div>

</div>

# Overview

[Ampersand](https://withampersand.com?trk=readme-github) is a declarative platform for SaaS builders who are creating product integrations. It allows you to:

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

The key components of the Ampersand platform include:

- Manifest file (`amp.yaml`): Define all your integrations, the APIs to connect to, the objects and fields for reading or writing, and the configuration options you want to expose to your customers.

- Ampersand server: a managed service that keeps track of each of your customer’s configurations, and makes the appropriate API calls to your customer’s SaaS, while optimizing for cost, handling retries and error message parsing.

- Embeddable UI components: open-source React components that you can embed to allow your end users to customize and manage their integrations. See [the repo](https://github.com/amp-labs/react) for more info.

- Dashboard: Provides deep observability into customer integrations, allowing you to monitor & troubleshoot with detailed logs.

Add enterprise-grade integrations to your SaaS this week. **[Get started for free](https://dashboard.withampersand.com/sign-up?trk=readme-github)**.

<div align="center">
    <img src="https://res.cloudinary.com/dycvts6vp/image/upload/v1724756323/media/hqukkkmpk96zavslpmw5.png" alt="Ampersand Overview" width="80%">
</div>

# Using connectors

## Supported connectors

Browse [the providers directory](https://github.com/amp-labs/connectors/tree/main/providers) to see a list of all the connectors that Ampersand supports, and which features are supported for each connector.

## Examples

Visit the [Ampersand docs](https://docs.withampersand.com?trk=readme-github) to learn about how to use connectors as a part of the Ampersand platform. 

See the [examples directory](https://github.com/amp-labs/connectors/tree/main/examples) for examples of how to use connectors as a standalone library.

| Provider      | Auth Connector  | Deep Connector | Authorization Method |
|---------------|-----------------|----------------|----------------------|
| **Salesforce** | [example](https://github.com/amp-labs/connectors/tree/main/examples/auth_connectors/salesforce) | [example](https://github.com/amp-labs/connectors/tree/main/examples/deep_connectors/salesforce) | OAuth2, Authorization Code    |
| **Adobe** | [example](https://github.com/amp-labs/connectors/tree/main/examples/auth_connectors/adobe)      | | OAuth2, Client Credentials | 
| **Anthropic** | [example](https://github.com/amp-labs/connectors/tree/main/examples/auth_connectors/antrhopic)  | | API Key               |
| **Blueshift** | [example](https://github.com/amp-labs/connectors/tree/main/examples/auth_connectors/blueshift) | | Basic Auth            |

# Contributors

Thankful to the OSS community for making Ampersand better every day.

<a href="https://github.com/amp-labs/connectors/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=amp-labs/connectors" />
</a>
