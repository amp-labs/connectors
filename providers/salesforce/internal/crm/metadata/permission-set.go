package metadata

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/salesforce/internal/crm/core"
)

const (
	PermissionSetType               = "PermissionSet"
	DefaultPermissionSetName        = "IntegrationCustomFieldVisibility"
	DefaultPermissionSetLabel       = "Custom Field Visibility for Integration"
	DefaultPermissionSetDescription = "Permission set for integration to be able to read value of custom fields." // nolint:lll
)

var (
	ErrPermissionSetNotFound         = errors.New("permission set for optional custom fields is not found")
	ErrPermissionSetAssignmentFailed = errors.New("creation of PermissionSetAssignment was unsuccessful")
)

func (a *Adapter) fetchPermissionSetID(ctx context.Context) (string, error) {
	url, err := a.getQueryURL()
	if err != nil {
		return "", err
	}

	soql := (&core.SOQLBuilder{}).
		SelectFields([]string{"Id", "Name"}).
		From(PermissionSetType).
		Where(fmt.Sprintf("Name='%v'", DefaultPermissionSetName)).
		String()

	url.WithQueryParam("q", soql)

	resp, err := a.ClientCRM.Get(ctx, url.String())
	if err != nil {
		return "", err
	}

	response, err := common.UnmarshalJSON[readPermissionSetResponse](resp)
	if err != nil {
		return "", err
	}

	for _, record := range response.Records {
		if record.Name == DefaultPermissionSetName {
			return record.Id, nil
		}
	}

	return "", ErrPermissionSetNotFound
}

type readPermissionSetResponse struct {
	TotalSize int  `json:"totalSize"`
	Done      bool `json:"done"`
	Records   []struct {
		Attributes any    `json:"attributes"`
		Id         string `json:"Id"`
		Name       string `json:"Name"`
	} `json:"records"`
}

func (a *Adapter) fetchUserID(ctx context.Context) (string, error) {
	url, err := a.getUserInfoURL()
	if err != nil {
		return "", err
	}

	resp, err := a.ClientCRM.Get(ctx, url.String())
	if err != nil {
		return "", err
	}

	response, err := common.UnmarshalJSON[readUserInfoResponse](resp)
	if err != nil {
		return "", err
	}

	return response.UserID, nil
}

type readUserInfoResponse struct {
	UserID string `json:"user_id"`
}

type createPermissionSetAssignmentPayload struct {
	UserID          string `json:"AssigneeId"`
	PermissionSetID string `json:"PermissionSetId"`
}

func (a *Adapter) assignPermissionSetToUser(ctx context.Context, userID, permissionSetID string) error {
	url, err := a.getSobjectsURL("PermissionSetAssignment")
	if err != nil {
		return err
	}

	payload := createPermissionSetAssignmentPayload{
		UserID:          userID,
		PermissionSetID: permissionSetID,
	}

	resp, err := a.ClientCRM.Post(ctx, url.String(), payload)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate PermissionSetAssignment") {
			// Success, because Assignment is already present.
			return nil
		}

		return err
	}

	response, err := common.UnmarshalJSON[createPermissionSetAssignmentResponse](resp)
	if err != nil {
		return err
	}

	if !response.Success {
		return ErrPermissionSetAssignmentFailed
	}

	return nil
}

type createPermissionSetAssignmentResponse struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Errors  []any  `json:"errors"`
}
