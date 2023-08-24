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
  Subdomain = "<subdomain>"
  OAuthClientId = "<client id>"
  OAuthClentSecret = "<client secret>"
  OAuthAccessToken = "<access token>"
  OAuthRefreshToken = "<refresh token>"
)

func main() {
  // Set up the OAuth2 config
  cfg := &oauth2.Config{
    ClientID:     OAuthClientId,
    ClientSecret: OAuthClentSecret,
    Endpoint: oauth2.Endpoint{
      AuthURL:   fmt.Sprintf("https://%s.my.salesforce.com/services/oauth2/authorize", Subdomain),
      TokenURL:  fmt.Sprintf("https://%s.my.salesforce.com/services/oauth2/token", Subdomain),
      AuthStyle: oauth2.AuthStyleInParams,
    },
  }

  // Set up the OAuth2 token (obtained from Salesforce by authenticating)
  tok := &oauth2.Token{
    AccessToken:  OAuthAccessToken,
    RefreshToken: OAuthRefreshToken,
    TokenType:    "bearer",
    Expiry:       time.Now().Add(-1 * time.Hour), // assume it's expired already, will re-fetch.
  }

  // Create the Salesforce client
  client, err := connectors.Salesforce.New(
    salesforce.WithClient(context.Background(), http.DefaultClient, cfg, tok),
    salesforce.WithSubdomain(Subdomain))
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
