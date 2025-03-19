package drift

const (
	list   = "list"
	object = "object"
	data   = "data"
)

func responseSchema(objectName string) (string, string) {
	switch objectName {
	case "users/list", "conversations/list", "teams/org", "users/meetings/org":
		return object, data
	case "playbooks/list", "playbooks/clp":
		return list, ""
	default:
		return object, ""
	}
}
