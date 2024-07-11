# Outreach connector


## Read
Read is used to list all records of a given type. For example, if you want to list all users, you would use the `Read` method with the `users` object.

### Example Usage

```

// Create the outreach connector instance 
conn, err := outreach.NewConnector(
		outreach.WithClient(ctx, http.DefaultClient, cfg, tok),
)

// Call Read to list records in a users object
res, err := conn.Read(context.TODO(),common.ReadParams{
		ObjectName: "users",
})
if err != nil {
	log.Fatal(err)
}

```

## Write
Write is used to create/update objects in outreach connector. For an instance creating an emailAddress object, you use the Write method and `emailAddress` object   

### Example usage
```

type Attribute struct {
	Email     string `json:"email"`
	EmailType string `json:"emailType"`
	Order     int    `json:"order"`
	Status    string `json:"status"`
}

type EmailAddress struct {
	Attributes Attribute `json:"attributes"`
	Type       string    `json:"type"`
}

// Create an outreach connector instance 
conn, err := outreach.NewConnector(
		outreach.WithClient(ctx, http.DefaultClient, cfg, tok),
)

attribute := Attribute{
		Email:     gofakeit.Email(),
		EmailType: "email",
		Order:     0,
		Status:    "null",
	}

	config := common.WriteParams{
		ObjectName: "emailAddresses",
		RecordData: EmailAddress{
			Attributes: attribute,
			Type:       "emailAddress",
		},
	}

result, err := conn.Write(ctx, config)
if err != nil {
	log.Fatal(err)
}
```

## Supported Objects 
Below is an exhaustive list of the supported Objects in the Outreach deep connector with their endpoint resources(ObjectName).

Outreach API version : v2

| Object | Resource | Method |
| :-------- | :------- | :-------- |
| Account | accounts | Read, Write |
| Audit | audits | Read |
| Audit Logs | auditLogs | Read |
| Call | calls | Read, Write |
| Call Disposition | callDispositions| Read, Write |
| Call Purpose | callPurpose | Read, Write |
| Compliance Request | complianceRequest| Read, Write |
| Content Category | contentCategories | Read, Write |
| Content Category Membership | contentCategoryMemberships | Read, Write |
| Content Category Ownership | contentCategoryOwnerships| Read, Write |
| Custom Duty | customDuties | write |
| Duty | duties | Read |
| Email Address | emailAddress | Read, Write |
| Event | events/:id(by  id only) | Read |
| Favorite | favorites | Read, Write |
| Mail Alias | mailAliases | Read |
| Mailbox | mailboxes | Read, Write |
| Mailing | mailings | Read, Write |
| Opportunity | opportunities | Read, Write |
| Opportunity Prospect Role |  opportunityProspectRoles | Read, Write |
| Opportunity Stage | opportunityStages | Read, Write |
| Org Setting  | orgSettings/:id (by id only) | Read, Write(Patch by id ony) |
| Persona | personas | Read, Write |
| Phone Number | phoneNumbers | Read, Write |
| Profile | profiles | Read, Write |
| Prospect | prospects | Read, Write |
| Recipient | recipients | Read, Write | 
| Role | roles | Read, Write |
| Ruleset | rulesets | Read, Write |
| Sequence | sequences | Read, Write |
| Sequence State | sequenceStates | Read, Write |
| Sequence Step | sequenceSteps | Read, Write |
| Sequence Template | sequenceTemplates | Read, Write |
| Snippet | snippets | Read, Write |
| Stage | stages | Read, Write |
| Task | tasks | Read, Write |
| Task Disposition | taskDispositions | Read, Write |
| Task Priority | taskPriorities | Read |
| Task Purpose | taskPurposes | Read, Write |
| Team | teams | Read, Write |
| Template | templates | Read, Write |
| templates | users | Read, Write |
| Webhook | webhooks | Read, Write |