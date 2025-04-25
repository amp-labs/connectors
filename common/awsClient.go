package common

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

type ctxKey string

const AWSServiceContextKey ctxKey = "AWSService"

// ErrRequestAWSMissingService is returned when deep connector implementation doesn't attach
// AWS service name into the context, therefore the request cannot be constructed for sending.
var ErrRequestAWSMissingService = errors.New("AWS request is missing Service name, supplied via context")

type AWSClient struct {
	client *http.Client
	cfg    aws.Config
	region string
}

func NewAWSClient(ctx context.Context, client *http.Client, accessKeyID, accessKeySecret, region string) (AuthenticatedHTTPClient, error) {
	sessionToken := "" // empty value signifies permanent credentials

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKeyID, accessKeySecret, sessionToken),
		),
	)
	if err != nil {
		return nil, err
	}

	return AWSClient{
		client: client,
		cfg:    cfg,
		region: region,
	}, nil
}

func (c AWSClient) Do(req *http.Request) (*http.Response, error) {
	// Sign the request
	ctx := req.Context()

	awsService, ok := ctx.Value(AWSServiceContextKey).(string)
	if !ok || len(awsService) == 0 {
		return nil, ErrRequestAWSMissingService
	}

	payload, err := extractPayload(req)
	if err != nil {
		return nil, err
	}

	sum := sha256.Sum256(payload)
	payloadHash := hex.EncodeToString(sum[:])

	creds, err := c.cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, err
	}

	signer := v4.NewSigner()
	err = signer.SignHTTP(ctx, creds, req, payloadHash, awsService, c.region, time.Now())
	if err != nil {
		return nil, err
	}

	return c.client.Do(req)
}

func (c AWSClient) CloseIdleConnections() {
	c.client.CloseIdleConnections()
}

func extractPayload(req *http.Request) ([]byte, error) {
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	req.Body.Close()
	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	return bodyBytes, nil
}
