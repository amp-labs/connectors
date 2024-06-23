package providers

const Mailgun Provider = "mailgun"

func init() {
	SetInfo(Mailgun, ProviderInfo{
		AuthType: Basic,
		BaseURL:  "https://api.mailgun.net/",
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
		PostAuthInfoNeeded: false,
	})
}
