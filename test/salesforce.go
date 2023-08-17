package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/amp-labs/connectors"
)

// To run this test, first generate a Salesforce Access token (https://ampersand.slab.com/posts/salesforce-api-guide-go1d9wnj#h0ciq-generate-an-access-token)

// Then add the token as a command line argument, e.g.
// go run test/salesforce.go '00DDp000000JQ4L!ASAAQCGoGPDpV2QkjXE.wANweSuGADZpWuh6FyY9eWUrmK6Gl4pEXG6e9qc3.KU9vqlyx_FRjlBdE6iWtbPH.yOuUbxGILpl'

// You can optionally add a second argument to specify the a Salesforce instance, or leave empty to use the Ampersand's dev instance.

// Ampersand's Salesforce dev instance
var instance = "ampersand-dev-ed.develop"

func main() {
	token := os.Args[1]
	if len(os.Args) > 2 {
		instance = os.Args[2]
	}
	salesforce := connectors.New(connectors.Salesforce, instance, func() (string, error) {
		return token, nil
	})
	res, err := salesforce.Read(context.Background(), connectors.ReadParams{
		ObjectName: "Account",
		Fields:     []string{"Id", "Name", "BillingCity", "SystemModstamp"},
	})
	if err != nil {
		panic(err)
	}

	js, _ := json.MarshalIndent(res, "", "  ")
	fmt.Println(string(js))
}
