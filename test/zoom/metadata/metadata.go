package main

import (
	"context"
	"fmt"
	"log"

	"github.com/amp-labs/connectors/test/zoom"
)

var objectName = "users"

func main() {
	ctx := context.Background()

	conn := zoom.GetZoomConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{objectName})

	if err != nil {
		log.Fatal(err)
	}

	// Print the results
	fmt.Println("Results: ", m.Result)

	fmt.Println("Errors: ", m.Errors)
}
