package gong

type User struct {
	ID                    string   `json:"id"`
	EmailAddress          string   `json:"emailAddress"`
	Created               string   `json:"created"`
	Active                bool     `json:"active"`
	EmailAliases          []string `json:"emailAliases"`
	TrustedEmailAddress   *string  `json:"trustedEmailAddress"`
	FirstName             string   `json:"firstName"`
	LastName              string   `json:"lastName"`
	Title                 *string  `json:"title"`
	PhoneNumber           *string  `json:"phoneNumber"`
	Extension             *string  `json:"extension"`
	PersonalMeetingUrls   []string `json:"personalMeetingUrls"`
	Settings              Settings `json:"settings"`
	ManagerID             *string  `json:"managerId"`
	MeetingConsentPageURL string   `json:"meetingConsentPageUrl"`
	SpokenLanguages       []string `json:"spokenLanguages"`
}

type Settings struct {
	WebConferencesRecorded        bool `json:"webConferencesRecorded"`
	PreventWebConferenceRecording bool `json:"preventWebConferenceRecording"`
	TelephonyCallsImported        bool `json:"telephonyCallsImported"`
	EmailsImported                bool `json:"emailsImported"`
	PreventEmailImport            bool `json:"preventEmailImport"`
	NonRecordedMeetingsImported   bool `json:"nonRecordedMeetingsImported"`
	GongConnectEnabled            bool `json:"gongConnectEnabled"`
}
