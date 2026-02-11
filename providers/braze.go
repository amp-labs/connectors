package providers

const Braze Provider = "braze"

// BrazeEU serves customers using braze's france data-center only, the rest should use Braze.
const BrazeEU Provider = "brazeEU"

func init() {
	SetInfo(Braze, ProviderInfo{
		DisplayName: "Braze",
		AuthType:    ApiKey,
		BaseURL:     "https://rest.{{.workspace}}.braze.com",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://www.braze.com/docs/api/basics/#creating-rest-api-keys",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1740565804/media/braze.com_1740565802.png",
				LogoURL: " https://res.cloudinary.com/dycvts6vp/image/upload/v1740565904/media/braze.com_1740565903.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1740565804/media/braze.com_1740565802.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1740565849/media/braze.com_1740565848.svg",
			},
		},
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
			Write:     true,
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Instance ID",
				},
			},
		},
	})

	SetInfo(BrazeEU, ProviderInfo{
		DisplayName: "BrazeEU",
		AuthType:    ApiKey,
		BaseURL:     "https://rest.{{.workspace}}.braze.eu",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://www.braze.com/docs/api/basics/#creating-rest-api-keys",
		},
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1740565804/media/braze.com_1740565802.png",
				LogoURL: " https://res.cloudinary.com/dycvts6vp/image/upload/v1740565904/media/braze.com_1740565903.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1740565804/media/braze.com_1740565802.png",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1740565849/media/braze.com_1740565848.svg",
			},
		},
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
			Write:     true,
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "Instance ID",
				},
			},
		},
	})

}
