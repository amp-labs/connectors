package main

import (
	"context"
	"log"
	"os"

	acculynx "github.com/amp-labs/connectors/test/acculynx"
	"github.com/amp-labs/connectors/test/utils"
)

func main() {
	ctx := context.Background()
	connector := acculynx.GetAccuLynxConnector(ctx)

	objectNames := []string{
		"acculynx/countries",
		"acculynx/units-of-measure",
		"calendars",
		"calendars/appointments",
		"company-settings/custom-fields",
		"company-settings/job-file-settings/document-folders",
		"company-settings/job-file-settings/insurance-companies",
		"company-settings/job-file-settings/job-categories",
		"company-settings/job-file-settings/photo-video-tags",
		"company-settings/job-file-settings/trade-types",
		"company-settings/job-file-settings/work-types",
		"company-settings/job-file-settings/workflow-milestones",
		"company-settings/leads/lead-sources",
		"company-settings/location-settings/account-types",
		"contacts",
		"contacts/contact-types",
		"contacts/custom-fields",
		"contacts/email-addresses",
		"contacts/phone-numbers",
		"estimates",
		"estimates/sections",
		"jobs",
		"jobs/contacts",
		"jobs/custom-fields",
		"jobs/estimates",
		"jobs/history",
		"jobs/invoices",
		"jobs/milestone-history",
		"jobs/representatives",
		"supplements",
		"supplements/items",
		"supplements/notations",
		"users",
	}

	m, err := connector.ListObjectMetadata(ctx, objectNames)
	if err != nil {
		log.Fatal(err)
	}

	utils.DumpJSON(m, os.Stdout)
}
