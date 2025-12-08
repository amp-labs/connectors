package ads

import (
	"context"
	_ "embed"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	liinternal "github.com/amp-labs/connectors/providers/linkedin/internal/linkedininternal"
)

type responseObject struct {
	Elements []map[string]any `json:"elements"`
}

func (c *Adapter) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := c.constructURL(objectName)
	if err != nil {
		return nil, err
	}

	if objectsWithSearchQueryParam.Has(objectName) {
		// For dmpSegments, metadata is fetched based on the associated ad account.
		// nolint:lll
		// Refer https://learn.microsoft.com/en-us/linkedin/marketing/matched-audiences/create-and-manage-segments?view=li-lms-2025-08&tabs=http#find-dmp-segments-by-account.
		if objectName == "dmpSegments" {
			url.WithQueryParam("q", "account")

			url.WithUnencodedQueryParam("account", "urn%3Ali%3AsponsoredAccount%3A"+c.AdAccountId)
		} else {
			url.WithQueryParam("q", "search")
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("LinkedIn-Version", liinternal.LinkedInVersion) // nolint:canonicalheader
	req.Header.Add("X-Restli-Protocol-Version", liinternal.ProtocolVersion)

	return req, nil
}

func (c *Adapter) parseSingleObjectMetadataResponse(
	ctx context.Context,
	objectName string,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ObjectMetadata, error) {
	objectMetadata := common.ObjectMetadata{
		Fields:      make(map[string]common.FieldMetadata),
		DisplayName: naming.CapitalizeFirstLetter(objectName),
	}

	data, err := common.UnmarshalJSON[responseObject](response)
	if err != nil {
		return nil, common.ErrFailedToUnmarshalBody
	}

	if len(data.Elements) == 0 {
		return nil, liinternal.ErrMetadataNotFound
	}

	// Using the first result data to generate the metadata.
	for field := range data.Elements[0] {
		objectMetadata.Fields[field] = common.FieldMetadata{
			DisplayName:  field,
			ValueType:    common.ValueTypeOther,
			ProviderType: "",
			Values:       nil,
		}
	}

	return &objectMetadata, nil
}
