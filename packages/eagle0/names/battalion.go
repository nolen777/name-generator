package main

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"

	"github.com/nolen777/name-generator/packages/eagle0/names/parser"
	"github.com/nolen777/name-generator/packages/eagle0/names/spaces_fetcher"
	"github.com/nolen777/name-generator/packages/eagle0/names/token"
)

var battalionTypes = []string{
	"light_infantry", "heavy_infantry", "light_cavalry", "heavy_cavalry", "longbowmen", "undead",
}
var battalionTokens map[string]token.StringConstructionToken

func initializeBattalionTokens() {
	rawTemplate, err := spaces_fetcher.GetFile("battalionNameConstruction.txt")
	if err != nil {
		panic(err)
	}
	battalionTokens, err = compileBattalionTokens(string(rawTemplate))
	if err != nil {
		panic(err)
	}
}

func compileBattalionTokens(template string) (map[string]token.StringConstructionToken, error) {
	tokens := make(map[string]token.StringConstructionToken, len(battalionTypes))
	template = strings.ReplaceAll(template, "\r", "")
	for _, battalionType := range battalionTypes {
		format := strings.NewReplacer(
			"$BATTALION_ADJECTIVE", "$"+battalionType+"_adj",
			"$BATTALION_NAME", "$"+battalionType,
		).Replace(template)
		tok, err := parser.ParseFrom(format)
		if err != nil {
			return nil, fmt.Errorf("invalid battalion template for %s: %w", battalionType, err)
		}
		tokens[battalionType] = tok
	}
	return tokens, nil
}

func validateRequest(request NameRequest) error {
	switch request.Kind {
	case "", "hero":
		return nil
	case "battalion":
		supported := false
		for _, battalionType := range battalionTypes {
			if request.BattalionType == battalionType {
				supported = true
				break
			}
		}
		if !supported {
			return fmt.Errorf("request %q: unsupported battalionType %q", request.Id, request.BattalionType)
		}
		if strings.TrimSpace(request.ProvinceName) == "" {
			return fmt.Errorf("request %q: provinceName is required for battalion names", request.Id)
		}
		return nil
	default:
		return fmt.Errorf("request %q: unsupported kind %q", request.Id, request.Kind)
	}
}

func requestError(err error, accept string) Response {
	if accept == "application/json" {
		body, _ := json.Marshal(map[string]string{"error": err.Error()})
		return Response{Body: string(body), StatusCode: "400", Headers: ResponseHeaders{ContentType: "application/json"}}
	}
	return Response{Body: "<html><h1>" + html.EscapeString(err.Error()) + "</h1></html>", StatusCode: "400", Headers: ResponseHeaders{ContentType: "text/html"}}
}
