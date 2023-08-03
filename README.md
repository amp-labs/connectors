# Ampersand Connectors

This is a Go library that makes it easier to make API calls to SaaS products such as Salesforce and Hubspot. It handles constructing the correct API requests from a configuration object, and pagination logic.

Sample usage:

```go
import (
  "github.com/amp-labs/connectors"
)

func main() {
  salesforce := connectors.NewConnector(connectors.Salesforce, "SALESFORCE_SUBDOMAIN", "ACCESS_TOKEN")

	result, err := salesforce.Read(connectors.ReadConfig{
		ObjectName: "Contact",
		Fields: [] string { "FirstName", "LastName", "Email" },
	})
	if err == nil {
		fmt.Printf("Result is %v", result)
	}
}
```
