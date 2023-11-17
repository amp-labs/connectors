package salesforce

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

// BulkWriteParams defines how we are writing data to a SaaS API.
type BulkWriteParams struct {
	// The name of the object we are writing, e.g. "Account"
	ObjectName string // required

	// The external ID of the object instance we are updating. Provided in the case of UPDATE, but not CREATE.
	ExternalId string // required

	// The path to the CSV file we are writing
	FilePath string // required
}

// BulkWriteResult is what's returned from writing data via the BulkWrite call.
type BulkWriteResult struct {
	// State is the state of the bulk job process
	State string `json:"state"`
	// JobId is the ID of the bulk job process
	JobId string `json:"jobId"`
}

type GetJobInfoResult struct {
	Id                     string  `json:"id"`
	Object                 string  `json:"object"`
	CreatedById            string  `json:"createdById"`
	ExternalIdFieldName    string  `json:"externalIdFieldName"`
	State                  string  `json:"state"`
	Operation              string  `json:"operation"`
	ColumnDelimiter        string  `json:"columnDelimiter"`
	LineEnding             string  `json:"lineEnding"`
	NumberRecordsFailed    float64 `json:"numberRecordsFailed"`
	NumberRecordsProcessed float64 `json:"numberRecordsProcessed"`

	ApexProcessingTime      float64 `json:"apexProcessingTime"`
	ApiActiveProcessingTime float64 `json:"apiActiveProcessingTime"`
	ApiVersion              float64 `json:"apiVersion"`
	ConcurrencyMode         string  `json:"concurrencyMode"`
	ContentType             string  `json:"contentType"`
	CreatedDate             string  `json:"createdDate"`
	JobType                 string  `json:"jobType"`
	Retries                 float64 `json:"retries"`
	SystemModstamp          string  `json:"systemModstamp"`
	TotalProcessingTime     float64 `json:"totalProcessingTime"`
}

var (
	ErrKeyNotFound     = errors.New("key not found")
	ErrUnknownNodeType = errors.New("unknown node type when parsing JSON")
	ErrInvalidType     = errors.New("invalid type")
	ErrInvalidJobState = errors.New("invalid job state")
)

func (c *Connector) BulkWrite( //nolint:funlen,cyclop
	ctx context.Context,
	config BulkWriteParams,
) (*BulkWriteResult, error) {
	// cretes batch upload job, returns json with id and other info
	res, err := c.createJob(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("createJob failed: %w", err)
	}

	resObject, err := res.GetObject()
	if err != nil {
		return nil, fmt.Errorf("parsing result of createJob failed: %w", errors.Join(err, common.ErrParseError))
	}

	state, err := resObject["state"].GetString()
	if err != nil {
		unpacked, _ := resObject["state"].Unpack()

		return nil, fmt.Errorf(
			"%w: expected salesforce job state to be string in response, got %T",
			ErrInvalidType,
			unpacked,
		)
	}

	if strings.ToLower(state) != "open" {
		return nil, fmt.Errorf("%w: expected job state to be open, got %s", ErrInvalidJobState, state)
	}

	jobId, err := resObject["id"].GetString()
	if err != nil {
		unpacked, _ := resObject["id"].Unpack()

		return nil, fmt.Errorf(
			"%w: expected salesforce job id to be string in response, got %T",
			ErrInvalidType,
			unpacked)
	}

	// upload csv and there is no response body other than status code
	_, err = c.uploadCSV(ctx, jobId, config)
	if err != nil {
		return nil, fmt.Errorf("uploadCSV failed: %w", err)
	}

	data, err := c.completeUpload(ctx, jobId)
	if err != nil {
		return nil, fmt.Errorf("completeUpload failed: %w", err)
	}

	dataObject, err := data.GetObject()
	if err != nil {
		return nil, fmt.Errorf("parsing result of completeUpload failed: %w", errors.Join(err, common.ErrParseError))
	}

	updatedJobId, err := dataObject["id"].GetString()
	if err != nil {
		unpacked, _ := dataObject["id"].Unpack()

		return nil, fmt.Errorf(
			"%w. expected salesforce job id to be string in response, got %T",
			ErrInvalidType,
			unpacked,
		)
	}

	completeState, err := dataObject["state"].GetString() //nolint:varnamelen
	if err != nil {
		unpacked, _ := resObject["state"].Unpack()

		return nil, fmt.Errorf(
			"%w. expected salesforce job state to be string in response, got %T",
			ErrInvalidType,
			unpacked,
		)
	}

	return &BulkWriteResult{
		JobId: updatedJobId,
		State: completeState,
	}, nil
}

