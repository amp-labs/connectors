package clari

func responseField(objectName string) string {
	switch objectName {
	case "audit/events":
		return "items"
	case "export/jobs":
		return "jobs"
	default:
		return ""
	}
}
