package mixmax

func responseField(objectName string) string {
	switch objectName {
	case "appointmentlinks/me", "userpreferences/me", "users/me":
		return "" // indicates we're reading data fields from the root level.
	default:
		return "results"
	}
}