func joinURLPath(baseURL string, paths ...string) (string, error) {
	location, err := url.JoinPath(baseURL, paths...)
	if err != nil {
		return "", errors.Join(err, common.ErrInvalidPathJoin)
	}

	return location, nil
}

func (c *Connector) createJob(ctx context.Context, config BulkWriteParams) (*ajson.Node, error) {
	location, err := joinURLPath(c.BaseURL, "jobs/ingest")
	if err != nil {
		return nil, err
	}

	body := map[string]interface{}{
		"object":              config.ObjectName,
		"externalIdFieldName": config.ExternalId,
		"contentType":         "CSV",
		"operation":           "upsert",
		"lineEnding":          "LF",
	}

	return c.post(ctx, location, body)
}

func (c *Connector) uploadCSV(ctx context.Context, jobId string, config BulkWriteParams) ([]byte, error) {
	file, err := os.ReadFile(config.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to upload CSV for %s: %w", config.FilePath, errors.Join(err, common.ErrReadFile))
	}

	location, err := joinURLPath(c.BaseURL, fmt.Sprintf("jobs/ingest/%s/batches", jobId))
	if err != nil {
		return nil, err
	}

	return c.putCSV(ctx, location, file)
}

func (c *Connector) completeUpload(ctx context.Context, jobId string) (*ajson.Node, error) {
	updateLoadCompleteBody := map[string]interface{}{
		"state": "UploadComplete",
	}

	location, err := joinURLPath(c.BaseURL, fmt.Sprintf("jobs/ingest/%s", jobId))
	if err != nil {
		return nil, err
	}

	return c.patch(ctx, location, updateLoadCompleteBody)
}

func (c *Connector) GetJobInfo(ctx context.Context, jobId string) (*GetJobInfoResult, error) {
	location, err := joinURLPath(c.BaseURL, "jobs/ingest", jobId)
	if err != nil {
		return nil, err
	}

	node, err := c.get(ctx, location)
	if err != nil {
		return nil, fmt.Errorf("getGetInfo failed: %w", errors.Join(err, common.ErrRequestFailed))
	}

	data, err := ajson.Marshal(node)
	if err != nil {
		return nil, fmt.Errorf("marshalling result of getGetInfo failed: %w", errors.Join(err, common.ErrParseError))
	}

	var info *GetJobInfoResult
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("unmarshalling result of getGetInfo failed: %w", errors.Join(err, common.ErrParseError))
	}

	return info, nil
}

/*
// Below are methods that will help us get results, but need CSVHTTPClient to be implemented
// Below code will work once CSVHTTPClient is implemented
// TODO: implement CSVHTTPClient

func (c *Connector) GetJobResult(ctx context.Context, jobId string) ([]byte, error) {
	location, err := joinURLPath(c.BaseURL, fmt.Sprintf("jobs/ingest/%s/successfulResults", jobId))
	if err != nil {
		return nil, err
	}

	return c.getCSV(ctx, location)
}

func (c *Connector) GetUnprocessedResults(ctx context.Context, jobId string) ([]byte, error) {
	location, err := joinURLPath(c.BaseURL, fmt.Sprintf("jobs/ingest/%s/unprocessedrecords", jobId))
	if err != nil {
		return nil, err
	}

	return c.getCSV(ctx, location)
}

func (c *Connector) GetFailedResults(ctx context.Context, jobId string) ([]byte, error) {
	location, err := joinURLPath(c.BaseURL, fmt.Sprintf("jobs/ingest/%s/failedResults", jobId))
	if err != nil {
		return nil, err
	}

	return c.getCSV(ctx, location)
}
*/
