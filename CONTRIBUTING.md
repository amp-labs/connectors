# Contributing to Connectors

When you want to contribute to Connectors, we highly suggest you include tests.

To test connectors, you'll need your own account and instance with the provider you're testing (like Salesforce or Hubspot). You'll also need to set up credentials (using oAuth2.0) to link the connectors to the providers.

Since each provider has different credential parameters and everyone has different environments, we've made a `CredentialRegistry` and `Reader` interface to make credential loading consistent. Please use these. We currently support `JSONReader`, `EnvVarReader`, and `ValueReader`.

For examples on how to do this, check out the tests in `test/salesforce` and `test/hubspot`.

Example:

`salesforceCred1.json`
```json
{
    "id": "someID",
    "projectId": "someProjectId",
    "providerApp": {
        "id": "someId",
        "externalRef": "externalRef:p0:e0:pr0",
        "provider": "salesforce",
        "clientId": "YOUR_CLIENT_ID1",
        "clientSecret": "YOUR_CLIENT_SECRET1",
        "scopes": null,
        "projectId": "someProjectId"
    },
}
```

`salesforceCred2.json`
```json
{
    "id": "someId2",
    "providerApp": {
        "id": "someId2",
        "externalRef": "externalRef:p0:e0:pr0",
        "provider": "salesforce",
        "clientId": "YOUR_CLIENT_ID2",
        "clientSecret": "YOUR_CLIENT_SECRET2",
        "scopes": null,
        "projectId": "someProjectId2"
    },
}
```

```go
	credentialsRegistry := utils.NewCredentialsRegistry()
    readers := []utils.Reader{
		&utils.EnvReader{
			EnvName: "ACCESS_TOKEN",  // This reads an environment variable named ACCESS_TOKEN
			CredKey: utils.AccessToken, // or some string key to get key
		},
		&utils.EnvReader{
			EnvName: "REFRESH_TOKEN", // This reads an environment variable named REFRESH_TOKEN
			CredKey: utils.RefreshToken,
		},
		&utils.JSONReader{
			FilePath: "salesforceCred1.json", // This reads from a file
			JSONPath: "$.providerApp.clientId",
			CredKey:  utils.ClientId,
		},
		&utils.JSONReader{
			FilePath: "salesforceCred2.json", // This reads from a file
			JSONPath: "$.providerApp.clientSecret",
			CredKey:  utils.ClientSecret,
		},
       
	}
	credentialsRegistry.AddReaders(readers...)
    credentialsRegistry.AddReader(
        &ValueReader{
            Val:     "some-salesforce-instance-subdomain", // Hard coded value
            CredKey: "workspaceRef",
        },
    )
	salesforceWorkspace := credentialsRegistry.MustString("workspaceRef")

	cfg := utils.SalesforceOAuthConfigFromRegistry(credentialsRegistry)
	tok := utils.SalesforceOauthTokenFromRegistry(credentialsRegistry)
	ctx := context.Background()

	sfc, err := connectors.Salesforce(
		salesforce.WithClient(ctx, http.DefaultClient, cfg, tok),
		salesforce.WithWorkspace(salesforceWorkspace))
	if err != nil {
		slog.Error("Error creating Salesforce connector", "error", err)

		return
	}

```