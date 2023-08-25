[![Go Report Card](https://goreportcard.com/badge/github.com/amp-labs/connectors)](https://goreportcard.com/report/github.com/amp-labs/connectors)

![Dependabot](https://img.shields.io/badge/dependabot-025E8C?style=for-the-badge&logo=dependabot&logoColor=white)

![GitHub top language](https://img.shields.io/github/languages/top/amp-labs/connectors)

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/amp-labs/connectors)

![GitHub](https://img.shields.io/github/license/amp-labs/connectors)

![GitHub last commit](https://img.shields.io/github/last-commit/amp-labs/connectors)

![GitHub issues](https://img.shields.io/github/issues/amp-labs/connectors)

![GitHub pull requests](https://img.shields.io/github/issues-pr/amp-labs/connectors)

![GitHub contributors](https://img.shields.io/github/contributors/amp-labs/connectors)

![GitHub Repo stars](https://img.shields.io/github/stars/amp-labs/connectors?style=social)

![GitHub watchers](https://img.shields.io/github/watchers/amp-labs/connectors?style=social)

![GitHub forks](https://img.shields.io/github/forks/amp-labs/connectors?style=social)

![GitHub followers](https://img.shields.io/github/followers/amp-labs?style=social)

![GitHub repo file count (file type)](https://img.shields.io/github/directory-file-count/amp-labs/connectors)

![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/amp-labs/connectors)

![GitHub repo size](https://img.shields.io/github/repo-size/amp-labs/connectors)

![GitHub commit activity](https://img.shields.io/github/commit-activity/m/amp-labs/connectors)

![GitHub language count](https://img.shields.io/github/languages/count/amp-labs/connectors)

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

// Replace with when the access token will expire,
// or leave as-is to have the token be refreshed right away.
var AccessTokenExpiry = time.Now().Add(-1 * time.Hour)

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
    Expiry:       AccessTokenExpiry,
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
