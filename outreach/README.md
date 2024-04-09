# Outreach connector


## Read
Read is used to list all records of a given type. For example, if you want to list all users, you would use the `Read` method with the `users` object.

### Example Usage

```
// Call Read to list records in an object
res, err := client.Read(context.TODO(),common.ReadParams{
		ObjectName: "users",
        })
if err != nil {
	log.Fatal(err)
}

```


