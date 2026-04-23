package main

import (
	"context"
	"log"
	"os"

	"github.com/amp-labs/connectors/test/odoo"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	connector := odoo.GetConnector(ctx)

	objectNames := []string{
		"res.partner",
		"crm.lead",
		"uom.uom",
		"res.currency",
		"res.company",
		"resource.calendar",
	}

	m, err := connector.ListObjectMetadata(ctx, objectNames)
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
