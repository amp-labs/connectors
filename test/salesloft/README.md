# Running instructions

## Step 1: prepare "salesloft-creds.json" file

Create a file called "salesloft-creds.json" in the root of the project with the following contents

    e.g. {
        "CLIENT_ID": "<client id goes here>",
        "CLIENT_SECRET": "<client secret goes here>",
        "ACCESS_TOKEN": "<access token goes here>",
        "REFRESH_TOKEN": "<refresh token goes here>"
    }

or export to an environment variable SALESLOFT_CRED_FILE by following command

$> export SALESLOFT_CRED_FILE=./salesloft-creds.json # or the path to your salesloft-creds.json file


In 1password, you can find a Salesloft creds.json file in the "Shared" vault. TODO this must be in 1password
Look for the title "Salesloft Sample OAuth Credentials".
The 1password item has an attached file called "creds.json" that contains the JSON.

## Step 2: run the following command

    $> go run test/salesloft/read/main.go