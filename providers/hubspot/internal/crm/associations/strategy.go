package associations

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/hubspot/internal/crm/core"
)

type Strategy struct {
	clientCRM  *common.JSONHTTPClient
	moduleInfo *providers.ModuleInfo
}

func NewStrategy(hubspotCRMClient *common.JSONHTTPClient, moduleInfo *providers.ModuleInfo) *Strategy {
	return &Strategy{
		clientCRM:  hubspotCRMClient,
		moduleInfo: moduleInfo,
	}
}

// FillAssociations fills the associations for the given object names and data.
// Note that the data is modified in place.
func (s Strategy) FillAssociations(
	ctx context.Context,
	fromObjName string,
	toAssociatedObjects []string,
	data []common.ReadResultRow,
) error {
	ids := getUniqueIDs(data)

	for _, associatedObject := range toAssociatedObjects {
		associations, err := s.fetchObjectAssociations(ctx, fromObjName, ids, associatedObject)
		if err != nil {
			return err
		}

		if len(associations) == 0 {
			continue
		}

		for i, row := range data {
			if assocs, ok := associations[row.Id]; ok {
				if data[i].Associations == nil {
					data[i].Associations = make(map[string][]common.Association)
				}

				data[i].Associations[associatedObject] = assocs
			}
		}
	}

	return nil
}

// nolint:lll
// https://developers.hubspot.com/docs/api-reference/crm-associations-v4/batch/post-crm-v4-associations-fromObjectType-toObjectType-batch-read
func (s Strategy) getAssociationsURL(fromObject, toObject string) (*urlbuilder.URL, error) {
	return urlbuilder.New(s.moduleInfo.BaseURL, core.APIVersion, "associations", fromObject, toObject, "batch/read")
}
