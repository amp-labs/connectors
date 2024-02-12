package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/salesforce"
	"golang.org/x/oauth2"
)

func main() {
	// cfg := &oauth2.Config{
	// 	ClientID:     "f49c98fd-02ce-4fea-bf88-a2425a2897d5",
	// 	ClientSecret: "cae05fe0-b474-4637-902d-bb0d60dc70b8",
	// 	Endpoint: oauth2.Endpoint{
	// 		AuthURL:   fmt.Sprintf("https://app.hubspot.com/oauth/authorize"),
	// 		TokenURL:  fmt.Sprintf("https://api.hubapi.com/oauth/v1/token"),
	// 		AuthStyle: oauth2.AuthStyleInParams,
	// 	},
	// }

	cfg := &oauth2.Config{
		ClientID:     "f49c98fd-02ce-4fea-bf88-a2425a2897d5",
		ClientSecret: "cae05fe0-b474-4637-902d-bb0d60dc70b8",
		Endpoint: oauth2.Endpoint{
			AuthURL:   fmt.Sprintf("https://ampersand-dev-ed.develop.my.salesforce.com/services/oauth2/authorize"),
			TokenURL:  fmt.Sprintf("https://ampersand-dev-ed.develop.my.salesforce.com/services/oauth2/token"),
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	// // Set up the OAuth2 token (obtained from Salesforce by authenticating)
	// tok := &oauth2.Token{
	// 	AccessToken:  "CIWZmq7PMRIRAAEBQAAAAAAAAAAYAAAAAAEYgYSMFSC6qKUdKLG3jQEyFIwXzTeCtUV_R1sEMlGebVGRKwsAOjUAAABBAAAAAMAHAAAAAAAAAIAAAAAAAAAADAAAAAAAAABgAAAAAAAAAPwAAAAAIAMAAAAAHEIUhVO0EWLup93kI5lftdxoZrJt5f1KA25hMVIAWgA",
	// 	RefreshToken: "b97e4700-321e-4a09-bd41-203c5b5c9fe4",
	// 	TokenType:    "bearer",
	// }

	// {
	// 	"access_token": ,
	// 	"token_type": "bearer",
	// 	"refresh_token": "b97e4700-321e-4a09-bd41-203c5b5c9fe4",
	// 	"expiry": "2024-01-10T15:47:59.287121-08:00"
	// }

	// Set up the OAuth2 token (obtained from Salesforce by authenticating)
	tok := &oauth2.Token{
		AccessToken:  "00DDp000000JQ4L!ASAAQLHVDPMX0LMLjQqnIFacPkIgvts8xs.p8yZstH0VVYJpyJxBnDE2mVefVmP0uMHVFNtxp43FT1fzY7DGhBKoS2ecb1Z0",
		RefreshToken: "5Aep861D25v7ERuX692Q9XrcV0pSW6qXfwe2fRtFbhzmz8.b82OqL5oxLxr_e7OoWhJjUG6FJLtBXyiT1ek8acz",
		TokenType:    "bearer",
	}

	// {
	// 	"access_token": ,
	// 	"token_type": "Bearer",
	// 	"refresh_token": ,
	// 	"expiry": "0001-01-01T00:00:00Z"
	// }

	// Create the Salesforce client
	// client, err := connectors.Hubspot.New(
	// 	hubspot.WithClient(context.Background(), http.DefaultClient, cfg, tok),
	// 	hubspot.WithModule(hubspot.ModuleCRM))
	// if err != nil {
	// 	panic(err)
	// }

	client, err := connectors.Salesforce.New(
		salesforce.WithClient(context.Background(), http.DefaultClient, cfg, tok),
		salesforce.WithSubdomain("ampersand-dev-ed.develop"))
	if err != nil {
		panic(err)
	}

	// client, err := connectors.New("hubspot", map[string]any{
	// 	"client": hubspot.WithClient(context.Background(), http.DefaultClient, cfg, tok),
	// 	"module": hubspot.ModuleCRM,
	// })

	// client, err := hubspot.NewConnector(
	// 	hubspot.WithClient(context.Background(), http.DefaultClient, cfg, tok, common.WithTokenUpdated(logToken)),
	// 	hubspot.WithModule(hubspot.ModuleCRM),
	// )
	//
	result, err := client.Read(context.Background(), common.ReadParams{
		ObjectName: "Account",
		Fields:     []string{"name"},
		// Since:      time.Now().Add(-111111 * time.Minute),
		// Deleted:    true,
	})

	// result, err := client.ListObjectMetadata(context.Background(), []string{"Account"})
	if err != nil {
		fmt.Println(err)
	} else {
		marshaled, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(string(marshaled))
		}
	}

	// result, err := client.Read(context.Background(), common.ReadParams{
	// 	ObjectName: "account",
	// 	Fields:     []string{"Id", "Name"},
	// })
	// if err != nil {
	// 	fmt.Println(err)
	// }
}

func logToken(oldToken *oauth2.Token, newToken *oauth2.Token) error {
	// fmt.Println("oldToken", oldToken)
	// fmt.Println("newToken", newToken)

	return nil
}
