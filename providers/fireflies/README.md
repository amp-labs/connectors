# Fireflies connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

--------------------------------------------------------------------------------
| Object             | Resource                              | Method          |
| -------------------| --------------------------------------| ----------------|
| users              | users                                 | read            |
| transcripts        | transcripts, deleteTranscript         | read, write     |
| bites              | bites, createBite                     | read, write     |
| userGroups         | user_groups                           | read            |
| activeMeetings     | active_meetings                       | read            |
| analytics          | analytics                             | read            |
| liveMeetings       | addToLiveMeeting                      | write           |
| meetingTitle       | updateMeetingTitle                    | write           |
| userRole           | setUserRole                           | write           |
| audio              | uploadAudio                           | write           |
| meetingPrivacy     | updateMeetingPrivacy                  | write           |
--------------------------------------------------------------------------------

Below objects supports incremental read
- transcripts
- analytics