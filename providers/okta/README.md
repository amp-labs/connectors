# Okta Connector

## Metadata Schema

The `metadata/schemas.json` file is generated from the Okta Management API OpenAPI specification.

### OpenAPI Spec

The OpenAPI specification file is located at `providers/okta/metadata/openapi/management-minimal.yaml`.
It was downloaded from the [okta/okta-management-openapi-spec](https://github.com/okta/okta-management-openapi-spec) repository
(version `2025.01.1`, `management-minimal.yaml` variant).

To update to a newer version, download the latest `management-minimal.yaml` from
`https://github.com/okta/okta-management-openapi-spec/tree/master/dist/<VERSION>/management-minimal.yaml`
and replace the existing file.

### Schema Generation

To regenerate the schemas, run the following command from the root of the repository:

```bash
go run ./scripts/openapi/okta/metadata
```

Check `providers/okta/metadata/schemas.json` for any side effects.

Note: The `devices` and `policies` objects are added manually in the generation script
because the OpenAPI spec uses `allOf`/`$ref` patterns that the tooling cannot auto-extract.

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
- [Okta Management OpenAPI Spec](https://github.com/okta/okta-management-openapi-spec)
- [Users API](https://developer.okta.com/docs/api/openapi/okta-management/management/tag/User/)
- [Groups API](https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Group/)
- [Applications API](https://developer.okta.com/docs/api/openapi/okta-management/management/tag/Application/)
