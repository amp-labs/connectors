package metadata

import (
	"context"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/xquery"
)

var (
	ErrDeployFailed  = errors.New("metadata: deploy failed")
	ErrDeployTimeout = errors.New("metadata: deploy timed out")
)

const (
	deployPollInterval    = 5 * time.Second
	deployPollMaxDuration = 5 * time.Minute
)

// DeployResult contains the outcome of a Salesforce Metadata API deployment.
type DeployResult struct {
	Done    bool
	Status  string
	Success bool
	ID      string
}

// DeployMetadataZip deploys a zip package to Salesforce via the Metadata API SOAP deploy operation.
// The zip should contain package.xml and the metadata components to deploy.
// This is used for deploying APEX triggers and other components that require the deploy() API
// rather than the upsertMetadata() API.
func (a *Adapter) DeployMetadataZip(ctx context.Context, zipData []byte) error {
	accessToken, present := common.GetAuthToken(ctx)
	if !present {
		return common.ErrMissingAccessToken
	}

	token := accessToken.String()

	deployID, err := a.deploy(ctx, token, zipData)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrDeployFailed, err)
	}

	result, err := a.pollDeployStatus(ctx, token, deployID)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrDeployFailed, err)
	}

	if !result.Success {
		return fmt.Errorf("%w: status=%s", ErrDeployFailed, result.Status)
	}

	return nil
}

// deploy sends a SOAP deploy request with the base64-encoded zip to the Metadata API.
// Returns the async deployment ID for status polling.
func (a *Adapter) deploy(ctx context.Context, accessToken string, zipData []byte) (string, error) {
	encodedZip := base64.StdEncoding.EncodeToString(zipData)

	payload := fmt.Sprintf(`<md:deploy xmlns:md="http://soap.sforce.com/2006/04/metadata">
  <md:ZipFile>%s</md:ZipFile>
  <md:DeployOptions>
    <md:allowMissingFiles>false</md:allowMissingFiles>
    <md:autoUpdatePackage>false</md:autoUpdatePackage>
    <md:checkOnly>false</md:checkOnly>
    <md:ignoreWarnings>false</md:ignoreWarnings>
    <md:performRetrieve>false</md:performRetrieve>
    <md:purgeOnDelete>false</md:purgeOnDelete>
    <md:rollbackOnError>true</md:rollbackOnError>
    <md:singlePackage>true</md:singlePackage>
    <md:testLevel>NoTestRun</md:testLevel>
  </md:DeployOptions>
</md:deploy>`, encodedZip)

	respBytes, err := a.performDeploySOAPRequest(ctx, []byte(payload), accessToken)
	if err != nil {
		return "", err
	}

	var resp deployResponse
	if err := xml.Unmarshal(respBytes, &resp); err != nil {
		return "", fmt.Errorf("failed to parse deploy response: %w", err)
	}

	return resp.Body.DeployResponse.Result.ID, nil
}

// pollDeployStatus polls the deployment status until completion or timeout.
func (a *Adapter) pollDeployStatus(ctx context.Context, accessToken, deployID string) (*DeployResult, error) {
	start := time.Now()

	for time.Since(start) < deployPollMaxDuration {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(deployPollInterval):
		}

		payload := fmt.Sprintf(`<md:checkDeployStatus xmlns:md="http://soap.sforce.com/2006/04/metadata">
  <md:asyncProcessId>%s</md:asyncProcessId>
  <md:includeDetails>true</md:includeDetails>
</md:checkDeployStatus>`, deployID)

		respBytes, err := a.performDeploySOAPRequest(ctx, []byte(payload), accessToken)
		if err != nil {
			return nil, err
		}

		var resp checkDeployStatusResponse
		if err := xml.Unmarshal(respBytes, &resp); err != nil {
			return nil, fmt.Errorf("failed to parse deploy status response: %w", err)
		}

		result := &resp.Body.CheckDeployStatusResponse.Result
		if result.Done {
			return &DeployResult{
				Done:    result.Done,
				Status:  result.Status,
				Success: result.Success,
				ID:      result.ID,
			}, nil
		}
	}

	return nil, ErrDeployTimeout
}

// performDeploySOAPRequest sends a raw SOAP request for deploy operations.
// Unlike performSOAPRequest, this uses a deploy-specific envelope without AllOrNoneHeader.
func (a *Adapter) performDeploySOAPRequest(ctx context.Context, xmlPayload []byte, accessToken string) ([]byte, error) {
	body, err := xquery.NewXML(xmlPayload)
	if err != nil {
		return nil, err
	}

	body, err = putInsideDeployEnvelope(body, accessToken)
	if err != nil {
		return nil, err
	}

	url, err := a.getSoapURL()
	if err != nil {
		return nil, err
	}

	resp, err := a.XMLClient.Post(ctx, url.String(), body, getDeploySOAPHeaders()...)
	if err != nil {
		return nil, errors.Join(ErrDeployFailed, err)
	}

	return []byte(resp.Body.RawXML()), nil
}

func putInsideDeployEnvelope(content *xquery.XML, accessToken string) (*xquery.XML, error) {
	template := `<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
                  xmlns:md="http://soap.sforce.com/2006/04/metadata">
    <soapenv:Header>
        <md:SessionHeader>
            <md:sessionId>TODO----accessToken</md:sessionId>
        </md:SessionHeader>
    </soapenv:Header>
    <soapenv:Body/>
</soapenv:Envelope>`

	envelope, err := xquery.NewXML([]byte(template))
	if err != nil {
		return nil, err
	}

	session := envelope.FindOne("//md:sessionId").GetChild()
	session.SetDataText(accessToken)

	envelope.FindOne("//soapenv:Body").SetDataNode(content)

	return envelope, nil
}

func getDeploySOAPHeaders() []common.Header {
	return []common.Header{
		{Key: "Content-Type", Value: "text/xml; charset=UTF-8"},
		{Key: "SOAPAction", Value: "deploy"},
	}
}

// XML types for deploy SOAP responses.
type deployResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		DeployResponse struct {
			Result struct {
				ID   string `xml:"id"`
				Done bool   `xml:"done"`
			} `xml:"result"`
		} `xml:"deployResponse"`
	} `xml:"Body"`
}

type checkDeployStatusResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		CheckDeployStatusResponse struct {
			Result struct {
				Done    bool   `xml:"done"`
				Status  string `xml:"status"`
				Success bool   `xml:"success"`
				ID      string `xml:"id"`
			} `xml:"result"`
		} `xml:"checkDeployStatusResponse"`
	} `xml:"Body"`
}
