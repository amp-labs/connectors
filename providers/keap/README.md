# Custom Fields

The `Read` operation will return custom fields alongside other native properties.
Keap API indexes custom fields using numbered identifiers without including human-readable names.
This issue is addressed in the implementation.


## Example
```json
{
  "id": 22,
  "given_name": "Erica",
  "jobtitle": "Product Owner" // custom field
} 
```


## Explanation
[List Contacts](https://developer.keap.com/docs/rest/#tag/Contact/operation/listContactsUsingGET)
When requesting the contacts resource, the response includes custom fields as follows:
```json
{
  "contacts": [
    {
      "id": 22,
      "given_name": "Erica",
      "custom_fields": [
        {
          "id": 6, // Obscure field name.
          "content": "Product Owner"
        },
      ]
    }
  ]
}
```
[Contacts Model](https://developer.keap.com/docs/rest/#tag/Contact/operation/retrieveContactModelUsingGET)
To determine the meaning of `"id": 6`, you can query the contacts model, which provides additional details:
```json
{
  "custom_fields": [
    {
      "id": 6,
      "label": "title",
      "options": [],
      "record_type": "CONTACT",
      "field_type": "Text",
      "field_name": "jobtitle", // Prefered field name.
      "default_value": null
    },
  ],
  "optional_properties": []
}
```
Combining both results, the connector's `Read` will return the following:
```json
{
  "id": 22,
  "given_name": "Erica",
  "jobtitle": "Product Owner"
} 
```
