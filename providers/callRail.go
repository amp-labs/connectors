package providers

const CallRail Provider = "callRail"

func init() {
	// CallRail Configuration
	SetInfo(CallRail, ProviderInfo{

        AuthType: ApiKey,
        BaseURL: "https://api.callrail.com",
        // For 6sense, the header needs to be 'Authorization: Token {your_api_key}'
        ApiKeyOpts: &ApiKeyOpts{
			AttachmentType: Header,
			Header: &ApiKeyOptsHeader{
				Name: "Authorization",
				ValuePrefix: "Token token=",
			},
            DocsURL: "https://apidocs.callrail.com/#getting-started",
		},
        // For another provider, ValuePrefix may not be needed
        // For example, if the expected header is 'X-Api-Key: {your_api_key}'
        /*
        ApiKeyOpts: &ApiKeyOpts{
			HeaderName: "X-Api-Key",
            DocsURL: "https://api.6sense.com/docs/#get-your-api-token",
		}, */
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
   
	})
}
