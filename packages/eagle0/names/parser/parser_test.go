package parser

import (
	"github.com/google/go-cmp/cmp"
	"github.com/nolen777/name-generator/packages/eagle0/names/token"
	"testing"
)

func TestParseLiteral(t *testing.T) {
	formatString := "\"Hello\" 3432"

	remaining, tok, err := parseLiteral(formatString)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if remaining != " 3432" {
		t.Errorf("Expected remaining string to be ' 3432', got '%s'", remaining)
	}
	if tok.Literal != "Hello" {
		t.Errorf("Expected token to be 'Hello', got '%v'", tok)
	}
}

func TestParseSmallLiteral(t *testing.T) {
	formatString := "\"hi\""

	remaining, tok, err := parseLiteral(formatString)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if remaining != "" {
		t.Errorf("Expected remaining string to be empty, got '%s'", remaining)
	}
	if tok.Literal != "hi" {
		t.Errorf("Expected token to be 'hi', got '%v'", tok)
	}
}

func TestParseUnicodeLiteral(t *testing.T) {
	formatString := "\"hi \u1F600\""

	remaining, tok, err := parseLiteral(formatString)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if remaining != "" {
		t.Errorf("Expected remaining string to be empty, got '%s'", remaining)
	}
	if tok.Literal != "hi \u1F600" {
		t.Errorf("Expected token to be 'hi \u1F600', got '%v'", tok)
	}
}

func TestParseTitle(t *testing.T) {
	formatString := "-\"hi\"+"

	result, err := ParseFrom(formatString)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := token.TitleCaseToken{Base: token.LiteralToken{Literal: "hi"}}
	if result != expected {
		t.Errorf("Expected token to be %v, got '%v'", expected, result)
	}
}

func TestParseOneof(t *testing.T) {
	formatString := "[0.25 \"hi\", 0.1 \"ho\", 0.75 \"bye\"]"

	tok, err := ParseFrom(formatString)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := token.OneofListToken{
		Entries: []token.OneofListEntry{
			{Token: token.LiteralToken{Literal: "hi"}, Weight: 0.25},
			{Token: token.LiteralToken{Literal: "ho"}, Weight: 0.1},
			{Token: token.LiteralToken{Literal: "bye"}, Weight: 0.75},
		},
	}

	if !cmp.Equal(tok, expected) {
		t.Errorf("Expected token to be %v, got '%v'", expected, tok)
	}
}

func TestSpacesInOneof(t *testing.T) {
	formatString := "[0.3 \" \", 0.7 \"-\"]"

	tok, err := ParseFrom(formatString)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := token.OneofListToken{
		Entries: []token.OneofListEntry{
			{Token: token.LiteralToken{Literal: " "}, Weight: 0.3},
			{Token: token.LiteralToken{Literal: "-"}, Weight: 0.7},
		},
	}

	if !cmp.Equal(tok, expected) {
		t.Errorf("Expected token to be %v, got '%v'", expected, tok)
	}
}

func TestParseNestedOneof(t *testing.T) {
	formatString := "[0.2 [0.3 \" \", 0.7 \"-\"]]"

	tok, err := ParseFrom(formatString)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := token.OneofListToken{
		Entries: []token.OneofListEntry{
			{Weight: 0.2, Token: token.OneofListToken{
				Entries: []token.OneofListEntry{
					{Token: token.LiteralToken{Literal: " "}, Weight: 0.3},
					{Token: token.LiteralToken{Literal: "-"}, Weight: 0.7}},
			},
			},
		},
	}

	if !cmp.Equal(tok, expected) {
		t.Errorf("Expected token to be %v, got '%v'", expected, tok)
	}
}

func TestParseOptional(t *testing.T) {
	formatString := "{0.25 \"hi\"}"

	tok, err := ParseFrom(formatString)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := token.OptionalToken{
		Token: token.LiteralToken{Literal: "hi"},
		Odds:  0.25,
	}

	if !cmp.Equal(tok, expected) {
		t.Errorf("Expected token to be %v, got '%v'", expected, tok)
	}
}

func TestParseListSelector(t *testing.T) {
	formatString := "$names"

	tok, err := ParseFrom(formatString)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := token.ListSelectionToken{
		ChoiceListName: "names",
		Filtered:       true,
	}
	if !cmp.Equal(tok, expected) {
		t.Errorf("Expected token to be %v, got '%v'", expected, tok)
	}
}

func TestParseOrdinal(t *testing.T) {
	formatString := "%192"

	tok, err := ParseFrom(formatString)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := token.OrdinalSelectionToken{
		Max: 192,
	}
	if !cmp.Equal(tok, expected) {
		t.Errorf("Expected token to be %v, got '%v'", expected, tok)
	}
}

func TestParseSubstitution(t *testing.T) {
	formatString := "@FOO"

	tok, err := ParseFrom(formatString)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := token.SubstitutionToken{
		Key: "FOO",
	}
	if !cmp.Equal(tok, expected) {
		t.Errorf("Expected token to be %v, got '%v'", expected, tok)
	}
}

func TestParseUnfilteredList(t *testing.T) {
	formatString := "#foo"

	tok, err := ParseFrom(formatString)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := token.ListSelectionToken{
		ChoiceListName: "foo",
		Filtered:       false,
	}
	if !cmp.Equal(tok, expected) {
		t.Errorf("Expected token to be %v, got '%v'", expected, tok)
	}
}

func TestReadDecimal(t *testing.T) {
	formatString, err := parseNext("0.25 \"hi\"", parseSequence{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	dec, remaining, err := readDecimal(formatString)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expectedRemaining, _ := parseNext(" \"hi\"", parseSequence{})
	if !cmp.Equal(remaining, expectedRemaining) {
		t.Errorf("Expected remaining string to be %v, got '%s'", expectedRemaining, remaining)
	}
	if dec != 0.25 {
		t.Errorf("Expected token to be '0.25', got '%v'", dec)
	}
}

func TestParseFrom(t *testing.T) {
	formatString := "{ 0.35 [0.75 $names, 0.25 \"Bumbler\", 0.1 %25, 0.4 @KEY] }"

	tok, err := ParseFrom(formatString)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := token.OptionalToken{
		Token: token.OneofListToken{
			Entries: []token.OneofListEntry{
				{Token: token.ListSelectionToken{ChoiceListName: "names", Filtered: true}, Weight: 0.75},
				{Token: token.LiteralToken{Literal: "Bumbler"}, Weight: 0.25},
				{Token: token.OrdinalSelectionToken{Max: 25}, Weight: 0.1},
				{Token: token.SubstitutionToken{Key: "KEY"}, Weight: 0.4},
			},
		},
		Odds: 0.35,
	}
	if !cmp.Equal(tok, expected) {
		t.Errorf("Expected token to be %v, got '%v'", expected, tok)
	}
}
