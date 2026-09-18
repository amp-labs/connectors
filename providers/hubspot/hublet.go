package hubspot

import (
	"regexp"
	"strings"

	"github.com/amp-labs/connectors/providers"
)

// hubletPattern matches HubSpot's hublet identifiers: a region prefix followed by
// an index, such as na1 or eu1. Anything else is not something we can safely turn
// into a hostname.
var hubletPattern = regexp.MustCompile(`^[a-z]{2}\d{1,3}$`) //nolint:gochecknoglobals

// APIDomainForHublet maps a portal's data hosting location — HubSpot calls these
// "hublets" — to the API host that serves it.
//
// HubSpot shards portals across regional data centers and routes api.hubapi.com
// using routing information baked into the access token. Tokens minted before that
// routing existed resolve to na1, so a portal that lives in eu1 is answered with a
// 488 naming its real hublet until the request is addressed to api-eu1.hubapi.com.
// See https://product.hubspot.com/blog/routing-api-traffic.
//
// na1 keeps the unsuffixed host, which is what every connection uses today. An
// empty or unrecognized location falls back to the same host rather than inventing
// one: the fallback at worst reproduces current behavior, while a guessed hostname
// would break a connection that works.
func APIDomainForHublet(hublet string) string {
	normalized := strings.ToLower(strings.TrimSpace(hublet))

	if normalized == "" || normalized == defaultHublet || !hubletPattern.MatchString(normalized) {
		return providers.DefaultHubspotApiDomain
	}

	return "api-" + normalized + ".hubapi.com"
}

// defaultHublet is the region api.hubapi.com resolves to for a token that carries
// no hublet routing of its own.
const defaultHublet = "na1"
