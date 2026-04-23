package core

const (
	APIVersion                 = "60.0"
	versionPrefix              = "v"
	version                    = versionPrefix + APIVersion
	RestAPISuffix              = "/services/data/" + version
	URISobjects                = RestAPISuffix + "/sobjects"
	URIToolingEventRelayConfig = RestAPISuffix + "/tooling/sobjects/EventRelayConfig"
)
