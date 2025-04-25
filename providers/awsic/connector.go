package awsic

// TODO
// sso
// ssoadmin
// ssooidc

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
)

type Connector struct {
}

func (c Connector) Print() {
	ctx := context.TODO()
	awsRegion := "us-east-2" // "us-west-2"

	// Using the SDK's default configuration, load additional config
	// and credentials values from the environment variables, shared
	// credentials, and shared configuration files
	accessKeyID := "TODO"
	accessKeySecret := "TODO"
	sessionToken := "" // permanent credentials
	//identityStoreID := "d-9a670e6550"

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(awsRegion),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKeyID, accessKeySecret, sessionToken),
		),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	//identityStoreClient := identitystore.NewFromConfig(cfg)

	//users, err := identityStoreClient.ListUsers(ctx, &identitystore.ListUsersInput{
	//	IdentityStoreId: goutils.Pointer(identityStoreID),
	//	Filters:         nil,
	//	MaxResults:      nil,
	//	NextToken:       nil,
	//})
	//
	//fmt.Print(users)

	ssoAdminClient := ssoadmin.NewFromConfig(cfg)

	instances, err := ssoAdminClient.ListInstances(ctx, &ssoadmin.ListInstancesInput{
		MaxResults: nil,
		NextToken:  nil,
	})

	fmt.Print(instances)
}
