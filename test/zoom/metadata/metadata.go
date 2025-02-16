package main

import (
	"context"
	"fmt"
	"log"

	"github.com/amp-labs/connectors/test/zoom"
)

func main() {
	ctx := context.Background()

	conn := zoom.GetZoomConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"users"})
	if err != nil {
		log.Fatal("error listing metadata for Zoom: ", err)
	}

	// Print the results
	fmt.Println("Results: ", m.Result)

	fmt.Println("Errors: ", m.Errors)
}
