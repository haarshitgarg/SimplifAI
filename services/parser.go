package services

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/chromedp/chromedp"
	"golang.org/x/net/html"
)

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

/////////////////////////////////////////////
///////////// Types for parser //////////////
/////////////////////////////////////////////

type assetType string

const(
	essential assetType = "essential"
	actionable assetType = "actionable"
	useless assetType = "useless"
	textinfo assetType = "textinfo"
	reparent assetType = "reparent"
)

/////////////////////////////////////////////
////// Exported functions for handlers //////
/////////////////////////////////////////////

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

	// Temporary
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

/////////////////////////////////////////////
////// Internal support functions ///////////
/////////////////////////////////////////////

// Parse the html given a raw html file
func (p *webParser) parseHTML(rawHtml string) (string, error) {
	// Get the node pointer of the rawHTML
	node, err := html.Parse(strings.NewReader(rawHtml))
	if err != nil {
		return "", err
	}

	_ = p.traverseNode(node)
	//info.Print(2)

	// We should have new node with only actionable elements so need to change the node to string now
	var buf strings.Builder
	html.Render(&buf, node)

	err = os.WriteFile("logs/parsedHTML.html", []byte(buf.String()), 0o666)
	if err != nil {
		fmt.Printf("Unable to write the html data to file. Error: %s\n", err)
	}

	return buf.String(), nil
}

// Traverse html node and restructe based on parsing logic
func (p *webParser) traverseNode(node *html.Node) *LLMInfoNode {
	if node == nil {
		return nil
	}

	if node.Type == html.ElementNode && node.Data == "body" {
		fmt.Println(GetDescription(node))
	}

	// TODO: Get all the info for a node to be given back to llm as input
	desc := GetDescription(node)
	info := NewLLMInfoNode(p.getSelectorID(node))
	info.UpdateElementDesc(desc)

	for c := node.FirstChild; c != nil; {
		next := c.NextSibling

		switch p.isUseful(c) {
		case essential, actionable, textinfo:
			// TODO Maybe edit the description of the current info node with this essential node
			childInfo := p.traverseNode(c)
			info.AppendChild(childInfo)
		case useless:
			node.RemoveChild(c)
		case reparent:
			p.reparentChildren(node, c)
		}

		c = next
	}

	return info
}

// Tells if element is is useful and returns the category, so you can ignore the complete node or keep it etc.
func (p *webParser) isUseful(node *html.Node) assetType {
	if node.Type != html.ElementNode {
		if node.Type != html.TextNode {
			return useless // Only elements can be actionable
		} else {
			return textinfo
		}
	}

	// Check if it's a naturally actionable tag
	actionableTags := map[string]assetType{
		"button":   actionable,
		"input":    actionable,
		"select":   actionable,
		"textarea": actionable,
		"a":        actionable,
		"form":     actionable,

		"html":   essential,
		"head":   essential,
		"body":   essential,
		"title":  essential,
		"link":   essential,
		"script": essential,

		"meta":   useless,
		"style":  useless,
	}

	if val, ok := actionableTags[node.Data]; ok {
		return val 
	}

	// Check for JavaScript events or other interactive attributes
	for _, attr := range node.Attr {
		if strings.HasPrefix(attr.Key, "on") || attr.Key == "href" {
			return actionable
		}
	}

	return essential 
}

// reparentChildren reparents all of src's child nodes to dst.
func (p *webParser) reparentChildren(dst, src *html.Node) {
	for {
		child := src.FirstChild
		if child == nil {
			break
		}
		src.RemoveChild(child)
		dst.AppendChild(child)
	}
}

// Get seleactor id based on a node
func (p *webParser) getSelectorID(node *html.Node) string {
	// TODO: Add more info on how to get an accurate selector id
	return ""
}

