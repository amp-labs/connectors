package breezy

import (
	"strings"

	"github.com/amp-labs/connectors/common/urlbuilder"
)

const (
	restAPIVersion       = "v3"
	companyIDPlaceholder = "{company_id}"
	positionIDPlaceholder = "{position_id}"

	// Collection vs resource paths differ: POST …/positions, PUT …/position/{id}.
	positionsCollectionPath = "/company/" + companyIDPlaceholder + "/positions"
	positionResourcePath    = "/company/" + companyIDPlaceholder + "/position/" + positionIDPlaceholder
	positionStatePath       = positionResourcePath + "/state"
)

func buildVersionedPathURL(baseURL string, path string) (*urlbuilder.URL, error) {
	return urlbuilder.New(baseURL, restAPIVersion, strings.TrimSpace(path))
}

func resolveObjectPath(path string, companyID string) string {
	return strings.ReplaceAll(path, companyIDPlaceholder, companyID)
}

func resolvePositionPath(path, companyID, positionID string) string {
	path = resolveObjectPath(path, companyID)

	return strings.ReplaceAll(path, positionIDPlaceholder, positionID)
}

func buildCompanyPositionsURL(baseURL, companyID string) (*urlbuilder.URL, error) {
	return buildVersionedPathURL(baseURL, resolveObjectPath(positionsCollectionPath, companyID))
}

func buildCompanyPositionURL(baseURL, companyID, positionID string) (*urlbuilder.URL, error) {
	return buildVersionedPathURL(baseURL, resolvePositionPath(positionResourcePath, companyID, positionID))
}

func buildCompanyPositionStateURL(baseURL, companyID, positionID string) (*urlbuilder.URL, error) {
	return buildVersionedPathURL(baseURL, resolvePositionPath(positionStatePath, companyID, positionID))
}
