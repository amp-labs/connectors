package salesforce

import (
	"context"
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

var (
	ErrKeyNotFound     = errors.New("key not found")
	ErrUnknownNodeType = errors.New("unknown node type when parsing JSON")
	ErrInvalidType     = errors.New("invalid type")
	ErrInvalidJobState = errors.New("invalid job state")
)

// BulkWriteResult is what's returned from writing data via the BulkWrite call.
type BulkWriteResult struct {
	// State is the state of the bulk job process
	State string `json:"state"`
	// JobId is the ID of the bulk job process
	JobId string `json:"jobId"`
}

func (c *Connector) BulkWrite( //nolint:funlen,cyclop
	ctx context.Context,
	config BulkWriteParams,
) (*BulkWriteResult, error) {
	// cretes batch upload job, returns json with id and other info
	res, err := c.createJob(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("createJob failed: %w", err)
	}

	jobCreateRes, err := ParseAjsonNodeToMap(res)
	if err != nil {
		return nil, fmt.Errorf("parsing result of createJob failed: %w", errors.Join(err, common.ErrParseError))
	}

	state, ok := jobCreateRes["state"].(string) //nolint:varnamelen
	if !ok {
		return nil, fmt.Errorf(
			"%w: expected salesforce job state to be string in response, got %T",
			ErrInvalidType,
			jobCreateRes["state"],
		)
	}

	if strings.ToLower(state) != "open" {
		return nil, fmt.Errorf("%w: expected job state to be open, got %s", ErrInvalidJobState, state)
	}

	jobId, ok := jobCreateRes["id"] //nolint:varnamelen
	if !ok {
		return nil, fmt.Errorf("%w for key %s in job create result: %v", ErrKeyNotFound, "id", jobCreateRes)
	}

	jobIdString, ok := jobId.(string) //nolint:varnamelen
	if !ok {
		return nil, fmt.Errorf("%w. expected id to be string, got %T", ErrInvalidType, jobId)
	}

	// upload csv and there is no response body other than status code
	_, err = c.uploadCSV(ctx, jobIdString, config)
	if err != nil {
		return nil, fmt.Errorf("uploadCSV failed: %w", err)
	}

	data, err := c.completeUpload(ctx, jobIdString)
	if err != nil {
		return nil, fmt.Errorf("completeUpload failed: %w", err)
	}

	id, ok := data["id"].(string) //nolint:varnamelen
	if !ok {
		return nil, fmt.Errorf(
			"%w. expected salesforce job id to be string in response, got %T",
			ErrInvalidType,
			data["id"],
		)
	}

	completeState, ok := data["state"].(string) //nolint:varnamelen
	if !ok {
		return nil, fmt.Errorf(
			"%w. expected salesforce job state to be string in response, got %T",
			ErrInvalidType,
			data["state"],
		)
	}

	return &BulkWriteResult{
		JobId: id,
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

func (c *Connector) completeUpload(ctx context.Context, jobId string) (map[string]interface{}, error) {
	updateLoadCompleteBody := map[string]interface{}{
		"state": "UploadComplete",
	}

	location, err := joinURLPath(c.BaseURL, fmt.Sprintf("jobs/ingest/%s", jobId))
	if err != nil {
		return nil, err
	}

	res, err := c.patch(ctx, location, updateLoadCompleteBody)
	if err != nil {
		return nil, fmt.Errorf("patch failed: %w", errors.Join(err, common.ErrRequestFailed))
	}

	return ParseAjsonNodeToMap(res)
}

func ParseAjsonNodeToMap(node *ajson.Node) (map[string]interface{}, error) {
	parsed := map[string]interface{}{}

	for _, key := range node.Keys() {
		data, err := node.GetKey(key)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrKeyNotFound, key)
		}

		switch data.Type() {
		case ajson.Null:
			parsed[key] = data.MustNull()
		case ajson.Numeric:
			parsed[key] = data.MustNumeric()
		case ajson.String:
			parsed[key] = data.MustString()
		case ajson.Bool:
			parsed[key] = data.MustBool()
		case ajson.Array:
			parsed[key] = data.MustArray()
		case ajson.Object:
			parsed[key] = data.MustObject()
		default:
			return nil, fmt.Errorf("%w: %d", ErrUnknownNodeType, data.Type())
		}
	}

	return parsed, nil
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

func (c *Connector) GetJobInfo(ctx context.Context, jobId string) (*GetJobInfoResult, error) {
	location, err := joinURLPath(c.BaseURL, "jobs/ingest", jobId)
	if err != nil {
		return nil, err
	}

	node, err := c.get(ctx, location)
	if err != nil {
		return nil, fmt.Errorf("getGetInfo failed: %w", errors.Join(err, common.ErrRequestFailed))
	}

	dataMap, err := ParseAjsonNodeToMap(node)
	if err != nil {
		return nil, fmt.Errorf("parsing result of getGetInfo failed: %w", errors.Join(err, common.ErrParseError))
	}

	// Below is omitting type assertion because we are already checking for type in ParseAjsonNodeToMap
	return &GetJobInfoResult{
		Id:                      dataMap["id"].(string),                       //nolint:forcetypeassert
		Object:                  dataMap["object"].(string),                   //nolint:forcetypeassert
		CreatedById:             dataMap["createdById"].(string),              //nolint:forcetypeassert
		ExternalIdFieldName:     dataMap["externalIdFieldName"].(string),      //nolint:forcetypeassert
		State:                   dataMap["state"].(string),                    //nolint:forcetypeassert
		Operation:               dataMap["operation"].(string),                //nolint:forcetypeassert
		ColumnDelimiter:         dataMap["columnDelimiter"].(string),          //nolint:forcetypeassert
		LineEnding:              dataMap["lineEnding"].(string),               //nolint:forcetypeassert
		NumberRecordsFailed:     dataMap["numberRecordsFailed"].(float64),     //nolint:forcetypeassert
		NumberRecordsProcessed:  dataMap["numberRecordsProcessed"].(float64),  //nolint:forcetypeassert
		ApexProcessingTime:      dataMap["apexProcessingTime"].(float64),      //nolint:forcetypeassert
		ApiActiveProcessingTime: dataMap["apiActiveProcessingTime"].(float64), //nolint:forcetypeassert
		ApiVersion:              dataMap["apiVersion"].(float64),              //nolint:forcetypeassert
		ConcurrencyMode:         dataMap["concurrencyMode"].(string),          //nolint:forcetypeassert
		ContentType:             dataMap["contentType"].(string),              //nolint:forcetypeassert
		CreatedDate:             dataMap["createdDate"].(string),              //nolint:forcetypeassert
		JobType:                 dataMap["jobType"].(string),                  //nolint:forcetypeassert
		Retries:                 dataMap["retries"].(float64),                 //nolint:forcetypeassert
		SystemModstamp:          dataMap["systemModstamp"].(string),           //nolint:forcetypeassert
		TotalProcessingTime:     dataMap["totalProcessingTime"].(float64),     //nolint:forcetypeassert
	}, nil
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
