package metadata

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

func performMetadataAPICall[R any](ctx context.Context, adapter *Strategy, payload any) (*R, error) {
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

func (a *Strategy) performSOAPRequest(ctx context.Context, xmlPayload []byte, accessToken string) ([]byte, error) {
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

	resp, err := a.xmlClient.Post(ctx, url.String(), body)
	if err != nil {
		return nil, errors.Join(ErrMetadataCreate, err)
	}

	return []byte(resp.Body.RawXML()), nil
}

func putInsideEnvelope(content *xquery.XML, accessToken string) (*xquery.XML, error) {
	template := `
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
                  xmlns:par="http://exacttarget.com/wsdl/partnerAPI">
    <soapenv:Header>
        <fueloauth xmlns="http://exacttarget.com/wsdl/partnerAPI">
            YOUR_OAUTH_TOKEN_HERE
        </fueloauth>
    </soapenv:Header>
    <soapenv:Body></soapenv:Body>
</soapenv:Envelope>
`

	envelope, err := xquery.NewXML([]byte(template))
	if err != nil {
		return nil, err
	}

	session := envelope.FindOne("//fueloauth").GetChild()
	session.SetDataText(accessToken)
	// Store user passed data within body tag.
	envelope.FindOne("//soapenv:Body").SetDataNode(content)

	return envelope, nil
}
