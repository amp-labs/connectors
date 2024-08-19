package main

import (
	"context"
	"fmt"
	"log"

	"github.com/amp-labs/connectors/test/marketo"
)

func main() {
	ctx := context.Background()

	conn := marketo.GetMarketoConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"channels", "emailTemplates"})
	if err != nil {
		log.Fatal(err)
	}

	// Print the results
	fmt.Println("Results: ", m.Result)
	fmt.Println("Errors: ", m.Errors)

}
