// nolint
package fireflies

type TypeKind string

const (
	KindScalar  TypeKind = "SCALAR"
	KindObject  TypeKind = "OBJECT"
	KindNonNull TypeKind = "NON_NULL"
	KindList    TypeKind = "LIST"
)

// TypeInfo represents the type information in the GraphQL schema.
type TypeInfo struct {
	Name   string     `json:"name"`
	Kind   TypeKind   `json:"kind"`
	OfType OfTypeInfo `json:"ofType"`
}

type OfTypeInfo struct {
	Name string   `json:"name"`
	Kind TypeKind `json:"kind"`
}

// Field represents a field in the GraphQL schema.
type Field struct {
	Name string   `json:"name"`
	Type TypeInfo `json:"type"`
}

// TypeMetadata represents the type metadata in the GraphQL schema.
type TypeMetadata struct {
	Name   string  `json:"name"`
	Fields []Field `json:"fields"`
}

// MetadataResponse represents the response structure for metadata queries.
// nolint
type MetadataResponse struct {
	Data struct {
		Type TypeMetadata `json:"__type"`
	} `json:"data"`
}

type Response struct {
	Errors any          `json:"errors"`
	Data   ResponseData `json:"data"`
}

type ResponseData struct {
	Users       []User       `json:"users,omitempty"`
	Transcripts []Transcript `json:"transcripts,omitempty"`
	Bites       []Bite       `json:"bites,omitempty"`
}

type User struct {
	UserId          string  `json:"user_id"`
	Email           string  `json:"email"`
	Name            string  `json:"name"`
	NumTranscripts  int     `json:"num_transcripts"`
	RecentMeeting   string  `json:"recent_meeting"`
	MinutesConsumed float64 `json:"minutes_consumed"`
	IsAdmin         bool    `json:"is_admin"`
	Integrations    any     `json:"integrations"`
}

type Transcript struct {
	ID        string `json:"id"`
	Sentences []struct {
		Index       int     `json:"index"`
		SpeakerName *string `json:"speaker_name"`
		SpeakerID   int     `json:"speaker_id"`
		Text        string  `json:"text"`
		RawText     string  `json:"raw_text"`
		StartTime   float64 `json:"start_time"`
		EndTime     float64 `json:"end_time"`
		AIFilters   struct {
			Task        *string `json:"task"`
			Pricing     *string `json:"pricing"`
			Metric      *string `json:"metric"`
			Question    *string `json:"question"`
			DateAndTime *string `json:"date_and_time"`
			TextCleanup *string `json:"text_cleanup"`
			Sentiment   *string `json:"sentiment"`
		} `json:"ai_filters"`
	} `json:"sentences"`
	Title    *string `json:"title"`
	Speakers *[]struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"speakers"`
	OrganizerEmail *string `json:"organizer_email"`
	MeetingInfo    struct {
		FredJoined    bool    `json:"fred_joined"`
		SilentMeeting bool    `json:"silent_meeting"`
		SummaryStatus *string `json:"summary_status"`
	} `json:"meeting_info"`
	CalendarID *string `json:"calendar_id"`
	User       struct {
		UserID          string  `json:"user_id"`
		Email           *string `json:"email"`
		Name            *string `json:"name"`
		NumTranscripts  int     `json:"num_transcripts"`
		RecentMeeting   *string `json:"recent_meeting"`
		MinutesConsumed float64 `json:"minutes_consumed"`
		IsAdmin         bool    `json:"is_admin"`
		Integrations    *string `json:"integrations"`
	} `json:"user"`
	FirefliesUsers   []string `json:"fireflies_users"`
	Participants     []string `json:"participants"`
	Date             int64    `json:"date"`
	TranscriptURL    *string  `json:"transcript_url"`
	AudioURL         *string  `json:"audio_url"`
	VideoURL         *string  `json:"video_url"`
	Duration         float64  `json:"duration"`
	MeetingAttendees *[]struct {
		DisplayName *string `json:"displayName"`
		Email       *string `json:"email"`
		PhoneNumber *string `json:"phoneNumber"`
		Name        *string `json:"name"`
		Location    *string `json:"location"`
	} `json:"meeting_attendees"`
	Summary *struct {
		Keywords           *string  `json:"keywords"`
		ActionItems        *string  `json:"action_items"`
		Outline            *string  `json:"outline"`
		ShorthandBullet    *string  `json:"shorthand_bullet"`
		Overview           *string  `json:"overview"`
		BulletGist         *string  `json:"bullet_gist"`
		Gist               *string  `json:"gist"`
		ShortSummary       *string  `json:"short_summary"`
		ShortOverview      *string  `json:"short_overview"`
		MeetingType        *string  `json:"meeting_type"`
		TopicsDiscussed    []string `json:"topics_discussed"`
		TranscriptChapters []string `json:"transcript_chapters"`
	} `json:"summary"`
	CalID        *string `json:"cal_id"`
	CalendarType *string `json:"calendar_type"`
	AppsPreview  struct {
		Outputs []struct {
			TranscriptID *string  `json:"transcript_id"`
			UserID       *string  `json:"user_id"`
			AppID        *string  `json:"app_id"`
			CreatedAt    *float64 `json:"created_at"`
			Title        *string  `json:"title"`
			Prompt       *string  `json:"prompt"`
			Response     *string  `json:"response"`
		} `json:"outputs"`
	} `json:"apps_preview"`
	MeetingLink *string `json:"meeting_link"`
}

type Bite struct {
	TranscriptId  string  `json:"transcript_id"`
	Name          string  `json:"name"`
	Id            string  `json:"id"`
	Thumbnail     string  `json:"thumbnail"`
	Preview       string  `json:"preview"`
	Status        string  `json:"status"`
	Summary       string  `json:"summary"`
	UserId        string  `json:"user_id"`
	StartTime     float64 `json:"start_time"`
	EndTime       float64 `json:"end_time"`
	SummaryStatus string  `json:"summary_status"`
	MediaType     string  `json:"media_type"`
	CreatedAt     string  `json:"created_at"`
	CreatedFrom   struct {
		Duration float64 `json:"duration"`
		Id       string  `json:"id"`
		Name     string  `json:"name"`
		Type     string  `json:"type"`
	} `json:"created_from"`
	Captions []struct {
		EndTime     string `json:"end_time"`
		Index       string `json:"index"`
		SpeakerId   string `json:"speaker_id"`
		SpeakerName string `json:"speaker_name"`
		StartTime   string `json:"start_time"`
		Text        string `json:"text"`
	} `json:"captions"`
	Sources []struct {
		Src  string `json:"src"`
		Type string `json:"type"`
	} `json:"sources"`
	Privacies []string `json:"privacies"`
	User      struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Picture   string `json:"picture"`
		Name      string `json:"name"`
		Id        string `json:"id"`
	} `json:"user"`
}
