package main

import (
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
	salesforce := connectors.NewConnector(connectors.Salesforce, instance, token)
	data, err := salesforce.MakeGetCall(connectors.GetCallConfig{
		Endpoint: "sobjects/Account/describe",
	})
	if err != nil {
		fmt.Printf("Error making GET call: %v", err)
		return
	}
	if j, err := json.MarshalIndent(data, "", "  "); err == nil {
		fmt.Println("Successfully made GET call, response:")
		fmt.Println(string(j))
	} else {
		fmt.Printf("Error parsing data: %v", err)
	}
}
