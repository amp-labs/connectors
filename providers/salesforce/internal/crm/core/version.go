package core

const (
	apiVersion                 = "60.0"
	versionPrefix              = "v"
	version                    = versionPrefix + apiVersion
	RestAPISuffix              = "/services/data/" + version
	URISobjects                = RestAPISuffix + "/sobjects"
	URIToolingEventRelayConfig = RestAPISuffix + "/tooling/sobjects/EventRelayConfig"
)
