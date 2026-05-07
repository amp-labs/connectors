package metadata

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
)

// TestLevel maps to Salesforce Metadata API DeployOptions.testLevel.
// See: https://developer.salesforce.com/docs/atlas.en-us.api_meta.meta/api_meta/meta_deploy.htm
type TestLevel string

const (
	// TestLevelNoTestRun runs no tests. Default; only valid for non-production deploys.
	TestLevelNoTestRun TestLevel = "NoTestRun"
	// TestLevelRunSpecifiedTests runs the test classes named in DeployOptions.runTests.
	// At least one runTests entry is required when this level is selected.
	TestLevelRunSpecifiedTests TestLevel = "RunSpecifiedTests"
	// TestLevelRunLocalTests runs all non-managed tests in the org.
	TestLevelRunLocalTests TestLevel = "RunLocalTests"
	// TestLevelRunAllTestsInOrg runs every test in the org including managed packages.
	TestLevelRunAllTestsInOrg TestLevel = "RunAllTestsInOrg"
)

var ErrDeployFailed = errors.New("metadata: deploy failed")

// DeployResult contains the outcome of a Salesforce Metadata API deployment.
type DeployResult struct {
	Done              bool
	Status            string
	Success           bool
	ID                string
	ErrorMessage      string
	ComponentFailures []ComponentFailure
}

// ComponentFailure describes a single component failure in a deployment.
type ComponentFailure struct {
	ComponentType string
	FullName      string
	Problem       string
	ProblemType   string
}

// DeployMetadataZip initiates a deploy of a zip package to Salesforce via the Metadata API
// SOAP deploy operation with testLevel=NoTestRun. Returns the async deployment ID for
// status polling. Use CheckDeployStatus to poll for completion.
func (a *Adapter) DeployMetadataZip(ctx context.Context, zipData []byte) (string, error) {
	return a.DeployMetadataZipWithTests(ctx, zipData, TestLevelNoTestRun, nil)
}

// DeployMetadataZipWithTests initiates a deploy of a zip package to Salesforce via the
// Metadata API SOAP deploy operation with the supplied testLevel. When testLevel is
// RunSpecifiedTests, runTests must contain at least one Apex test class name that
// exists in the org (or is included in the same zip).
//
// Salesforce requires testLevel=RunSpecifiedTests, RunLocalTests, or RunAllTestsInOrg
// for deploys to production orgs; sandbox/dev deploys may use NoTestRun.
func (a *Adapter) DeployMetadataZipWithTests(
	ctx context.Context, zipData []byte, testLevel TestLevel, runTests []string,
) (string, error) {
	deployID, err := a.deploy(ctx, zipData, testLevel, runTests)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrDeployFailed, err)
	}

	return deployID, nil
}

// CheckDeployStatus checks the status of an async deployment once and returns the result.
// The caller is responsible for polling in a loop until Done is true.
func (a *Adapter) CheckDeployStatus(ctx context.Context, deployID string) (*DeployResult, error) {
	payload := fmt.Sprintf(`<md:checkDeployStatus xmlns:md="http://soap.sforce.com/2006/04/metadata">
  <md:asyncProcessId>%s</md:asyncProcessId>
  <md:includeDetails>true</md:includeDetails>
</md:checkDeployStatus>`, deployID)

	resp, err := performDeploySOAPRequest[checkDeployStatusResponse](ctx, a, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to parse deploy status response: %w", err)
	}

	result := &resp.CheckDeployStatusResponse.Result

	failures := make([]ComponentFailure, len(result.Details.ComponentFailures))
	for i, cf := range result.Details.ComponentFailures {
		failures[i] = ComponentFailure{
			ComponentType: cf.ComponentType,
			FullName:      cf.FullName,
			Problem:       cf.Problem,
			ProblemType:   cf.ProblemType,
		}
	}

	return &DeployResult{
		Done:              result.Done,
		Status:            result.Status,
		Success:           result.Success,
		ID:                result.ID,
		ErrorMessage:      result.ErrorMessage,
		ComponentFailures: failures,
	}, nil
}

// deploy sends a SOAP deploy request with the base64-encoded zip to the Metadata API.
// Returns the async deployment ID for status polling.
//
// When testLevel is RunSpecifiedTests, the SOAP body emits one <md:runTests> element
// per entry in runTests. The Salesforce SOAP API requires this element to appear once
// per test class to run; ordering of DeployOptions sub-elements follows the WSDL.
func (a *Adapter) deploy(
	ctx context.Context, zipData []byte, testLevel TestLevel, runTests []string,
) (string, error) {
	encodedZip := base64.StdEncoding.EncodeToString(zipData)

	if testLevel == "" {
		testLevel = TestLevelNoTestRun
	}

	var runTestsXML strings.Builder

	if testLevel == TestLevelRunSpecifiedTests {
		for _, name := range runTests {
			runTestsXML.WriteString("    <md:runTests>")
			runTestsXML.WriteString(name)
			runTestsXML.WriteString("</md:runTests>\n")
		}
	}

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
%s    <md:singlePackage>true</md:singlePackage>
    <md:testLevel>%s</md:testLevel>
  </md:DeployOptions>
</md:deploy>`, encodedZip, runTestsXML.String(), testLevel)

	resp, err := performDeploySOAPRequest[deployResponse](ctx, a, payload)
	if err != nil {
		return "", fmt.Errorf("failed to parse deploy response: %w", err)
	}

	return resp.DeployResponse.Result.ID, nil
}

func getDeploySOAPHeaders() []common.Header {
	return []common.Header{
		{Key: "Content-Type", Value: "text/xml; charset=UTF-8"},
		{Key: "SOAPAction", Value: "deploy"},
	}
}

// XML types for deploy SOAP responses (body content only, wrapped by Envelope[R]).
type deployResponse struct {
	DeployResponse struct {
		Result struct {
			ID   string `xml:"id"`
			Done bool   `xml:"done"`
		} `xml:"result"`
	} `xml:"deployResponse"`
}

type checkDeployStatusResponse struct {
	CheckDeployStatusResponse struct {
		Result struct {
			Done         bool   `xml:"done"`
			Status       string `xml:"status"`
			Success      bool   `xml:"success"`
			ID           string `xml:"id"`
			ErrorMessage string `xml:"errorMessage"`
			Details      struct {
				ComponentFailures []struct {
					ComponentType string `xml:"componentType"`
					FullName      string `xml:"fullName"`
					Problem       string `xml:"problem"`
					ProblemType   string `xml:"problemType"`
				} `xml:"componentFailures"`
			} `xml:"details"`
		} `xml:"result"`
	} `xml:"checkDeployStatusResponse"`
}
