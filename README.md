# Ampersand Connectors

This is a Go library that makes it easier to make API calls to SaaS products such as Salesforce and Hubspot. It handles constructing the correct API requests from a configuration object, and pagination logic.

Sample usage:

```go
import (
  "context"
  "fmt"
  "net/http"
  "time"

  "github.com/amp-labs/connectors"
  "github.com/amp-labs/connectors/salesforce"
  "golang.org/x/oauth2"
)

const (
  // Replace these with your own values.
  Workspace = "<workspace>"
  OAuthClientId = "<client id>"
  OAuthClentSecret = "<client secret>"
  OAuthAccessToken = "<access token>"
  OAuthRefreshToken = "<refresh token>"

)

// Replace with when the access token will expire,
// or leave as-is to have the token be refreshed right away.
var AccessTokenExpiry = time.Now().Add(-1 * time.Hour)

func main() {
  // Set up the OAuth2 config
  cfg := &oauth2.Config{
    ClientID:     OAuthClientId,
    ClientSecret: OAuthClentSecret,
    Endpoint: oauth2.Endpoint{
      AuthURL:   fmt.Sprintf("https://%s.my.salesforce.com/services/oauth2/authorize", Workspace),
      TokenURL:  fmt.Sprintf("https://%s.my.salesforce.com/services/oauth2/token", Workspace),
      AuthStyle: oauth2.AuthStyleInParams,
    },
  }

  // Set up the OAuth2 token (obtained from Salesforce by authenticating)
  tok := &oauth2.Token{
    AccessToken:  OAuthAccessToken,
    RefreshToken: OAuthRefreshToken,
    TokenType:    "bearer",
    Expiry:       AccessTokenExpiry,
  }

  // Create the Salesforce client
  client, err := connectors.Salesforce(
    salesforce.WithClient(context.Background(), http.DefaultClient, cfg, tok),
    salesforce.WithWorkspace(Workspace))
  if err != nil {
    panic(err)
  }

  // Make a request to Salesforce
  result, err := client.Read(context.Background(), connectors.ReadConfig{
    ObjectName: "Contact",
    Fields: []string{"FirstName", "LastName", "Email"},
  })
  if err == nil {
    fmt.Printf("Result is %v", result)
  }
}
```

## Provider catalog

The provider catalog is a YAML file that contains information about all SaaS providers that Ampersand supports. It contains information such as the base URL, and the operations that the provider supports. This is currently stored in the `providers/providers.yaml` file.

## Ways to initialize a Connector

There are 3 ways to initialize a Connector:

1. Initializing a provider-specific Connector (returns a concrete type). This method of initialization will allow you to use methods that only exist for that provider.

```go
client, err := connectors.Salesforce(
    salesforce.WithClient(context.Background(), http.DefaultClient, cfg, tok),
    salesforce.WithWorkspace(Workspace))
```

2. Initializing a generic Connector (returns an interface). This method of initialization will only allow you to use methods that are common to all providers. This is helpful if you would like your code to be provider-agnostic.

```go
client, err := connectors.Salesforce.New(
    salesforce.WithClient(context.Background(), http.DefaultClient, cfg, tok),
    salesforce.WithWorkspace(Workspace))
```

3. With string parameter for API name (this is useful if you are parsing the API name from a config file, but should be avoided otherwise because it is not typesafe). This returns a generic Connector (returns an interface).

```go
client, err := connectors.New("salesforce", map[string]any{"workspace": "salesforce-instance-name"})
```

## Basic connectors

Basic connectors allow you to proxy through requests to a SaaS provider via Ampersand. 

#### Adding a new provider

To add a new basic connector that allows proxying through the ampersand platform, you need to add a new entry to the `providers.yaml` file. The entry should have some required fields which are tagged with `validate: required` in the `providers/types.go` file.


### Initialization

**Note**: If your provider requires variables to be replaced in the catalog (providers.yaml), use the `WithCatalogSubstitutions` option to replace placeholders with actual values. 
If you don't, the provider info will be incomplete and the connector will not work. 

For example, a provider may use `{{workspace}}` in an option (maybe in the base URL) which needs to be replaced with an actual customer instance name. In that case, you would initialize the connector like this:

```go
conn, err := basic.NewConnector(
    providers.SomeProvider,
    basic.WithClient(context.Background(), http.DefaultClient, cfg, tok),
    
    // WithCatalogSubstitutions allows you to replace placeholders in the catalog (providers.yaml) with actual values.
    basic.WithCatalogSubstitutions(map[string]string{
		"workspace": "customer-instance-ref"
	}),
)
```
