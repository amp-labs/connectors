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
		"checklists",
		"job_types",
		"price_book/material_categories",
		"price_book/materials",
		"price_book/price_forms",
		"service_zones",
	}

	m, err := connector.ListObjectMetadata(ctx, objectNames)
	if err != nil {
		utils.Fail(err.Error())
	}
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
