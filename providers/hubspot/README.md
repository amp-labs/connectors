# Hubspot connector

This connector has two ways of reading data from Hubspot:
1. Read
2. Search

## Read
Read is used to list all records of a given type. For example, if you want to list all contacts, you would use the `Read` method with the `contacts` object.

### Example
To list all contacts with the fields `email`, `firstname`, and `lastname`:
```
client.Read(context.Background(), &common.ReadParams{
    ObjectName: "contacts",
    Fields:    []string{"email", "firstname", "lastname"},
})
```

If the 'Since' field in the `ReadParams` is set, the connector will use the search endpoint to filter records using the `lastmodifieddate` property. However, this result set is limited to a maximum of 10,000 records. This limit is applicable to any call made via the `Search` endpoint. Read more @ https://developers.hubspot.com/docs/api/crm/search#limitations.

## Search
Search is used to find records of a given type that match a given query. For example, if you want to find all contacts with the name "John", you would use the `Search` method with the `contacts` object.

### Example

To find all contacts with the firstname "Brian" and a last modified date after "2023-10-26T17:56:14.834Z", sorted by the `hs_object_id` field in descending order:
```
result, err := client.Search(context.Background(), hubspot.SearchParams{
    ObjectName: "contacts",
    FilterGroups: []hubspot.FilterGroup{
        {
            Filters: []hubspot.Filter{
                {
                    FieldName: "firstname",
                    Operator:  hubspot.FilterOperatorTypeEQ,
                    Value:     "Brian",
                },
                {
                    FieldName: "lastname",
                    Operator:  hubspot.FilterOperatorTypeEQ,
                    Value:     "Halligan (Sample Contact)",
                },
            },
        },
        {
            Filters: []hubspot.Filter{
                {
                    FieldName: "firstname",
                    Operator:  hubspot.FilterOperatorTypeEQ,
                    Value:     "Maria",
                },
            },
        },
    },
    SortBy: []hubspot.SortBy{
        {
            PropertyName: "hs_object_id",
            Direction:    hubspot.SortDirectionDesc,
        },
    },
})
```
