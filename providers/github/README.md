
# GitHub Connector

## Authentication Methods
GitHub provides multiple authentication methods:
- Personal Access Token (PAT)
- OAuth2 (used by this connector)
- GitHub App Authentication(JSON Web Tokens)
- Fine-grained Personal Access Token (Beta)


### Authentication Edge Cases
Some endpoints only support specific authentication methods. For example:
- Notification endpoints require Fine-grained access tokens.
- Some endpoints don't support OAuth authentication at all
- Enterprise-specific endpoints might require different authentication methods

### Since Parameter Variations
The `since` parameter behavior varies across endpoints:

1. Timestamp-based (most common):
   - Format: ISO 8601 (`YYYY-MM-DDTHH:MM:SSZ`)
   - Example: `?since=2024-03-15T14:30:45Z`

2. ID-based (numeric):
   - Some endpoints use `since` as an ID filter instead of a timestamp
   - Format: Integer
   - Example: `?since=12345`
   - Used in endpoints like:
     - `/users` (filters by user ID)
