package zendeskchat

func responseFields(objectName string) string {
	switch objectName {
	case "chats", "incremental/chats":
		return "chats"
	case "incremental/agent_events":
		return "agent_events"
	case "incremental/agent_timeline":
		return "agent_timeline"
	case "incremental/conversions":
		return "conversions"
	case "incremental/department_events":
		return "department_events"
	default:
		return ""
	}
}
