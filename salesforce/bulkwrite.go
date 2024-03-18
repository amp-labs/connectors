package salesforce

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/amp-labs/connectors/common"
)

const (
	Insert BulkWriteMode = "insert"
	Upsert BulkWriteMode = "upsert"
	Update BulkWriteMode = "update"

	JobStateAborted        = "Aborted"
	JobStateFailed         = "Failed"
	JobStateComplete       = "JobComplete"
	JobStateInProgress     = "InProgress"
	JobStateUploadComplete = "UploadComplete"

	sfIdFieldName    = "sf__Id"
	sfErrorFieldName = "sf__Error"
)

var (
	ErrKeyNotFound      = errors.New("key not found")
	ErrUnknownNodeType  = errors.New("unknown node type when parsing JSON")
	ErrInvalidType      = errors.New("invalid type")
	ErrInvalidJobState  = errors.New("invalid job state")
	ErrUnsupportedMode  = errors.New("unsupported mode")
	ErrReadToByteFailed = errors.New("failed to read data to bytes")
)

// BulkWriteParams defines how we are writing data to a SaaS API.
type BulkWriteParams struct {
	// The name of the object we are writing, e.g. "Account"
	ObjectName string // required

	// The name of a field on the object which is an External ID. Provided in the case of upserts, not inserts
	ExternalIdField string // required

	// The path to the CSV file we are writing
	CSVData io.Reader // required

	// Salesforce operation mode
	Mode BulkWriteMode
}

type BulkWriteMode string

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

type FailInfo struct {
	FailureType   string              `json:"failureType"`
	FailedUpdates map[string][]string `json:"failedUpdates,omitempty"`
	FailedCreates map[string][]string `json:"failedCreates,omitempty"`
	Reason        string              `json:"reason,omitempty"`
}

type JobResults struct {
	JobId          string            `json:"jobId"`
	State          string            `json:"state"`
	FailureDetails *FailInfo         `json:"failureDetails,omitempty"`
	JobInfo        *GetJobInfoResult `json:"jobInfo,omitempty"`
	Message        string            `json:"message,omitempty"`
}

func (c *Connector) BulkWrite( //nolint:funlen,cyclop
	ctx context.Context,
	config BulkWriteParams,
) (*BulkWriteResult, error) {
	// Only support upsert for now
	switch config.Mode {
	case Upsert:
		break
	case Insert:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedMode, "insert")
	case Update:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedMode, "Update")
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedMode, config.Mode)
	}

	// cretes batch upload job, returns json with id and other info
	res, err := c.createJob(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("createJob failed: %w", err)
	}

	resObject, err := res.Body.GetObject()
	if err != nil {
		return nil, fmt.Errorf(
			"parsing result of createJob failed: %w",
			errors.Join(err, common.ErrParseError),
		)
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

	dataObject, err := data.Body.GetObject()
	if err != nil {
		return nil, fmt.Errorf("parsing result of completeUpload failed: %w", errors.Join(err, common.ErrParseError))
	}

	updatedJobId, err := dataObject["id"].GetString()
	if err != nil {
		unpacked, _ := dataObject["id"].Unpack()

		return nil, fmt.Errorf(
			"%w: expected salesforce job id to be string in response, got %T",
			ErrInvalidType,
			unpacked,
		)
	}

	completeState, err := dataObject["state"].GetString() //nolint:varnamelen
	if err != nil {
		unpacked, _ := resObject["state"].Unpack()

		return nil, fmt.Errorf(
			"%w: expected salesforce job state to be string in response, got %T",
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

func (c *Connector) createJob(ctx context.Context, config BulkWriteParams) (*common.JSONHTTPResponse, error) {
	location, err := joinURLPath(c.BaseURL, "jobs/ingest")
	if err != nil {
		return nil, err
	}

	body := map[string]interface{}{
		"object":              config.ObjectName,
		"externalIdFieldName": config.ExternalIdField,
		"contentType":         "CSV",
		"operation":           "upsert",
		"lineEnding":          "LF",
	}

	return c.post(ctx, location, body)
}

func (c *Connector) uploadCSV(ctx context.Context, jobId string, config BulkWriteParams) ([]byte, error) {
	location, err := joinURLPath(c.BaseURL, fmt.Sprintf("jobs/ingest/%s/batches", jobId))
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(config.CSVData)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to upload CSV to salesforce: %w",
			errors.Join(ErrReadToByteFailed, err),
		)
	}

	return c.putCSV(ctx, location, data)
}

func (c *Connector) completeUpload(ctx context.Context, jobId string) (*common.JSONHTTPResponse, error) {
	updateLoadCompleteBody := map[string]interface{}{
		"state": JobStateUploadComplete,
	}

	location, err := joinURLPath(c.BaseURL, "jobs/ingest/"+jobId)
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

	rsp, err := c.get(ctx, location)
	if err != nil {
		return nil, fmt.Errorf("getGetInfo failed: %w", errors.Join(err, common.ErrRequestFailed))
	}

	info, err := common.UnmarshalJSON[GetJobInfoResult](rsp)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling result of getGetInfo failed: %w", errors.Join(err, common.ErrParseError))
	}

	return info, nil
}

