package compoundfields

// compoundFieldFlattenedByObject is the source of truth for standard Salesforce
// compound field sub-components and their flattened column names (from the
// Salesforce API Object Reference). Shape: object -> compound -> sub-field ->
// flattened column name. All map keys and string values are lowercased for
// case-insensitive handling end-to-end.
//
//nolint:gochecknoglobals // static reference data
var compoundFieldFlattenedByObject = map[string]map[string]map[string]string{
	// Account:
	// https://developer.salesforce.com/docs/atlas.en-us.object_reference.meta/object_reference/sforce_api_objects_account.htm
	"account": {
		"billingaddress": {
			"street":           "billingstreet",
			"city":             "billingcity",
			"state":            "billingstate",
			"statecode":        "billingstatecode",
			"postalcode":       "billingpostalcode",
			"country":          "billingcountry",
			"countrycode":      "billingcountrycode",
			"latitude":         "billinglatitude",
			"longitude":        "billinglongitude",
			"geocodeaccuracy":  "billinggeocodeaccuracy",
		},
		"shippingaddress": {
			"street":           "shippingstreet",
			"city":             "shippingcity",
			"state":            "shippingstate",
			"statecode":        "shippingstatecode",
			"postalcode":       "shippingpostalcode",
			"country":          "shippingcountry",
			"countrycode":      "shippingcountrycode",
			"latitude":         "shippinglatitude",
			"longitude":        "shippinglongitude",
			"geocodeaccuracy":  "shippinggeocodeaccuracy",
		},
	},
	// Contact:
	// https://developer.salesforce.com/docs/atlas.en-us.object_reference.meta/object_reference/sforce_api_objects_contact.htm
	"contact": {
		"mailingaddress": {
			"street":           "mailingstreet",
			"city":             "mailingcity",
			"state":            "mailingstate",
			"statecode":        "mailingstatecode",
			"postalcode":       "mailingpostalcode",
			"country":          "mailingcountry",
			"countrycode":      "mailingcountrycode",
			"latitude":         "mailinglatitude",
			"longitude":        "mailinglongitude",
			"geocodeaccuracy":  "mailinggeocodeaccuracy",
		},
		"otheraddress": {
			"street":           "otherstreet",
			"city":             "othercity",
			"state":            "otherstate",
			"statecode":        "otherstatecode",
			"postalcode":       "otherpostalcode",
			"country":          "othercountry",
			"countrycode":      "othercountrycode",
			"latitude":         "otherlatitude",
			"longitude":        "otherlongitude",
			"geocodeaccuracy":  "othergeocodeaccuracy",
		},
		// Name compound: flattened column names match the sub-field API names.
		"name": {
			"firstname":   "firstname",
			"lastname":    "lastname",
			"middlename":  "middlename",
			"salutation":  "salutation",
			"suffix":      "suffix",
		},
	},
	// Lead:
	// https://developer.salesforce.com/docs/atlas.en-us.object_reference.meta/object_reference/sforce_api_objects_lead.htm
	"lead": {
		"address": {
			"street":           "street",
			"city":             "city",
			"state":            "state",
			"statecode":        "statecode",
			"postalcode":       "postalcode",
			"country":          "country",
			"countrycode":      "countrycode",
			"latitude":         "latitude",
			"longitude":        "longitude",
			"geocodeaccuracy":  "geocodeaccuracy",
		},
		"name": {
			"firstname":  "firstname",
			"lastname":   "lastname",
			"middlename": "middlename",
			"salutation": "salutation",
			"suffix":     "suffix",
		},
	},
}
