package main

import (
	"context"
	"log"
	"os"

	gusto "github.com/amp-labs/connectors/test/gusto"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	connector := gusto.GetConnector(ctx)

	objectNames := []string{
		"employees",
		"companies",
		"locations",
		"departments",
		"contractors",
		"payrolls",
		"pay_schedules",
		"pay_periods",
		"earning_types",
		"company_benefits",
		"jobs",
		"compensations",
		"employee_benefits",
		"garnishments",
		"home_addresses",
		"work_addresses",
		"admins",
		"contractor_payments",
		"custom_fields",
		"time_off_activities",
	}

	m, err := connector.ListObjectMetadata(ctx, objectNames)
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
