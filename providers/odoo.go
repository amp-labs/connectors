package providers

const Odoo Provider = "odoo"

func init() {
	SetInfo(Odoo, ProviderInfo{
		DisplayName: "Odoo",
		ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name:        "Authorization",
				ValuePrefix: "Bearer ",
			},
			DocsURL: "https://www.odoo.com/documentation/19.0/developer/reference/external_api.html#configuration",
		},
		// e.g. yourcompany.odoo.com,odoo.yourdomain.com
		BaseURL:  "https://{{.workspace}}",
		AuthType: ApiKey,
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
		Media: &Media{
			DarkMode: &MediaTypeDarkMode{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1776183638/media/odoo.com_1776183637.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1776183246/media/odoo.com_1776183245.svg",
			},
			Regular: &MediaTypeRegular{
				IconURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1776183638/media/odoo.com_1776183637.jpg",
				LogoURL: "https://res.cloudinary.com/dycvts6vp/image/upload/v1776183223/media/odoo.com_1776183222.svg",
			},
		},
		Metadata: &ProviderMetadata{
			Input: []MetadataItemInput{
				{
					Name:        "workspace",
					DisplayName: "API Domain",
					Prompt:      "Enter your Odoo domain (e.g. yourcompany.odoo.com or odoo.yourdomain.com)",
				},
			},
		},
	})
}
