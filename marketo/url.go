package marketo

import (
	"fmt"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

var restAPIPrefix string = "rest"

func (c *Connector) getURL(params common.ReadParams) (*urlbuilder.URL, error) {
	// make sure the object is in lowercase format.
	objName := strings.ToLower(params.ObjectName)

	bURL := strings.Join([]string{c.BaseURL, restAPIPrefix, c.Module, objName}, "/")
	bURL += ".json"

	link, err := urlbuilder.New(bURL)
	if err != nil {
		return nil, err
	}

	// This affects  a very few number of objects.
	// Leads, Deleted Leads, Lead Changes,
	if !params.Since.IsZero() {
		time := params.Since.Format(time.RFC3339)
		fmtTime := fmt.Sprintf("%v", time)
		link.WithQueryParam("sinceDatetime", fmtTime)
	}

	return link, nil
}
