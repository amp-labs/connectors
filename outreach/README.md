# Outreach connector


## Read
Read is used to list all records of a given type. For example, if you want to list all users, you would use the `Read` method with the `users` object.

### Example Usage

```

// Create the outreach connector instance 
// This assumes you called the instance client

// Call Read to list records in a users object
res, err := client.Read(context.TODO(),common.ReadParams{
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

// Create an outreach connector instance (any method you prefer)
// This Assumes you named the instance client

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

result, err := client.Write(ctx, config)
if err != nil {
	log.Fatal(err)
}
```


