package zohocrm

// uniqueFields maps the fields to the uniquely required fields.
var uniqueFields = map[string]string{ // nolint:gochecknoglobals
	"sic_code":                 "SIC_Code",
	"skype_id":                 "Skype_ID",
	"num_sent":                 "Num_sent",
	"what_id":                  "What_Id",
	"who_id":                   "Who_Id",
	"all_day":                  "All_day",
	"zip_code":                 "ZIP_Code",
	"cti_entry":                "CTI_Entry",
	"call_duration_in_seconds": "Call_Duration_in_seconds",
	"caller_id":                "Caller_ID",
	"scheduled_in_crm":         "Scheduled_In_CRM",
}
