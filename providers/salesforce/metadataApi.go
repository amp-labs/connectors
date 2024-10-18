package salesforce

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/xquery"
)

var ErrCreateMetadata = errors.New("error in CreateMetadata")

// CreateMetadata creates custom metadata.
// Requires non-expired access token to be passed directly.
// According to documentation every XML type of request must include SessionHeader with access token.
// See: https://developer.salesforce.com/docs/atlas.en-us.api.meta/api/sforce_api_header_sessionheader.htm.
func (c *Connector) CreateMetadata(ctx context.Context, data []byte, accessToken string) (string, error) {
	body, err := xquery.NewXML(data)
	if err != nil {
		return "", err
	}

	body, err = putInsideEnvelope(body, accessToken)
	if err != nil {
		return "", err
	}

	url, err := c.getSoapURL()
	if err != nil {
		return "", err
	}

	resp, err := c.XML.Post(ctx, url.String(), body, getSOAPHeaders()...)
	if err != nil {
		return "", errors.Join(ErrCreateMetadata, err)
	}

	return resp.Body.RawXML(), nil
}

func putInsideEnvelope(content *xquery.XML, accessToken string) (*xquery.XML, error) {
	template := `
<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsd="http://www.w3.org/2001/XMLSchema"
                  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
    <soapenv:Header xmlns="http://soap.sforce.com/2006/04/metadata">
        <AllOrNoneHeader>
            <allOrNone>true</allOrNone>
        </AllOrNoneHeader>
        <SessionHeader>
            <sessionId>TODO----accessToken</sessionId>
        </SessionHeader>
    </soapenv:Header>
    <soapenv:Body xmlns="http://soap.sforce.com/2006/04/metadata"/>
</soapenv:Envelope>
`

	envelope, err := xquery.NewXML([]byte(template))
	if err != nil {
		return nil, err
	}

	session := envelope.FindOne("//sessionId").GetChild()
	session.SetDataText(accessToken)
	// Store user passed data within body tag.
	envelope.FindOne("//soapenv:Body").SetDataNode(content)

	return envelope, nil
}

func getSOAPHeaders() []common.Header {
	// SOAP API definition specifies that SOAPAction header should be empty string
	// but if we set to "", API will error, so we use "''" instead as a workaround.
	//
	// For related information you can read Salesforce stackexchange:
	// https://salesforce.stackexchange.com/a/49273
	//
	// The SOAP API spec states that missing value of a header is compensated by information found in URI.
	// But can be used by server side for routing purposes.
	// https://www.w3.org/TR/2000/NOTE-SOAP-20000508/#_Toc478383528
	return []common.Header{{
		Key:   "Content-Type",
		Value: "text/xml",
	}, {
		Key:   "SOAPAction",
		Value: "''",
	}}
}
