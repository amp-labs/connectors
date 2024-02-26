# Contributing to Connectors

When you want to contribute to Connectors, we highly suggest you include tests.

To test connectors, you'll need your own account and instance with the provider you're testing (like Salesforce or Hubspot). You'll also need to set up credentials (using oAuth2.0) to link the connectors to the providers.

Since each provider has different credential parameters and everyone has different environments, we've made a `CredentialRegistry` and `Reader` interface to make credential loading consistent. Please use these. We currently support `JSONReader`, `EnvVarReader`, and `ValueReader`, and we're planning to add support for `yaml` files too.

For examples on how to do this, check out the tests in `test/salesforce` and `test/hubspot`.

Example:

```go
	credentialsRegistry := utils.NewCredentialsRegistry()
    readers := []utils.Reader{
		&utils.EnvReader{
			EnvName: "ACCESS_TOKEN",
			CredKey: utils.AccessToken, // or some string key to get key
		},
		&utils.EnvReader{
			EnvName: "REFRESH_TOKEN",
			CredKey: utils.RefreshToken,
		},
		&utils.JSONReader{
			FilePath: "salesforceCred1.json",
			JSONPath: "$.providerApp.clientId",
			CredKey:  utils.ClientId,
		},
		&utils.JSONReader{
			FilePath: "salesforceCred2.json",
			JSONPath: "$.providerApp.clientSecret",
			CredKey:  utils.ClientSecret,
		},
       
	}
	credentialsRegistry.AddReaders(readers...)
    credentialsRegistry.AddReader(
        &ValueReader{
            Val:     "some-salesforce-instance-subdomain",
            CredKey: "workspaceRef",
        },
    )
	salesforceWorkspace := credentialsRegistry.MustString("workspaceRef")

	cfg := utils.SalesforceOAuthConfigFromRegistry(credentialsRegistry)
	tok := utils.SalesforceOauthTokenFromRegistry(credentialsRegistry)
	ctx := context.Background()

	// Create a new Salesforce connector, with a token provider that uses the sfdx CLI to fetch an access token.
	sfc, err := connectors.Salesforce(
		salesforce.WithClient(ctx, http.DefaultClient, cfg, tok),
		salesforce.WithWorkspace(salesforceWorkspace))
	if err != nil {
		slog.Error("Error creating Salesforce connector", "error", err)

		return
	}

```