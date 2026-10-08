package main

import (
	"context"
	"encoding/json"
	"math/rand"
	"os"
	"strings"
	"testing"

	"github.com/nolen777/name-generator/packages/eagle0/names/parser"
)

// Exercise the production parser and handler using checked-in configuration,
// without requiring Spaces credentials or changing the production fetch path.
func TestMain(m *testing.M) {
	initializeOnce.Do(func() {
		words, err := os.ReadFile("names.tsv")
		if err != nil {
			panic(err)
		}
		femaleCtx, maleCtx, otherCtx = contextsFromTSV(string(words))
		heroTemplate, err := os.ReadFile("nameConstruction.txt")
		if err != nil {
			panic(err)
		}
		stringConstructionToken, err = parser.ParseFrom(strings.ReplaceAll(string(heroTemplate), "\r", ""))
		if err != nil {
			panic(err)
		}
		battalionTemplate, err := os.ReadFile("battalionNameConstruction.txt")
		if err != nil {
			panic(err)
		}
		battalionTokens, err = compileBattalionTokens(string(battalionTemplate))
		if err != nil {
			panic(err)
		}
	})
	os.Exit(m.Run())
}

func TestNamesMixedBatch(t *testing.T) {
	var event Event
	err := json.Unmarshal([]byte(`{"requests":[
        {"id":"legacy","gender":"female"},
        {"id":"hero","kind":"hero","gender":"male"},
        {"id":"infantry","kind":"battalion","battalionType":"heavy_infantry","provinceName":"Westmarch"},
        {"id":"undead","kind":"battalion","battalionType":"undead","provinceName":"Eastmarch"}
    ],"http":{"headers":{"accept":"application/json"}}}`), &event)
	if err != nil {
		t.Fatal(err)
	}
	response := Names(context.Background(), event)
	if response.StatusCode != "200" {
		t.Fatalf("unexpected response: %+v", response)
	}
	var body jsonBody
	if err := json.Unmarshal([]byte(response.Body), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Names) != len(event.Requests) {
		t.Fatalf("unexpected names: %+v", body)
	}
	for i, name := range body.Names {
		if name.Id != event.Requests[i].Id || strings.TrimSpace(name.Name) == "" {
			t.Fatalf("bad result at %d: %+v", i, name)
		}
	}
}

func TestBattalionNamesAllTypes(t *testing.T) {
	for _, battalionType := range battalionTypes {
		t.Run(battalionType, func(t *testing.T) {
			ctx := otherCtx
			ctx.LiteralSubstitutions = map[string]string{"PLACE": "Westmarch"}
			rng := rand.New(rand.NewSource(42))
			seenProvince := false
			for i := 0; i < 1000; i++ {
				name, err := battalionTokens[battalionType].Next(rng, ctx)
				if err != nil {
					t.Fatal(err)
				}
				if strings.TrimSpace(name) == "" || strings.Contains(name, "BATTALION_") || strings.Contains(name, "PLACE") {
					t.Fatalf("invalid generated name: %q", name)
				}
				seenProvince = seenProvince || strings.Contains(name, "Westmarch")
			}
			if !seenProvince {
				t.Fatal("province never substituted")
			}
		})
	}
}

func TestNamesInvalidRequests(t *testing.T) {
	requests := []NameRequest{
		{Id: "bad-kind", Kind: "province"},
		{Id: "bad-type", Kind: "battalion", BattalionType: "unknown", ProvinceName: "Westmarch"},
		{Id: "no-place", Kind: "battalion", BattalionType: "longbowmen"},
		{Id: "blank-place", Kind: "battalion", BattalionType: "light_cavalry", ProvinceName: "  "},
	}
	for _, request := range requests {
		t.Run(request.Id, func(t *testing.T) {
			response := Names(context.Background(), Event{
				Requests: []NameRequest{request},
				Http:     httpInfo{Headers: headers{Accept: "application/json"}},
			})
			if response.StatusCode != "400" || response.Headers.ContentType != "application/json" {
				t.Fatalf("unexpected response: %+v", response)
			}
			var body map[string]string
			if err := json.Unmarshal([]byte(response.Body), &body); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(body["error"], request.Id) {
				t.Fatalf("missing request ID: %+v", body)
			}
		})
	}
}

func TestBattalionTemplateCRLF(t *testing.T) {
	template, err := os.ReadFile("battalionNameConstruction.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := compileBattalionTokens(strings.ReplaceAll(string(template), "\n", "\r\n")); err != nil {
		t.Fatal(err)
	}
	if _, err := compileBattalionTokens("invalid"); err == nil {
		t.Fatal("invalid template accepted")
	}
}

func TestNamesHTMLEscapesNames(t *testing.T) {
	response := htmlSuccess([]NameResponse{{Id: "b1", Name: "<script>alert(1)</script> Infantry"}})
	if strings.Contains(response.Body, "<script>") || !strings.Contains(response.Body, "&lt;script&gt;") {
		t.Fatalf("name was not escaped: %s", response.Body)
	}
}
