package mixmax

func responseField(objectName string) string {
	switch objectName {
	case "appointmentlinks/me", "userpreferences/me", "users/me":
		return ""
	default:
		return "results"
	}
}
