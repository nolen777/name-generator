package token_generator

import (
	"github.com/nolen777/name-generator/packages/eagle0/names/token"
	"slices"
	"strings"
)

func CreateTokenFromMapAndFormatStringWithSubstitutions(
	namesMap map[string][]string,
	formatString string,
	atSpecializations []string,
	substitutions map[string]string,
) token.StringConstructionToken {
	formatStringWithSubstitutions := formatString
	if substitutions != nil {
		for key, value := range substitutions {
			formatStringWithSubstitutions = strings.ReplaceAll(formatStringWithSubstitutions, key, value)
		}
	}

	stringsMap := map[string][]string{}
	unfilteredStringsMap := map[string][]string{}

	for key, values := range namesMap {
		unfilteredStringsMap[key] = append(unfilteredStringsMap[key], values...)
		keyParts := strings.Split(key, "@")
		if len(keyParts) == 1 {
			stringsMap[key] = append(stringsMap[key], values...)
		} else if slices.Contains(atSpecializations, keyParts[1]) {
			stringsMap[keyParts[0]] = append(stringsMap[keyParts[0]], values...)
		}
	}
	// get a token from the parser

	return nil
}
