package custom

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/xquery"
)

var (
	ErrMetadataCreate  = errors.New("metadata: SOAP request failed")
	ErrMetadataMarshal = errors.New("metadata: xml.MarshalIndent failed")
)

// performMetadataAPICall executes a Salesforce Metadata API operation using a typed request and response.
//
// Use this helper when working with structured metadata payloads (e.g., custom fields or permission definitions).
// It automatically handles SOAP envelope wrapping, session headers,
// and unmarshalling the response into the expected [R] type.
func performMetadataAPICall[R any](ctx context.Context, adapter *Adapter, payload any) (*R, error) {
	data, err := xml.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMetadataMarshal, err)
	}

	accessToken, present := common.GetAuthToken(ctx)
	if !present {
		return nil, common.ErrMissingAccessToken
	}

	response, err := adapter.performSOAPRequest(ctx, data, accessToken.String())
	if err != nil {
		return nil, err
	}

	var envelope Envelope[R]
	if err = xml.Unmarshal(response, &envelope); err != nil {
		return nil, err
	}

	return &envelope.Body, nil
}

type Envelope[B any] struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    B        `xml:"Body"`
}

// performSOAPRequest sends a SOAP request to the Salesforce API using the provided xmlPayload.
// The query in the payload determines the operation (e.g., read, create, update).
//
// A valid, non-expired accessToken must be passed directly.
// Each SOAP request must include a SessionHeader with the access token, as required by Salesforce.
// See: https://developer.salesforce.com/docs/atlas.en-us.api.meta/api/sforce_api_header_sessionheader.htm
//
// Returns the raw SOAP response body.
func (a *Adapter) performSOAPRequest(ctx context.Context, xmlPayload []byte, accessToken string) ([]byte, error) {
	body, err := xquery.NewXML(xmlPayload)
	if err != nil {
		return nil, err
	}

	body, err = putInsideEnvelope(body, accessToken)
	if err != nil {
		return nil, err
	}

	url, err := a.getSoapURL()
	if err != nil {
		return nil, err
	}

	resp, err := a.XMLClient.Post(ctx, url.String(), body, getSOAPHeaders()...)
	if err != nil {
		return nil, errors.Join(ErrMetadataCreate, err)
	}

	return []byte(resp.Body.RawXML()), nil
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
