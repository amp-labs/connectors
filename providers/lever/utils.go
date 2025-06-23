package lever

import (
	"fmt"
	"net/url"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	apiVersion      = "v1"
	DefaultPageSize = 100
)

var (
	EndpointWithCreatedAtRange = datautils.NewSet( //nolint:gochecknoglobals
		"audit_events",
		"requisitions",
	)

	EndpointWithUpdatedAtRange = datautils.NewSet( //nolint:gochecknoglobals
		"postings",
		"opportunities",
	)

	EndpointWithUploadedAtRange = datautils.NewSet( //nolint:gochecknoglobals
		"resumes",
		"files",
	)

	EndpointWithOpportunityID = datautils.NewSet( //nolint:gochecknoglobals
		"feedback",
		"files",
		"interviews",
		"notes",
		"offers",
		"panels",
		"forms",
		"referrals",
		"resumes",
		"addLinks",
		"removeLinks",
		"addTags",
		"removeTags",
		"addSources",
		"removeSources",
		"stage",
		"archived",
		"apply",
	)

	// Below endpoints having PUT method but no recordID.
	EndpointWithPutMethodNoRecordId = datautils.NewSet( //nolint:gochecknoglobals
		"stage",
		"archived",
	)

	// Below endpoints having "userId" in the URL Path.
	EndpointWithUserId = datautils.NewSet( //nolint:gochecknoglobals
		"deactivate",
		"reactivate",
	)

	// Below write endpoints requires QueryParam "perform_as" in the URL.
	EndPointWithPerformAsQueryParam = datautils.NewSet( //nolint:gochecknoglobals
		"feedback",
		"files",
		"interviews",
		"panels",
		"opportunities",
		"postings",
	)
)

func (c *Connector) constructURL(objName string) string {
	if EndpointWithOpportunityID.Has(objName) {
		return fmt.Sprintf("opportunities/%s/%s", c.opportunityId, objName)
	}

	if EndpointWithUserId.Has(objName) {
		return fmt.Sprintf("users/%s/%s", c.userId, objName)
	}

	if objName == "apply" {
		return fmt.Sprintf("postings/%s/%s", c.postingId, objName)
	}

	return objName
}

func makeNextRecordsURL(reqLink *url.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		url, err := urlbuilder.FromRawURL(reqLink)
		if err != nil {
			return "", err
		}

		hasNextPage, err := jsonquery.New(node).BoolWithDefault("hasNext", false)
		if err != nil {
			return "", err
		}

		if hasNextPage {
			pagination, err := jsonquery.New(node).StringRequired("next")
			if err != nil {
				return "", err
			}

			url.WithQueryParam("offset", pagination)

			return url.String(), nil
		}

		return "", nil
	}
}
