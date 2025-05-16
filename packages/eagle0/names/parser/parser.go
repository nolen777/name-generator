package parser

import (
	"fmt"
	"github.com/nolen777/name-generator/packages/eagle0/names/token"
	"strconv"
	"strings"
	"unicode"
)

func ParseFrom(formatString string) (token.StringConstructionToken, error) {
	result, err := parseNext(formatString, parseSequence{})
	if err != nil {
		return nil, err
	}

	parsedResult, err := tokenize(result)
	if err != nil {
		return nil, err
	}
	if len(parsedResult.Remaining) != 0 {
		return nil, err
	}
	return parsedResult.ParsedToken, nil
}

func parseNext(remaining string, acc parseSequence) (parseSequence, error) {
	if remaining == "" {
		return acc, nil
	}

	switch remaining[0] {
	case ' ', '\n':
		return parseNext(remaining[1:], acc)
	case '"':
		newRemaining, tok, err := parseLiteral(remaining)
		if err != nil {
			return nil, err
		}
		return parseNext(newRemaining, append(acc, t{tok}))
	default:
		return parseNext(remaining[1:], append(acc, character{rune(remaining[0])}))
	}
}

type characterOrToken interface {
	Equals(other characterOrToken) bool
	IsLetterOrUnderscore() bool
	IsDigit() bool
}

type character struct {
	R rune
}

func (c character) Equals(other characterOrToken) bool {
	if otherChar, ok := other.(character); ok {
		return c.R == otherChar.R
	}
	return false
}
func (c character) IsLetterOrUnderscore() bool {
	return unicode.IsLetter(c.R) || c.R == '_'
}
func (c character) IsDigit() bool {
	return unicode.IsDigit(c.R)
}

type t struct {
	T token.StringConstructionToken
}

func (t t) Equals(other characterOrToken) bool {
	return false
}
func (t t) IsLetterOrUnderscore() bool {
	return false
}
func (t t) IsDigit() bool {
	return false
}

type parseSequence []characterOrToken

func ToString(ts parseSequence) string {
	str := ""
	for _, cOrT := range ts {
		switch cOrT.(type) {
		case character:
			str += string(cOrT.(character).R)
		case t:
			str += "|TOK|"
		}
	}
	return str
}

type parseResult struct {
	Remaining   parseSequence
	ParsedToken token.StringConstructionToken
}

func parseLiteral(formatString string) (string, token.LiteralToken, error) {
	if !strings.HasPrefix(formatString, "\"") {
		return formatString, token.LiteralToken{}, nil
	}
	remaining := formatString[1:]
	acc := ""

	newRemaining := remaining
	for i, r := range remaining {
		newRemaining = remaining[i:]
		switch r {
		case '"':
			return newRemaining[1:], token.LiteralToken{Literal: acc}, nil

		default:
			acc += string(r)
		}
	}

	return formatString, token.LiteralToken{}, fmt.Errorf("Missing closing \" in literal")
}

func insideBalanced(ts parseSequence, open rune, closed rune) (parseSequence, parseSequence, error) {
	if len(ts) == 0 {
		return nil, nil, fmt.Errorf("Empty sequence")
	}
	if !ts[0].Equals(character{open}) {
		return nil, nil, fmt.Errorf("Expected %c at start of sequence", open)
	}
	acc := parseSequence{}
	openCount := 1
	for i, cOrT := range ts[1:] {
		if cOrT.Equals(character{open}) {
			openCount++
		} else if cOrT.Equals(character{closed}) {
			openCount--

			if openCount == 0 {
				return ts[i+2:], acc, nil
			}
		}
		acc = append(acc, cOrT)
	}

	str := ToString(ts)
	return nil, nil, fmt.Errorf("Missing closing %c in sequence: %s", closed, str)
}

