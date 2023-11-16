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

var (
	ErrKeyNotFound     = errors.New("key not found")
	ErrUnknownNodeType = errors.New("unknown node type")
	ErrInvalidType     = errors.New("invalid type")
	ErrInvalidJobState = errors.New("invalid job state")
)

func (c *Connector) BulkWrite(ctx context.Context, config common.BulkWriteParams) (*common.BulkWriteResult, error) {
	// cretes batch upload job, returns json with id and other info
	res, err := c.createJob(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("createJob failed: %w", err)
	}

	jobData, err := ParseNodeToMap(res)
	if err != nil {
		return nil, fmt.Errorf("parseNodeToMap failed: %w", errors.Join(err, common.ErrParseError))
	}

	state, ok := jobData["state"].(string) //nolint:varnamelen
	if !ok {
		return nil, fmt.Errorf(
			"%w. expected salesforce job state to be string in response, got %T",
			ErrInvalidType,
			jobData["state"],
		)
	}

	if strings.ToLower(state) != "open" {
		return nil, fmt.Errorf("%w: expected job state to be open, got %s", ErrInvalidJobState, state)
	}

	jobId, ok := jobData["id"] //nolint:varnamelen
	if !ok {
		return nil, fmt.Errorf("%w for key %s in %v", ErrKeyNotFound, "id", jobData)
	}

	jobIdString, ok := jobId.(string) //nolint:varnamelen
	if !ok {
		return nil, fmt.Errorf("%w. expected id to be string, gott %T", ErrInvalidType, jobId)
	}

	// upload csv and there is no response body other than status code
	_, err = c.uploadCSV(ctx, jobIdString, config)
	if err != nil {
		return nil, fmt.Errorf("uploadCSV failed: %w", err)
	}

	//
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

	state, ok := data["state"].(string) //nolint:varnamelen
	if !ok {
		return nil, fmt.Errorf(
			"%w. expected salesforce job state to be string in response, got %T",
			ErrInvalidType,
			data["state"],
		)
	}

	return &common.BulkWriteResult{
		JobId: id,
		State: state,
	}, nil
}

func joinURLPath(baseURL string, paths ...string) (string, error) {
	location, err := url.JoinPath(baseURL, paths...)
	if err != nil {
		return "", errors.Join(err, common.ErrInvalidPathJoin)
	}

	return location, nil
}

func (c *Connector) createJob(ctx context.Context, config common.BulkWriteParams) (*ajson.Node, error) {
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

func (c *Connector) uploadCSV(ctx context.Context, jobId string, config common.BulkWriteParams) ([]byte, error) {
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

	return ParseNodeToMap(res)
}

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

func (c *Connector) GetAllJobs(ctx context.Context) (*ajson.Node, error) {
	location, err := joinURLPath(c.BaseURL, "jobs/ingest")
	if err != nil {
		return nil, err
	}

	return c.get(ctx, location)
}

func (c *Connector) GetJobInfo(ctx context.Context, jobId string) (*ajson.Node, error) {
	location, err := joinURLPath(c.BaseURL, "jobs/ingest", jobId)
	if err != nil {
		return nil, err
	}

	return c.get(ctx, location)
}

func (c *Connector) FailedResults(ctx context.Context, jobId string) ([]byte, error) {
	location, err := joinURLPath(c.BaseURL, fmt.Sprintf("jobs/ingest/%s/failedResults", jobId))
	if err != nil {
		return nil, err
	}

	return c.getCSV(ctx, location)
}

func ParseNodeToMap(node *ajson.Node) (map[string]interface{}, error) {
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
