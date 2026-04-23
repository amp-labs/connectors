package procore

func resolveAPIPath(objectName string, companyId string) string {
	if objectName == "projects" {
		return "rest/v1.0/companies/" + companyId + "/projects"
	}

	if objectName == "offices" {
		return "rest/v1.0/offices" + "?company_id=" + companyId
	}

	if objectName == "operations" {
		return "rest/v2.0/companies/" + companyId + "/async_operations"
	}

	if objectName == "operations" {
		return "rest/v2.0/companies/" + companyId + "/async_operations"
	}

	if objectName == "programs" {
		return "rest/v1.0/companies/" + companyId + "/programs"
	}

	return "rest/v1.0/" + objectName
}
