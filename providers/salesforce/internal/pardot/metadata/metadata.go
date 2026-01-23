package metadata

import (
	"context"
	"encoding/xml"
	"fmt"

	"github.com/amp-labs/connectors"
)

func (a *Strategy) ListObjectMetadata(ctx context.Context, objectNames []string) (*connectors.ListObjectMetadataResult, error) {
	reqs := make([]ObjectDefinitionRequest, 0, len(objectNames))
	for _, name := range objectNames {
		reqs = append(reqs, ObjectDefinitionRequest{
			ObjectType: name,
		})
	}

	payload := DefinitionRequestMsg{
		XmlnsPar:         "http://exacttarget.com/wsdl/partnerAPI",
		DescribeRequests: reqs,
	}

	resp, err := performMetadataAPICall[any](ctx, a, payload)
	if err != nil {
		return nil, err
	}

	fmt.Println(resp)

	return nil, nil
}

type DefinitionRequestMsg struct {
	XMLName          xml.Name                  `xml:"par:DefinitionRequestMsg"`
	XmlnsPar         string                    `xml:"xmlns:par,attr"`
	DescribeRequests []ObjectDefinitionRequest `xml:"par:DescribeRequests>par:ObjectDefinitionRequest"`
}

type ObjectDefinitionRequest struct {
	ObjectType string `xml:"par:ObjectType"`
}
