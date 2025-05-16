package token

import (
	"testing"
)

var emptyContext = StringConstructionContext{}

type fixedRandomSource struct {
	IntNValue    int
	Float64Value float64
}

func (s fixedRandomSource) Intn(n int) int {
	return s.IntNValue
}
func (s fixedRandomSource) Float64() float64 {
	return s.Float64Value
}

func TestLiteralToken(t *testing.T) {
	literal := "Hello"
	token := LiteralToken{Literal: literal}
	result, err := token.Next(nil, emptyContext)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != literal {
		t.Errorf("Expected %s, got %s", literal, result)
	}
}

func TestSubstitutionToken_errorsIfKeyNotPresent(t *testing.T) {
	token := SubstitutionToken{Key: "missing_key"}

	_, err := token.Next(nil, emptyContext)
	if err == nil {
		t.Errorf("Expected error, got nil")
	} else if err.Error() != "missing key: missing_key" {
		t.Errorf("Expected error 'missing key: missing_key', got '%s'", err.Error())
	}
}

func TestSubstitutionToken(t *testing.T) {
	token := SubstitutionToken{Key: "key"}
	literalSubstitutions := map[string]string{
		"key": "value",
	}

	result, err := token.Next(nil, StringConstructionContext{LiteralSubstitutions: literalSubstitutions})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "value" {
		t.Errorf("Expected 'value', got '%s'", result)
	}
}

func TestSequenceToken(t *testing.T) {
	token1 := LiteralToken{Literal: "Hello"}
	token2 := LiteralToken{Literal: "World"}
	token3 := LiteralToken{Literal: "!"}

	token := SequenceToken{Tokens: []StringConstructionToken{token1, token2, token3}}
	result, err := token.Next(nil, emptyContext)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := "HelloWorld!"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestOrdinalSelectionToken(t *testing.T) {
	token := OrdinalSelectionToken{Max: 26}

	r := fixedRandomSource{IntNValue: 12}
	result, err := token.Next(r, emptyContext)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "13th" {
		t.Errorf("Expected '13th', got '%s'", result)
	}
}

func TestOptionalToken_lowValueReturnsToken(t *testing.T) {
	optionalToken := OptionalToken{
		Token: LiteralToken{Literal: "Optional"},
		Odds:  0.3,
	}
	r := fixedRandomSource{Float64Value: 0.2}

	result, err := optionalToken.Next(r, emptyContext)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "Optional" {
		t.Errorf("Expected 'Optional', got '%s'", result)
	}
}

func TestOptionalToken_highValueReturnsEmpty(t *testing.T) {
	optionalToken := OptionalToken{
		Token: LiteralToken{Literal: "Optional"},
		Odds:  0.3,
	}
	r := fixedRandomSource{Float64Value: 0.4}

	result, err := optionalToken.Next(r, emptyContext)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "" {
		t.Errorf("Expected '', got '%s'", result)
	}
}

func TestOneofListToken_lowRoll(t *testing.T) {
	oneofToken := OneofListToken{
		Entries: []OneofListEntry{
			{Token: LiteralToken{Literal: "Option1"}, Weight: 0.5},
			{Token: LiteralToken{Literal: "Option2"}, Weight: 0.3},
			{Token: LiteralToken{Literal: "Option3"}, Weight: 0.2},
		},
	}

	r := fixedRandomSource{Float64Value: 0.1}

	result, err := oneofToken.Next(r, emptyContext)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "Option1" {
		t.Errorf("Expected 'Option1', got '%s'", result)
	}
}

func TestOneofListToken_middleRoll(t *testing.T) {
	oneofToken := OneofListToken{
		Entries: []OneofListEntry{
			{Token: LiteralToken{Literal: "Option1"}, Weight: 0.5},
			{Token: LiteralToken{Literal: "Option2"}, Weight: 0.3},
			{Token: LiteralToken{Literal: "Option3"}, Weight: 0.2},
		},
	}

	r := fixedRandomSource{Float64Value: 0.6}

	result, err := oneofToken.Next(r, emptyContext)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "Option2" {
		t.Errorf("Expected 'Option2', got '%s'", result)
	}
}

func TestOneofListToken_highRoll(t *testing.T) {
	oneofToken := OneofListToken{
		Entries: []OneofListEntry{
			{Token: LiteralToken{Literal: "Option1"}, Weight: 0.5},
			{Token: LiteralToken{Literal: "Option2"}, Weight: 0.3},
			{Token: LiteralToken{Literal: "Option3"}, Weight: 0.2},
		},
	}

	r := fixedRandomSource{Float64Value: 0.9}

	result, err := oneofToken.Next(r, emptyContext)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "Option3" {
		t.Errorf("Expected 'Option3', got '%s'", result)
	}
}

func TestListSelectionToken(t *testing.T) {
	token := ListSelectionToken{
		ChoiceListName: "names",
		Filtered:       true,
	}

	contextWithChoices := StringConstructionContext{
		ChoiceListMap: map[string][]string{
			"names": {"Option1", "Option2", "Option3"},
		},
	}

	r := fixedRandomSource{IntNValue: 0}

	result, err := token.Next(r, contextWithChoices)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "Option1" {
		t.Errorf("Expected 'Option1', got '%s'", result)
	}
}

func TestTitleCaseToken(t *testing.T) {
	token := TitleCaseToken{
		Base: LiteralToken{Literal: "hello world"},
	}

	result, err := token.Next(nil, emptyContext)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := "Hello World"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestTitleCaseToken_uncapitalizedWords(t *testing.T) {
	token := TitleCaseToken{
		Base: LiteralToken{Literal: "hello to the world"},
	}

	result, err := token.Next(nil, emptyContext)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := "Hello to the World"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestTitleCaseToken_preserveExistingCapitals(t *testing.T) {
	token := TitleCaseToken{
		Base: LiteralToken{Literal: "hello wOrLD"},
	}

	result, err := token.Next(nil, emptyContext)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := "Hello WOrLD"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}
