package marketo

import "errors"

var ErrEmptyResultResponse = errors.New("writing reponded with an empty result")

var (
	metadataPageSize     string = "1"         //nolint:gochecknoglobals
	assetsQueryParameter string = "maxReturn" //nolint:gochecknoglobals
	leadsQueryParameter  string = "batchSize" //nolint:gochecknoglobals
)
