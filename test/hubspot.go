package main

import (
	"context"
	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/hubspot"
	"os"
)

func main() {
	os.Exit(testHubspot())
}

func testHubspot() int {

	// must initialize with connectors.Hubspot and NOT connectors.Hubspot.New
	// in order to get a hubspot.Connector back, instead of common.Connector
	c, _ := connectors.Hubspot(hubspot.WithModule(hubspot.APIModule{Label: "crm", Version: "v3"}))

	params := hubspot.SearchParams{}
	c.Search(context.Background(), params)

	return 0
}
