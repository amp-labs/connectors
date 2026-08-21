// nolint:forbidigo
package main

import (
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/tools/fileconv"
	"github.com/amp-labs/connectors/tools/scrapper"
)

type Method struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Scope struct {
	Name        string   `json:"name"`
	URL         string   `json:"url"`
	Description string   `json:"description"`
	TokenTypes  []string `json:"tokenTypes"`
	Methods     []Method `json:"methods"`
}

type GroupedMethods struct {
	Both     map[string]string `json:"both"`
	UserOnly map[string]string `json:"userOnly"`
	BotOnly  map[string]string `json:"botOnly"`
}

func main() {
	var scopes []Scope
	goutils.MustBeNil(scrapper.LoadFile("scripts/scraper/slack/2-methods/scopes.json", &scopes))

	grouped := groupMethodsByTokenType(scopes)

	goutils.MustBeNil(fileconv.Flusher{}.ToFile("scripts/scraper/slack/3-views/operations.json", grouped))
}

func groupMethodsByTokenType(scopes []Scope) GroupedMethods {
	botMethods := make(datautils.Map[string, string])
	userMethods := make(datautils.Map[string, string])

	for _, scope := range scopes {
		for _, method := range scope.Methods {
			for _, tokenType := range scope.TokenTypes {
				switch tokenType {
				case "User":
					userMethods[method.Name] = method.URL
				case "Bot":
					botMethods[method.Name] = method.URL
				}
			}
		}
	}

	botMethodNames := botMethods.KeySet()
	userMethodNames := userMethods.KeySet()

	both := botMethodNames.Intersection(userMethodNames)
	userOnly := userMethodNames.Subtract(botMethodNames)
	botOnly := botMethodNames.Subtract(userMethodNames)

	return GroupedMethods{
		Both:     botMethods.ShallowSubset(both),
		UserOnly: userMethods.ShallowSubset(userOnly),
		BotOnly:  botMethods.ShallowSubset(botOnly),
	}
}
