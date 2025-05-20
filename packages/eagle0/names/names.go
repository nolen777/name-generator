package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nolen777/name-generator/packages/eagle0/names/parser"
	"github.com/nolen777/name-generator/packages/eagle0/names/spaces_fetcher"
	"github.com/nolen777/name-generator/packages/eagle0/names/token"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type headers struct {
	Accept string `json:"accept"`
}

type httpInfo struct {
	Headers headers `json:"headers"`
	Method  string  `json:"method"`
	Path    string  `json:"path"`
}

type NameRequest struct {
	Id     string `json:"id"`
	Gender string `json:"gender"`
}

type Event struct {
	Requests []NameRequest `json:"requests"`
	Http     httpInfo      `json:"http"`
}

type ResponseHeaders struct {
	ContentType string `json:"Content-Type"`
}

type NameResponse struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Response struct {
	Body       string          `json:"body"`
	StatusCode string          `json:"statusCode"`
	Headers    ResponseHeaders `json:"headers"`
}

var femaleCtx token.StringConstructionContext
var maleCtx token.StringConstructionContext
var otherCtx token.StringConstructionContext

var stringConstructionToken token.StringConstructionToken

func init() {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		femaleCtx, maleCtx, otherCtx = generateContexts()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		stringConstructionString := fetchStringConstructionToken()
		tok, err := parser.ParseFrom(stringConstructionString)
		if err != nil {
			fmt.Println("Error parsing string construction token: ", err)
			panic(err)
		}
		stringConstructionToken = tok
	}()

	wg.Wait()
}

func Names(ctx context.Context, event Event) Response {
	info := event.Http
	headers := info.Headers

	rGen := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Get the requests
	requests := generateRequests(event, rGen)

	nameResponses := []NameResponse{}
	for _, request := range requests {
		scCtx := otherCtx
		if request.Gender == "female" {
			scCtx = femaleCtx
		} else if request.Gender == "male" {
			scCtx = maleCtx
		}
		name, err := stringConstructionToken.Next(rGen, scCtx)
		if err != nil {
			fmt.Println("Error generating name: ", err)
			return Response{
				Body:       "<html><h1>Error generating name</h1></html>",
				StatusCode: "500",
				Headers: ResponseHeaders{
					ContentType: "text/html",
				},
			}
		}
		nameResponses = append(nameResponses, NameResponse{
			Id:   request.Id,
			Name: name,
		})
	}

	if headers.Accept == "application/json" {
		fmt.Println("returning json")
		return jsonSuccess(nameResponses)
	}
	if headers.Accept == "text/html" {
		fmt.Println("returning html")
	}
	return htmlSuccess(nameResponses)
}

type jsonBody struct {
	Names []NameResponse `json:"names"`
}

func jsonSuccess(nameResponses []NameResponse) Response {
	bodyObj, err := json.Marshal(jsonBody{Names: nameResponses})
	if err != nil {
		fmt.Println("Error marshalling JSON: ", err)
		return Response{
			Body:       "<html><h1>Error marshalling JSON</h1></html>",
			StatusCode: "500",
			Headers: ResponseHeaders{
				ContentType: "text/html",
			},
		}
	}
	return Response{
		Body:       string(bodyObj),
		StatusCode: "200",
		Headers: ResponseHeaders{
			ContentType: "application/json",
		},
	}
}

func htmlSuccess(nameResponses []NameResponse) Response {
	names := make([]string, len(nameResponses))
	for i, nameResponse := range nameResponses {
		names[i] = nameResponse.Name
	}
	return Response{
		Body:       "<html>\r\n" + strings.Join(names, "\r\n<p>\r\n") + "</html>\r\n",
		StatusCode: "200",
		Headers: ResponseHeaders{
			ContentType: "text/html",
		},
	}
}

func generateRequests(event Event, rGen *rand.Rand) []NameRequest {
	requests := event.Requests
	if len(requests) == 0 {
		fmt.Println("No requests found")
		for i := 0; i < 20; i++ {
			roll := rGen.Float64()
			gender := "other"
			if roll < 0.4 {
				gender = "female"
			} else if roll < 0.8 {
				gender = "male"
			}
			requests = append(requests, NameRequest{
				Id:     fmt.Sprintf("%d", i),
				Gender: gender,
			})
		}
	}
	return requests
}

func generateContexts() (token.StringConstructionContext, token.StringConstructionContext, token.StringConstructionContext) {
	femaleNameWords := map[string][]string{}
	maleNameWords := map[string][]string{}
	unfilteredNameWords := map[string][]string{}

	namesTsvBytes, err := spaces_fetcher.GetFile("names.tsv")
	if err != nil {
		panic(err)
	}
	namesTsv := string(namesTsvBytes)

	nameLines := strings.Split(namesTsv, "\r\n")
	nameTitles := strings.Split(nameLines[0], "\t")
	titleBuckets := make([]string, len(nameTitles))
	for i := range nameTitles {
		components := strings.Split(nameTitles[i], "@")
		nameTitles[i] = components[0]
		titleBuckets[i] = ""
		if len(components) > 1 {
			titleBuckets[i] = components[1]
		}
		maleNameWords[nameTitles[i]] = []string{}
		femaleNameWords[nameTitles[i]] = []string{}
		unfilteredNameWords[nameTitles[i]] = []string{}
	}
	for _, line := range nameLines[1:] {
		for i, entry := range strings.Split(line, "\t") {
			if entry == "" {
				continue
			}
			title := nameTitles[i]
			bucket := titleBuckets[i]

			unfilteredNameWords[title] = append(unfilteredNameWords[title], entry)
			switch bucket {
			case "female":
				femaleNameWords[title] = append(femaleNameWords[title], entry)
				break
			case "male":
				maleNameWords[title] = append(maleNameWords[title], entry)
				break
			default:
				femaleNameWords[title] = append(femaleNameWords[title], entry)
				maleNameWords[title] = append(maleNameWords[title], entry)
				break
			}
		}
	}
	maleCtx := token.StringConstructionContext{
		ChoiceListMap:           maleNameWords,
		UnfilteredChoiceListMap: unfilteredNameWords,
	}
	femaleCtx := token.StringConstructionContext{
		ChoiceListMap:           femaleNameWords,
		UnfilteredChoiceListMap: unfilteredNameWords,
	}
	otherCtx := token.StringConstructionContext{
		ChoiceListMap:           unfilteredNameWords,
		UnfilteredChoiceListMap: unfilteredNameWords,
	}

	return femaleCtx, maleCtx, otherCtx
}

func fetchStringConstructionToken() string {
	rawStringConstructionToken, err := spaces_fetcher.GetFile("nameConstruction.txt")
	if err != nil {
		panic(err)
	}
	removeCRs := strings.ReplaceAll(string(rawStringConstructionToken), "\r", "")
	removeNLs := strings.ReplaceAll(removeCRs, "\n", "")
	return removeNLs
}
