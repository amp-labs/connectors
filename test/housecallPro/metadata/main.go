package main

import (
	"context"
	"log"
	"os"

	housecallpro "github.com/amp-labs/connectors/test/housecallPro"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	connector := housecallpro.GetConnector(ctx)

	objectNames := []string{
		"customers",
		"employees",
		"estimates",
		"jobs",
		"job_fields/job_types",
		"leads",
		"lead_sources",
		"price_book/material_categories",
		"price_book/materials",
		"price_book/price_forms",
		"service_zones",
		"routes",
		"events",
		"tags",
		"invoices",
	}

	m, err := connector.ListObjectMetadata(ctx, objectNames)
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
