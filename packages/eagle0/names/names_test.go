package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func nameCount(html string) int {
	lines := strings.Split(html, "\r\n")
	count := 0
	for _, line := range lines {
		if len(line) > 0 && line[0] != '<' {
			count++
		}
	}
	return count
}

func TestNames_noParams(t *testing.T) {
	// Test with a valid name
	event := Event{}
	ctx := context.WithValue(context.Background(), "function_version", "1.0")
	response := Names(ctx, event)
	if response.StatusCode != "200" {
		t.Errorf("Expected status code to be '200', got '%s'", response.StatusCode)
	}

	count := nameCount(response.Body)
	if count != 20 {
		t.Errorf("Expected body to contain 20 names, got '%d'", count)
	}
}

func TestNames_withParams(t *testing.T) {
	event := Event{
		Requests: []NameRequest{
			{Id: "4h", Gender: "female"},
			{Id: "6", Gender: "male"},
		},
	}
	ctx := context.WithValue(context.Background(), "function_version", "1.0")
	response := Names(ctx, event)
	if response.StatusCode != "200" {
		t.Errorf("Expected status code to be '200', got '%s'", response.StatusCode)
	}

	count := nameCount(response.Body)
	if count != 2 {
		t.Errorf("Expected body to contain 2 names, got '%d'", count)
	}
}

func TestNames_acceptJson(t *testing.T) {
	event := Event{
		Http: httpInfo{Headers: headers{
			Accept: "application/json",
		}},
	}
	ctx := context.WithValue(context.Background(), "function_version", "1.0")
	response := Names(ctx, event)
	if response.StatusCode != "200" {
		t.Errorf("Expected status code to be '200', got '%s'", response.StatusCode)
	}

	var jb jsonBody
	err := json.Unmarshal([]byte(response.Body), &jb)
	if err != nil {
		t.Errorf("Expected valid JSON, got error: %v", err)
	}

	if len(jb.Names) != 20 {
		t.Errorf("Expected 20 names, got %d", len(jb.Names))
	}
}

func TestNames_noSpaceBeforeComma(t *testing.T) {
	event := Event{
		Http: httpInfo{Headers: headers{
			Accept: "text/html",
		}},
	}

	ctx := context.WithValue(context.Background(), "function_version", "1.0")
	response := Names(ctx, event).Body

	if strings.Contains(response, " ,") {
		t.Errorf("Expected no space before comma in HTML response")
	}
}
