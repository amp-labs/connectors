package talkdesk

import "github.com/amp-labs/connectors/internal/datautils"

var (
	filtersByUpdates  = datautils.NewSet("do-not-call-lists", "record-lists", "prompts") // nolint: gochecknoglobals
	filtersByCreation = datautils.NewSet("campaigns", "")                                // nolint: gochecknoglobals

	// "cm/core/va/cases" updates by update_at.

	// responseKeys represent the fields that stores list of records in the read response
	// these are manually retrieved.
	// example:   https://docs.talkdesk.com/reference/record-lists
	responseKeys = map[string][]string{ // nolint: gochecknoglobals
		"apps":                             {"_embedded", "apps"},
		"contacts":                         {"_embedded", "contacts"},
		"ring-groups":                      {"_embedded", "ring_groups"},
		"do-not-call-lists":                {"_embedded", "do_not_call_lists"},
		"bulk-imports/users":               {"_embedded", "imports"},
		"express/accounts":                 {"_embedded", "accounts"},
		"record-lists":                     {"_embedded", "record_lists"},
		"identity/activities":              {"_embedded"},
		"routing/attributes":               {"_embedded", "attributes"},
		"routing/attribute-categories":     {"_embedded", "attribute_categories"},
		"contacts/fetch":                   {"_embedded", "contacts"},
		"prompts":                          {"_embedded", "prompts"},
		"insights/available-queries":       {"_embedded", "available_queries"},
		"guardian/users":                   {"_embedded", "users"},
		"guardian/cases":                   {"_embedded", "cases"},
		"calls-quality":                    {"_embedded", "calls_quality"},
		"users":                            {"_embedded", "users"},
		"campaigns":                        {"_embedded", "campaigns"},
		"campaigns/scripts":                {"_embedded", "scripts"},
		"campaigns/results":                {"_embedded", "results"},
		"service-providers-resource-types": {"_embedded", "resource_types"},
		"users/resource-types":             {"_embedded", "resource_types"},
		"v2/users":                         {"_embedded", "users"},
		"account-wallet":                   {"_embedded", "wallets"},
		"cases":                            {"_embedded", "cases"},
		"cases/fields":                     {"_embedded", "fields"},
		"teams":                            {"_embedded", "teams"},
		"express/products":                 {"_embedded", "products"},
		"express/subscriptions":            {"_embedded", "subscriptions"},
		"express/invoices":                 {"_embedded", "invoices"},
		"queues":                           {"_embedded", "queues"},
		"webhooks":                         {"_embedded", "webhooks"},
		"call-recordings":                  {"_embedded", "call_recordings"},
	}
)
