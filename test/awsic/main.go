package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
)

func main() {
	main2()

	//awsic.Connector{}.Print()
}

func main2() {
	// Load config (from env or shared config)
	awsRegion := "us-east-2"
	awsService := "sso"
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
	}

	// Your request
	payload := []byte(`{"MaxResults":10}`) // Example payload
	req, err := http.NewRequest("POST", "https://"+awsService+"."+awsRegion+".amazonaws.com/", bytes.NewReader(payload))
	if err != nil {
		panic(err)
	}

	// Required headers
	req.Header.Set("Content-Type", "application/x-amz-json-1.1")
	req.Header.Set("X-Amz-Target", "SWBExternalService.ListInstances")

	// Sign the request
	signer := v4.NewSigner()
	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		panic(err)
	}

	sum := sha256.Sum256(payload)
	payloadHash := hex.EncodeToString(sum[:])

	err = signer.SignHTTP(ctx, creds, req, payloadHash, awsService, awsRegion, time.Now())
	if err != nil {
		panic(err)
	}

	// Send it
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	fmt.Println("Status:", resp.Status)
	fmt.Println("Response:", string(respBody))
}
