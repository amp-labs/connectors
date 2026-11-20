package core

const (
	// ObjectNameMessages represents all messages in the mailbox of the currently authorized user.
	// See: https://learn.microsoft.com/en-us/graph/api/user-list-messages?view=graph-rest-1.0
	ObjectNameMessages         = "me/messages"
	VirtualObjectDrafts        = "AMPERSAND-drafts"       // based on "/me/messages"
	VirtualObjectSentMessages  = "AMPERSAND-sentMessages" // based on "/me/messages"
	VirtualObjectInboxMessages = "AMPERSAND-messages"     // based on "/me/messages"
)