func parseListSelector(ts parseSequence) (parseResult, error) {
	if len(ts) == 0 || !ts[0].Equals(character{'$'}) {
		return parseResult{ts, nil}, fmt.Errorf("Expected $ at start of list selector")
	}

	listName, remaining, err := readString(ts[1:])
	if err != nil {
		return parseResult{ts, nil}, err
	}
	if listName == "" {
		return parseResult{ts, nil}, fmt.Errorf("Empty list name")
	}
	return parseResult{Remaining: remaining, ParsedToken: token.ListSelectionToken{ChoiceListName: listName, Filtered: true}}, nil
}

func parseUnfilteredListSelector(ts parseSequence) (parseResult, error) {
	if len(ts) == 0 || !ts[0].Equals(character{'#'}) {
		return parseResult{ts, nil}, fmt.Errorf("Expected # at start of unfiltered list selector")
	}

	listName, remaining, err := readString(ts[1:])
	if err != nil {
		return parseResult{ts, nil}, err
	}
	if listName == "" {
		return parseResult{ts, nil}, fmt.Errorf("Empty list name")
	}
	return parseResult{Remaining: remaining, ParsedToken: token.ListSelectionToken{ChoiceListName: listName}}, nil
}

func parseTitle(ts parseSequence) (parseResult, error) {
	remaining, inner, err := insideBalanced(ts, '-', '+')
	if err != nil {
		return parseResult{ts, nil}, err
	}

	innerParsed, err := tokenize(inner)
	if err != nil {
		return parseResult{ts, nil}, err
	}
	if innerParsed.ParsedToken == nil {
		return parseResult{ts, nil}, fmt.Errorf("Empty inner result")
	}
	if innerParsed.Remaining != nil && len(innerParsed.Remaining) > 0 {
		return parseResult{ts, nil}, fmt.Errorf("Expected empty remaining after inner parse")
	}

	return parseResult{Remaining: remaining, ParsedToken: token.TitleCaseToken{Base: innerParsed.ParsedToken}}, nil
}

func parseOrdinal(ts parseSequence) (parseResult, error) {
	if len(ts) == 0 || !ts[0].Equals(character{'%'}) {
		return parseResult{ts, nil}, fmt.Errorf("Expected %% at start of ordinal")
	}
	ts = ts[1:]

	maxString := ""
	for len(ts) > 0 {
		h := ts[0]

		if h.IsDigit() {
			maxString += string(h.(character).R)
			ts = ts[1:]
		} else {
			break
		}
	}

	if maxString == "" {
		return parseResult{ts, nil}, fmt.Errorf("Empty ordinal max value")
	}
	maxVal, err := strconv.Atoi(maxString)
	if err != nil {
		return parseResult{ts, nil}, fmt.Errorf("Invalid ordinal max value: %s", maxString)
	}
	return parseResult{Remaining: ts, ParsedToken: token.OrdinalSelectionToken{Max: maxVal}}, nil
}

func parseSubstitution(ts parseSequence) (parseResult, error) {
	if len(ts) == 0 || !ts[0].Equals(character{'@'}) {
		return parseResult{ts, nil}, fmt.Errorf("Expected %% at start of ordinal")
	}

	str, remaining, err := readString(ts[1:])
	if err != nil {
		return parseResult{ts, nil}, err
	}
	if str == "" {
		return parseResult{ts, nil}, fmt.Errorf("Empty ordinal max value")
	}
	return parseResult{Remaining: remaining, ParsedToken: token.SubstitutionToken{Key: str}}, nil
}

func parseOptional(ts parseSequence) (parseResult, error) {
	remaining, inner, err := insideBalanced(ts, '{', '}')
	if err != nil {
		return parseResult{ts, nil}, err
	}

	dec, innerTok, err := readDecimalTokenPair(inner)
	if err != nil {
		return parseResult{ts, nil}, err
	}
	if innerTok.Remaining != nil && len(innerTok.Remaining) > 0 {
		return parseResult{ts, nil}, fmt.Errorf("Expected empty remaining after inner parse")
	}

	return parseResult{Remaining: remaining, ParsedToken: token.OptionalToken{Odds: dec, Token: innerTok.ParsedToken}}, nil
}

