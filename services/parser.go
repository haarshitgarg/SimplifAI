package services

import (
	"context"
	"fmt"
	"os"
	"strings"

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

// Handler function to be passed along with the recursive traverse function. This function type decides what operations are performed when you traverse the html
type traverseHandler func(parent *html.Node, child *html.Node)

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

	return htmlPage, nil
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

// Get the HTML of a given URL. Right now the browser instance is closed.
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

// Parse the html given a raw html file
func (p *webParser) parseHTML(rawHtml string) (string, error) {
	// Get the node pointer of the rawHTML
	node, err := html.Parse(strings.NewReader(rawHtml))
	if err != nil {
		return "", err
	}

	handler := func(parent *html.Node, child *html.Node) {
		if p.isActionable(child) {
			fmt.Printf("Type: %v\n", parent.Type)
			fmt.Printf("Type: %s\n", parent.Data)
			fmt.Printf("Attribute: %v\n\n", parent.Attr)
			parent.AppendChild(child)
		}
	}

	processedNode := p.traverseNode(node, handler)

	// We should have new node with only actionable elements so need to change the node to string now
	var buf strings.Builder
	html.Render(&buf, processedNode)

	return buf.String(), nil
}

// Traverse html node. Give in a handler to perform relevant things on these nodes
func (p *webParser) traverseNode(node *html.Node, handler traverseHandler) *html.Node {
	if node == nil {
		return nil
	}

	m := &html.Node{
		Type:     node.Type,
		DataAtom: node.DataAtom,
		Data:     node.Data,
		Attr:     make([]html.Attribute, len(node.Attr)),
	}
	copy(m.Attr, node.Attr)

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		// TODO: Do something with the handler here and perform whatever is required
		handler(m, p.traverseNode(c, handler))
	}

	return m
}

// Tells if element is actionable
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
		"meta":   true,
		"link":   true,
		"script": true,
		"style":  true,
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
