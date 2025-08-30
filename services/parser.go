package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"golang.org/x/net/html"
)

// TODO: Learn how to implement the business logic of writing a parser

// The reason we have a WebParser interface is because we can swap the constructor to return mock web parser if we require it for testing
type WebParser interface {
	// Given a URL it parses the html and then returns a refined html file
	Parse(url string) (string, error)
	// Given a html body at `filepath` it parses it and then returns a refined html file
	ParseHTML(filepath string) (string, error)
}

// The is the actual web parser
type webParser struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// Returns a new Web parser
func NewWebParser() WebParser {
	ctx, cancel := chromedp.NewContext(context.Background())
	return &webParser{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Parse the input stream of http data and return a nice formated string useful for llm
func (p *webParser) Parse(rawURL string) (string, error) {
	fmt.Println("Calling Parse in parser.go service")
	htmlPage, err := p.getHTML(rawURL)
	if err != nil {
		return "", nil
	}
	parsedHTML, err := p.parseHTML(htmlPage)
	if err != nil {
		return "", nil
	}

	go p.openTabWithHtml(parsedHTML)

	return parsedHTML, nil
}

// Parse the HTML given the filepath of the html file
func (p *webParser) ParseHTML(filepath string) (string, error) {
	fmt.Printf("Calling ParseHTML to parse HTML for filepath: %s\n", filepath)
	filepath = fmt.Sprintf("htmls/%s", filepath)
	buf, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	// Internal function to parse any html file
	return p.parseHTML(string(buf))
}

// Get the HTML of a given URL
func (p *webParser) getHTML(rawURL string) (string, error) {
	fmt.Printf("getHTML function called to get html for %s\n", rawURL)

	var html_str string
	err := chromedp.Run(
		p.ctx,
		chromedp.Navigate(rawURL),
		chromedp.WaitReady("body"),
		chromedp.OuterHTML("html", &html_str),
	)
	if err != nil {
		return "", fmt.Errorf("Could not find the html body for url: %s\nError: %s\n", rawURL, err)
	}

	return html_str, nil
}

// Open a new tab with a given html
func (p *webParser) openTabWithHtml(rawHTML string) error {
	fmt.Println("Called openTabWithHtml to open html file in new tab")
	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("start-maximized", true),
	)
	allocCtx, _ := chromedp.NewExecAllocator(p.ctx, opts...)
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	encoded := base64.StdEncoding.EncodeToString([]byte(rawHTML))
	dataURL := "data:text/html;base64,"+encoded
	chromedp.Run(
		ctx,
		chromedp.Navigate(dataURL),
		chromedp.Sleep(50*time.Second),
	)

	return nil
}

// Parse the html given a raw html file
func (p *webParser) parseHTML(rawHtml string) (string, error) {
	// Get the node pointer of the rawHTML
	node, err := html.Parse(strings.NewReader(rawHtml))
	if err != nil {
		return "", err
	}

	p.traverseNode(node)

	// We should have new node with only actionable elements so need to change the node to string now
	var buf strings.Builder
	html.Render(&buf, node)

	return buf.String(), nil
}

// Traverse html node and restructe based on parsing logic
func (p *webParser) traverseNode(node *html.Node) {
	if node == nil {
		return
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if !p.isActionable(c) {
			reparentChildren(node, c)
		} else {
			p.traverseNode(c)
		}
	}
}

// Tells if element is actionable or essential for the html
func (p *webParser) isActionable(node *html.Node) bool {
	if node.Type != html.ElementNode {
		return false // Only elements can be actionable
	}

	// Check if it's a naturally actionable tag
	actionableTags := map[string]bool{
		"button":   true,
		"input":    true,
		"select":   true,
		"textarea": true,
		"a":        true,
		"form":     true,
	}

	if actionableTags[node.Data] {
		return true
	}

	// Keep the essensial tags. TODO: review these tags
	essentialTags := map[string]bool{
		"html":   true,
		"head":   true,
		"body":   true,
		"title":  true,
		"meta":   false,
		"link":   true,
		"script": true,
		"style":  false,
	}

	if essentialTags[node.Data] {
		return true
	}

	// Check for JavaScript events or other interactive attributes
	for _, attr := range node.Attr {
		if strings.HasPrefix(attr.Key, "on") || attr.Key == "href" {
			return true
		}
	}

	return false
}

// reparentChildren reparents all of src's child nodes to dst.
func reparentChildren(dst, src *html.Node) {
	for {
		child := src.FirstChild
		if child == nil {
			break
		}
		src.RemoveChild(child)
		dst.AppendChild(child)
	}
}

