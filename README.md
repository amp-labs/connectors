# Ampersand Connectors

This is a Go library that makes it easier to make API calls to SaaS products such as Salesforce and Hubspot. It handles constructing the correct API requests from a configuration object, and pagination logic.

Sample usage:

```go
import (
  "github.com/amp-labs/connectors"
)

func main() {
	result, err := connectors.Read(connectors.ReadConfig{
		API: connectors.Salesforce,
		ObjectName: "Contact",
		Fields: [] string { "FirstName", "LastName", "Email" },
		AccessToken: "ACCESS_TOKEN",
		WorkspaceID: "SALESFORCE_SUBDOMAIN",
	})
	if err == nil {
		fmt.Printf("Result is %v", result)
	}
  
}

```