func parseOneof(ts parseSequence) (parseResult, error) {
	remaining, inner, err := insideBalanced(ts, '[', ']')
	if err != nil {
		return parseResult{ts, nil}, err
	}

	entries := []token.OneofListEntry{}
	for len(inner) > 0 {
		t := inner[0]
		if t.Equals(character{','}) {
			inner = inner[1:]
			continue
		}
		dec, innerResult, err := readDecimalTokenPair(inner)
		if err != nil {
			return parseResult{ts, nil}, err
		}
		inner = innerResult.Remaining
		entries = append(entries, token.OneofListEntry{Weight: dec, Token: innerResult.ParsedToken})
	}

	return parseResult{Remaining: remaining, ParsedToken: token.OneofListToken{Entries: entries}}, nil
}

func tokenize(ts parseSequence) (parseResult, error) {
	remaining := ts
	acc := []token.StringConstructionToken{}

	for len(remaining) > 0 {
		head := remaining[0]
		switch head.(type) {
		case t:
			acc = append(acc, head.(t).T)
			remaining = remaining[1:]
		case character:
			var pr parseResult
			var err error

			switch head.(character).R {
			case '{':
				pr, err = parseOptional(remaining)

			case '[':
				pr, err = parseOneof(remaining)

			case '$':
				pr, err = parseListSelector(remaining)

			case '#':
				pr, err = parseUnfilteredListSelector(remaining)

			case '-':
				pr, err = parseTitle(remaining)

			case '%':
				pr, err = parseOrdinal(remaining)

			case '@':
				pr, err = parseSubstitution(remaining)

			case ',':
				goto finish

			default:
				return parseResult{ParsedToken: nil, Remaining: ts}, fmt.Errorf("Unexpected character %c", head.(character).R)
			}

			if err != nil {
				return pr, err
			}
			acc = append(acc, pr.ParsedToken)
			remaining = pr.Remaining
		}
	}

finish:
	if len(acc) == 0 {
		return parseResult{ParsedToken: nil, Remaining: ts}, fmt.Errorf("Empty token sequence")
	}
	if len(acc) == 1 {
		return parseResult{ParsedToken: acc[0], Remaining: remaining}, nil
	}
	return parseResult{ParsedToken: token.SequenceToken{Tokens: acc}, Remaining: remaining}, nil
}

func splitParseSequence(ts parseSequence, r rune) ([]parseSequence, error) {
	entries := []parseSequence{}
	for len(ts) > 0 {
		nextEntry := parseSequence{}
		for len(ts) > 0 {
			head := ts[0]
			ts = ts[1:]

			if head.Equals(character{r}) {
				break
			}
			nextEntry = append(nextEntry, head)
		}
		entries = append(entries, nextEntry)
	}

	return entries, nil
}

func readString(ts parseSequence) (string, parseSequence, error) {
	str := ""

	for len(ts) > 0 {
		if ts[0].IsLetterOrUnderscore() {
			str += string(ts[0].(character).R)
			ts = ts[1:]
		} else {
			break
		}
	}

	return str, ts, nil
}

func readDecimal(ts parseSequence) (float64, parseSequence, error) {
	decPart := ""

	for len(ts) > 0 {
		if ts[0].Equals(character{'.'}) || ts[0].IsDigit() {
			decPart += string(ts[0].(character).R)
			ts = ts[1:]
		} else {
			break
		}
	}

	if decPart == "" {
		return 0, ts, fmt.Errorf("Expected decimal part")
	}
	decValue, err := strconv.ParseFloat(decPart, 64)
	return decValue, ts, err
}

func readDecimalTokenPair(ts parseSequence) (float64, parseResult, error) {
	decValue, rem, err := readDecimal(ts)
	if err != nil {
		return 0, parseResult{Remaining: rem}, err
	}
	innerParsed, err := tokenize(rem)
	if err != nil {
		return 0, parseResult{Remaining: rem}, err
	}
	return decValue, innerParsed, nil
}
