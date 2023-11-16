package salesforce

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

func (c *Connector) BulkWrite(ctx context.Context, config common.BulkWriteParams) (*common.BulkWriteResult, error) {
	res, err := c.createJob(ctx, config)
	if err != nil {
		fmt.Println("createJob failed", err)
		return nil, err
	}

	jobData, err := NodeParser(res)

	jobId, ok := jobData["id"]
	if !ok {
		fmt.Println("Job ID not found")
		return nil, err
	}

	jobIdString, ok := jobId.(string)
	if !ok {
		fmt.Println("Job ID not string")
		return nil, fmt.Errorf("Job ID not string: %v", jobId)
	}

	_, err = c.uploadCSV(ctx, jobIdString, config)
	if err != nil {
		fmt.Println("uploadCSV failed", err)
		return nil, err
	}

	data, err := c.completeUpload(ctx, jobIdString)
	if err != nil {
		fmt.Println("completeUpload failed", err)
		return nil, err
	}

	return &common.BulkWriteResult{
		JobId: data["id"].(string),
		State: data["state"].(string),
	}, nil

}

func (c *Connector) createJob(ctx context.Context, config common.BulkWriteParams) (*ajson.Node, error) {
	location, joinErr := url.JoinPath(fmt.Sprintf("%s/jobs/ingest/", c.BaseURL))
	if joinErr != nil {
		fmt.Println(joinErr)
		return nil, joinErr
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
		fmt.Println("readfile failed", err)
		return nil, err
	}
	fmt.Println("file", len(file))

	jobUrl, err := url.JoinPath(fmt.Sprintf("%s/jobs/ingest/%s/batches", c.BaseURL, jobId))
	if err != nil {
		fmt.Println("url join failed", err)
		return nil, err
	}
	return c.put(ctx, jobUrl, file)
}

func (c *Connector) completeUpload(ctx context.Context, jobId string) (map[string]interface{}, error) {
	updateLoadCompleteBody := map[string]interface{}{
		"state": "UploadComplete",
	}

	completeUrl, err := url.JoinPath(fmt.Sprintf("%s/jobs/ingest/%s", c.BaseURL, jobId))
	if err != nil {
		fmt.Println("url", err)
		return nil, err
	}

	res3, err := c.patch(ctx, completeUrl, updateLoadCompleteBody)
	if err != nil {
		fmt.Println("Patch failed", err)
		return nil, err
	}

	return NodeParser(res3)
}

func (c *Connector) GetJobResult(ctx context.Context, jobId string) ([]byte, error) {
	resultUrl, err := url.JoinPath(fmt.Sprintf("%s/jobs/ingest/%s/successfulResults", c.BaseURL, jobId))
	if err != nil {
		fmt.Println("url", err)
		return nil, err
	}

	return c.getCSV(ctx, resultUrl)
}

func (c *Connector) GetUnprocessedResults(ctx context.Context, jobId string) ([]byte, error) {
	resultUrl, err := url.JoinPath(fmt.Sprintf("%s/jobs/ingest/%s/unprocessedrecords", c.BaseURL, jobId))
	if err != nil {
		return nil, err
	}

	return c.getCSV(ctx, resultUrl)
}

func (c *Connector) GetAllJobs(ctx context.Context) (*ajson.Node, error) {
	resultUrl, err := url.JoinPath(fmt.Sprintf("%s/jobs/ingest", c.BaseURL))
	if err != nil {
		return nil, err
	}

	fmt.Println("Alljob url:", resultUrl)

	return c.get(ctx, resultUrl)
}

func (c *Connector) GetJobInfo(ctx context.Context, jobId string) (*ajson.Node, error) {
	resultUrl, err := url.JoinPath(fmt.Sprintf("%s/jobs/ingest/%s", c.BaseURL, jobId))
	if err != nil {
		return nil, err
	}

	return c.get(ctx, resultUrl)
}

func (c *Connector) FailedResults(ctx context.Context, jobId string) ([]byte, error) {
	resultUrl, err := url.JoinPath(fmt.Sprintf("%s/jobs/ingest/%s/failedResults", c.BaseURL, jobId))
	if err != nil {
		return nil, err
	}

	return c.getCSV(ctx, resultUrl)
}

func NodeParser(node *ajson.Node) (map[string]interface{}, error) {

	parsed := map[string]interface{}{}

	for _, key := range node.Keys() {
		data, err := node.GetKey(key)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		switch data.Type() {
		case 0:
			parsed[key] = nil
		case 1:
			parsed[key] = data.MustNumeric()
		case 2:
			parsed[key] = data.MustString()
		case 3:
			parsed[key] = data.MustBool()
		case 4:
			parsed[key] = data.MustArray()
		case 5:
			parsed[key] = data.MustObject()
		default:
			return nil, fmt.Errorf("unknown type: %d", data.Type())
		}
	}

	return parsed, nil

}
