package providers

import "github.com/amp-labs/connectors/common"

const AWS Provider = "aws"

const ModuleAWSIdentityCenter common.ModuleID = "aws-identity-center"

func init() {
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
			Proxy:     false,
			Read:      false,
			Subscribe: false,
			Write:     false,
		},
		Modules: &Modules{
			ModuleAWSIdentityCenter: {
				// TODO serviceDomain changes based on the request. This is not global to the connector.
				BaseURL:     "https://{{.serviceDomain}}.{{.region}}.amazonaws.com",
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
					Name: "region",
				},
				{
					Name: "identityStoreId",
				},
				{
					Name: "instanceArn",
				},
				// IMPORTANT: The 'serviceDomain' variable is figured out in the connector,
				// at runtime. It is not part of the metadata.
			},
		},
	})
}
