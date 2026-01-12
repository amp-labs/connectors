package closecrm

func addTrailingSlashIfNeeded(urlString string) string {
	if len(urlString) > 0 && urlString[len(urlString)-1] != '/' {
		urlString += "/"
	}

	return urlString
}
