package services

import (
	"bytes"
	"context"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/haarshitgarg/SimplifAI/types"
	"golang.org/x/net/html"
)

// TODO: Learn how to implement the business logic of writing a parser

// The reason we have a WebParser interface is because we can swap the constructor to return mock web parser if we require it for testing
type WebParser interface {
	Parse(r *string) (string, error)
	GetHTML(url string) (string, error)
}

// This is the actual web parser
type webParser struct{}

func NewWebParser() WebParser {
	return &webParser{}
}

func (p *webParser) Parse(r *string) (string, error) {
	log.Println("Parsing a http response to a llm friendly response")
	log.Printf("Received the following to parse: %s", *r)

	return *r + " processed", nil
}

func (p *webParser) GetHTML(url string) (string, error) {
	log.Println("CreateHeadlessChromeInstance function called")

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var html_data string
	var res []byte

	err := chromedp.Run(
		ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.Sleep(1*time.Second),
		chromedp.CaptureScreenshot(&res),
		chromedp.OuterHTML("html", &html_data),
	)

	if err != nil {
		log.Printf("Could not navigate to the url. Error: %s", err)
		return "Could not load the html", err
	}

	// Store the screenshot
	err = os.WriteFile("screenshots/Screenshot.png", res, 0o644)
	if err != nil {
		log.Printf("Could not write the screenshot")
		return "", err
	}

	doc, err := html.Parse(bytes.NewReader([]byte(html_data)))
	constructElementsList(doc)

	return html_data, nil
}

func constructElementsList(doc *html.Node) []types.ElementInfo {
	log.Println("constructElementsList function called")
	var acElements []types.ElementInfo
	log.Printf("Printing the html node data %s", doc.Type)

	return acElements
}
