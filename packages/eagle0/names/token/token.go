package token

import (
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strconv"
	"strings"
)

type StringConstructionContext struct {
	ChoiceListMap           map[string][]string
	UnfilteredChoiceListMap map[string][]string
	LiteralSubstitutions    map[string]string
}

type TokenRandomSource interface {
	Float64() float64
	Intn(n int) int
}

type StringConstructionToken interface {
	Next(rand TokenRandomSource, ctx StringConstructionContext) (string, error)
}

type LiteralToken struct {
	Literal string
}

func (token LiteralToken) Next(rand TokenRandomSource, ctx StringConstructionContext) (string, error) {
	return token.Literal, nil
}

type SubstitutionToken struct {
	Key string
}

func (token SubstitutionToken) Next(rand TokenRandomSource, ctx StringConstructionContext) (string, error) {
	value, ok := ctx.LiteralSubstitutions[token.Key]

	if !ok {
		return "", fmt.Errorf("missing key: %s", token.Key)
	}
	return value, nil
}

type SequenceToken struct {
	Tokens []StringConstructionToken
}

func (token SequenceToken) Next(rand TokenRandomSource, ctx StringConstructionContext) (string, error) {
	result := ""
	for _, token := range token.Tokens {
		next, err := token.Next(rand, ctx)
		if err != nil {
			return "", err
		}
		result += next
	}
	return result, nil
}

type OptionalToken struct {
	Token StringConstructionToken
	Odds  float64
}

func (token OptionalToken) Next(rand TokenRandomSource, ctx StringConstructionContext) (string, error) {
	r := rand.Float64()
	if r < token.Odds {
		return token.Token.Next(rand, ctx)
	}
	return "", nil
}

type OneofListEntry struct {
	Token  StringConstructionToken
	Weight float64
}

func (entry OneofListEntry) ToString(rand TokenRandomSource, ctx StringConstructionContext) (string, error) {
	return entry.Token.Next(rand, ctx)
}

type OneofListToken struct {
	Entries []OneofListEntry
}

func (token OneofListToken) Next(rand TokenRandomSource, ctx StringConstructionContext) (string, error) {
	totalWeight := 0.0
	for _, entry := range token.Entries {
		totalWeight += entry.Weight
	}

	randomValue := rand.Float64() * totalWeight
	for _, entry := range token.Entries {
		randomValue -= entry.Weight
		if randomValue <= 0 {
			return entry.ToString(rand, ctx)
		}
	}
	panic("Should not reach here")
}

type ListSelectionToken struct {
	ChoiceListName string
	Filtered       bool
}

func (token ListSelectionToken) Next(rand TokenRandomSource, ctx StringConstructionContext) (string, error) {
	var list []string
	var ok bool
	if token.Filtered {
		list, ok = ctx.ChoiceListMap[token.ChoiceListName]
	} else {
		list, ok = ctx.UnfilteredChoiceListMap[token.ChoiceListName]
	}
	if !ok {
		return "", fmt.Errorf("missing list: %s", token.ChoiceListName)
	}
	return list[rand.Intn(len(list))], nil
}

type OrdinalSelectionToken struct {
	Max int
}

func (token OrdinalSelectionToken) Next(rand TokenRandomSource, ctx StringConstructionContext) (string, error) {
	value := rand.Intn(token.Max-1) + 1
	aval := strconv.Itoa(value)
	if value%100 == 11 || value%100 == 12 || value%100 == 13 {
		return aval + "th", nil
	}
	switch value % 10 {
	case 1:
		return aval + "st", nil
	case 2:
		return aval + "nd", nil
	case 3:
		return aval + "rd", nil
	default:
		return aval + "th", nil
	}
}

var uncapitalizedWords = map[string]bool{
	"and": true, "but": true, "for": true, "or": true, "nor": true, "the": true, "a": true, "an": true, "to": true, "as": true, "of": true,
}

type TitleCaseToken struct {
	Base StringConstructionToken
}

func (token TitleCaseToken) Next(rand TokenRandomSource, ctx StringConstructionContext) (string, error) {
	caser := cases.Title(language.AmericanEnglish, cases.NoLower)
	str, err := token.Base.Next(rand, ctx)
	if err != nil {
		return "", err
	}

	words := strings.Split(str, " ")
	newWords := make([]string, len(words))

	for i, word := range words {
		if i == 0 || i == len(words)-1 {
			newWords[i] = caser.String(word)
		} else if _, ok := uncapitalizedWords[word]; ok {
			newWords[i] = strings.ToLower(word)
		} else {
			newWords[i] = caser.String(word)
		}
	}
	return strings.Join(newWords, " "), nil
}
