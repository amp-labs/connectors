package aws

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/aws/internal/identitystore"
	"github.com/amp-labs/connectors/providers/aws/internal/ssoadmin"
)

func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*connectors.ListObjectMetadataResult, error) {
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	result := common.NewListObjectMetadataResult()

	for _, name := range objectNames {
		if data, ok := findObjectMetadata(name); ok {
			result.Result[name] = *data
		} else {
			result.Errors[name] = common.ErrObjectNotSupported
		}
	}

	return result, nil
}

// Try locating metadata for an object across AWS services.
func findObjectMetadata(objectName string) (*common.ObjectMetadata, bool) {
	data, err := identitystore.Schemas.SelectOne(providers.ModuleAWSIdentityCenter, objectName)
	if err == nil {
		return data, true
	}

	data, err = ssoadmin.Schemas.SelectOne(providers.ModuleAWSIdentityCenter, objectName)
	if err == nil {
		return data, true
	}

	return nil, false
}
