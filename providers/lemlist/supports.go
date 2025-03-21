package lemlist

func responseSchema(objectName string) (string, string) {
	// team --> an object
	// api/team/senders --> an array of objects
	// api/team/credits -->object
	// api/campaigns --> object campaigns array
	// api/activities --> an array of objects
	// api/unsubscirbes --> an array of objects
	// api/hooks --> an array of objects
	// api/database/filters -->array of objects
	// api/schema/people --> an object
	// api/schema/companies --> an object
	switch objectName {
	case "campaigns", "schedules":
		return object, objectName
	case "team/senders", "activities", "unsubscirbes", "hooks", "database/filters":
		return list, ""
	default:
		return object, ""
	}
}
