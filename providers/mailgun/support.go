package mailgun

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/providers/mailgun/metadata"
)

// readPagination describes how an object paginates on read.
type readPagination string

const (
	// Response has paging.next or paging.Next as a full URL or relative path.
	readPaginationPagingNext readPagination = "paging_next"

	// Response has total_count; request uses limit + skip.
	readPaginationTotalCountSkip readPagination = "total_count_skip"

	// Response has total; request uses limit + skip.
	readPaginationTotalSkip readPagination = "total_skip"

	// Request accepts limit only; no next page is resolved.
	readPaginationLimitOnly readPagination = "limit_only"

	// Single-page response; no pagination query params.
	readPaginationNone readPagination = "none"
)

// objectReadPagination maps each readable object to its pagination strategy.
//
// paging_next:
//   - templates, ip_warmups, forwards, alerts/slack/channels, lists/pages, dkim/keys
//   - dynamic_pools/domains, dynamic_pools/history (paging.Next; history uses Limit)
//
// total_count_skip:
//   - domains, lists, routes, ips/details
//
// total_skip:
//   - users, accounts/subaccounts, bounce-classification/domains
//
// limit_only:
//   - bounce-classification/stats (API has no skip/cursor; single page only)
//
// none:
//   - webhooks, alerts/settings, ip_pools, ip_whitelist, ips, keys, sandbox/auth_recipients,
//     thresholds/alerts/send, thresholds/hits, thresholds/limits, accounts/subaccounts/ip_pools,
//     domains/dynamic_pools/assignable, dynamic_pools
//
//nolint:gochecknoglobals
var objectReadPagination = map[string]readPagination{
	"templates":                        readPaginationPagingNext,
	"ip_warmups":                       readPaginationPagingNext,
	"forwards":                         readPaginationPagingNext,
	"alerts/slack/channels":            readPaginationPagingNext,
	"lists/pages":                      readPaginationPagingNext,
	"dkim/keys":                        readPaginationPagingNext,
	"dynamic_pools/domains":            readPaginationPagingNext,
	"dynamic_pools/history":            readPaginationPagingNext,
	"domains":                          readPaginationTotalCountSkip,
	"lists":                            readPaginationTotalCountSkip,
	"routes":                           readPaginationTotalCountSkip,
	"ips/details":                      readPaginationTotalCountSkip,
	"users":                            readPaginationTotalSkip,
	"accounts/subaccounts":             readPaginationTotalSkip,
	"bounce-classification/domains":    readPaginationTotalSkip,
	"bounce-classification/stats":      readPaginationLimitOnly,
	"webhooks":                         readPaginationNone,
	"alerts/settings":                  readPaginationNone,
	"ip_pools":                         readPaginationNone,
	"ip_whitelist":                     readPaginationNone,
	"ips":                              readPaginationNone,
	"keys":                             readPaginationNone,
	"sandbox/auth_recipients":          readPaginationNone,
	"thresholds/alerts/send":           readPaginationNone,
	"thresholds/hits":                  readPaginationNone,
	"thresholds/limits":                readPaginationNone,
	"accounts/subaccounts/ip_pools":    readPaginationNone,
	"domains/dynamic_pools/assignable": readPaginationNone,
	"dynamic_pools":                    readPaginationNone,
}

func paginationForObject(objectName string) readPagination {
	if strategy, ok := objectReadPagination[objectName]; ok {
		return strategy
	}

	return readPaginationNone
}

func supportedOperations() components.EndpointRegistryInput {
	readSupport := metadata.Schemas.ObjectNames().GetList(common.ModuleRoot)

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}
