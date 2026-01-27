package main

import (
	"context"
	"log"
	"os"

	connTest "github.com/amp-labs/connectors/test/acuityscheduling"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()

	conn := connTest.GetAcuitySchedulingConnector(ctx)

	m, err := conn.ListObjectMetadata(ctx, []string{"appointments", "appointment-types", "clients"})
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
