# Okta Connector

## Metadata Schema

The `metadata/schemas.json` file was generated from the [Okta OpenAPI specification](https://developer.okta.com/docs/api/).

### Schema Generation

The schema was created by:
1. Downloading the Okta Management API OpenAPI spec
2. Extracting relevant object definitions (users, groups, apps, etc.)
3. Converting to the connector's static schema format

### Supported Objects

| Object               | Resource             |
| -------------------- | -------------------- |
| Users                | users                |
| Groups               | groups               |
| Applications         | apps                 |
| System Log           | logs                 |
| Devices              | devices              |
| Identity Providers   | idps                 |
| Authorization Servers| authorizationServers |
| Trusted Origins      | trustedOrigins       |
| Network Zones        | zones                |
| Brands               | brands               |
| Domains              | domains              |
| Authenticators       | authenticators       |
| Policies             | policies             |
| Event Hooks          | eventHooks           |
| Features             | features             |

### API Reference

- [Okta API Documentation](https://developer.okta.com/docs/api/)
- [Users API](https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/)
- [Groups API](https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Group/)
- [Applications API](https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Application/)
