package providers

import "github.com/amp-labs/connectors/common"

const AWS Provider = "aws"

const ModuleAWSIdentityCenter common.ModuleID = "aws-identity-center"

func init() { //nolint:funlen
	SetInfo(AWS, ProviderInfo{
		DisplayName: "Amazon Web Services",
		AuthType:    Basic,
		BaseURL:     "https", // TODO is it ok that this value will be empty? Validation enforces some value.
		Support: Support{
			BulkWrite: BulkWriteSupport{
				Insert: false,
				Update: false,
				Upsert: false,
				Delete: false,
			},
			Proxy:     true,
			Read:      true,
			Subscribe: false,
			Write:     false,
		},
		DefaultModule: ModuleAWSIdentityCenter,
		Modules: &Modules{
			ModuleAWSIdentityCenter: {
				// TODO: The service domain changes based on the request. This is not global to the connector.
				// This is also not a metadata field. It's decided based on the request by the connector's logic.
				// We are special casing this for now, but we'll revisit this in the future to decide how to model this case.
				// Using the <<>> syntax to indicate that this is a special case. Find '<<SERVICE_DOMAIN>>' in the connector
				// to understand how this is used.
				BaseURL:     "https://<<SERVICE_DOMAIN>>.{{.region}}.amazonaws.com",
				DisplayName: "AWS Identity Center",
				Support: Support{
					Read:      false,
					Subscribe: false,
					Write:     false,
				},
			},
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1746046733/media/aws_1746046732.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1746046794/media/aws_1746046793.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1746046733/media/aws_1746046732.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1746046777/media/aws_1746046777.svg",
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "region",
					DisplayName: "Region",
					ModuleDependencies: &ModuleDependencies{
						ModuleAWSIdentityCenter: ModuleDependency{},
					},
				},
				{
					Name:        "identityStoreId",
					DisplayName: "Identity Store ID",
					ModuleDependencies: &ModuleDependencies{
						ModuleAWSIdentityCenter: ModuleDependency{},
					},
				},
				{
					Name:        "instanceArn",
					DisplayName: "Instance ARN",
					ModuleDependencies: &ModuleDependencies{
						ModuleAWSIdentityCenter: ModuleDependency{},
					},
				},
				// IMPORTANT: The 'serviceDomain' variable is figured out in the connector,
				// at runtime. It is not part of the metadata.
			},
		},
	})
}