func (c *Connector) GetJobResults(ctx context.Context, jobId string) (*JobResults, error) {
	jobInfo, err := c.GetJobInfo(ctx, jobId)
	if err != nil {
		return nil, fmt.Errorf("failed to get job information: %w", err)
	}

	if jobInfo.State != JobStateComplete {
		// Take care of failed, aborted, in progress, and upload complete cases
		// We don't need to query Salesforce for these cases
		return c.getIncompleteJobResults(jobInfo), err
	}

	if jobInfo.State == JobStateComplete && jobInfo.NumberRecordsFailed == 0 {
		// Complete success case, no need to query Salesforce
		return &JobResults{
			JobId:   jobInfo.Id,
			State:   jobInfo.State,
			JobInfo: jobInfo,
		}, nil
	}

	return c.getPartialFailureDetails(ctx, jobInfo)
}

func (c *Connector) getJobResults(ctx context.Context, jobId string) (*http.Response, error) {
	location, err := joinURLPath(c.BaseURL, fmt.Sprintf("jobs/ingest/%s/failedResults", jobId))
	if err != nil {
		return nil, err
	}

	req, err := common.MakeJSONGetRequest(ctx, location, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get request: %w", err)
	}

	// Get the connector's JSONHTTPClient, which is a special HTTPClient that handles JSON responses,
	// and use it's underlying http.Client to make the request.
	return c.Client.HTTPClient.Client.Do(req)
}

//nolint:funlen
func (c *Connector) getPartialFailureDetails(ctx context.Context, jobInfo *GetJobInfoResult) (*JobResults, error) {
	// Query Salesforce to get partial failure details
	res, err := c.getJobResults(ctx, jobInfo.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get job results: %w", err)
	}
	defer res.Body.Close()
	// if partially failed, we need to get the error message for each record
	reader := csv.NewReader(res.Body)

	failInfo := &FailInfo{
		FailureType:   "Partial",
		FailedUpdates: make(map[string][]string),
		FailedCreates: make(map[string][]string),
	}

	var sfIdColIdx, sfErrorColIdx, externalIdColIdx int

	rowIdx := 0

	for {
		record, err := reader.Read()
		if err != nil {
			// end of file
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, err
		}

		// Get column index of sf__Id, sf__Error, and externalIdFieldName in header row
		// Salesforce API responses may not be consistent with the order of columns, so we need to get the index
		if rowIdx == 0 {
			indiceMap, err := getColumnIndice(record, []string{sfIdFieldName, sfErrorFieldName, jobInfo.ExternalIdFieldName})
			if err != nil {
				return nil, err
			}

			sfIdColIdx = indiceMap[sfIdFieldName]
			sfErrorColIdx = indiceMap[sfErrorFieldName]
			externalIdColIdx = indiceMap[jobInfo.ExternalIdFieldName]
			rowIdx++

			continue
		}

		sfId := record[sfIdColIdx]
		failureMap := failInfo.FailedUpdates
		errMsg := record[sfErrorColIdx]
		externalId := record[externalIdColIdx]

		if sfId == "" {
			// If sf__Id is empty, it means the record is not updated, so it's a create failure
			failureMap = failInfo.FailedCreates
		}

		if failureMap == nil {
			failureMap[errMsg] = []string{}
		}

		failureMap[errMsg] = append(failureMap[errMsg], externalId)
	}

	return &JobResults{
		FailureDetails: failInfo,
		JobId:          jobInfo.Id,
		State:          jobInfo.State,
		JobInfo:        jobInfo,
		Message:        "Some records are not processed successfully. Please refer to the 'failureDetails' for more details.",
	}, nil
}

func getColumnIndice(record []string, columnNames []string) (map[string]int, error) {
	indices := make(map[string]int)

	for _, columnName := range columnNames {
		found := false

		for j, value := range record {
			if strings.EqualFold(value, columnName) {
				indices[columnName] = j
				found = true

				break
			}
		}

		if !found {
			return nil, fmt.Errorf("%w: '%s'", ErrKeyNotFound, columnName)
		}
	}

	return indices, nil
}

func (c *Connector) getIncompleteJobResults(jobInfo *GetJobInfoResult) *JobResults {
	jobResult := &JobResults{
		JobId:   jobInfo.Id,
		State:   jobInfo.State,
		JobInfo: jobInfo,
	}

	switch {
	case jobInfo.State == JobStateInProgress || jobInfo.State == JobStateUploadComplete:
		jobResult.Message = "Job is still in progress. Please try again later."
	case jobInfo.State == JobStateAborted:
		jobResult.Message = "Job aborted. Please refer to the JobInfo for more details."
	case jobInfo.State == JobStateFailed:
		//nolint:lll
		jobResult.Message = "No records processed successfully. This is likely due the CSV being empty or issues with CSV column names."
	default:
		jobResult.Message = "Job is in an unknown state."
	}

	return jobResult
}
