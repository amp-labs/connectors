# Gong connector


## Read
Read is used to list all records of a given type. For example, if you want to list all users, you would use the `Read` method with the `users` object or `calls`.

### Example Usage

```

gong := GetGongConnector(context.Background(), DefaultCredsFile)

	config := connectors.ReadParams{
		ObjectName: "calls", 
		Fields:     []string{"url"},
	}

	result, err := gong.Read(context.Background(), config)
	if err != nil {
		slog.Error("Error reading from Gong", "error", err)
		return 1
	}

```


