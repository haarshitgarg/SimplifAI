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

	info := p.traverseNode(node)
	info.Print(2)

	// Enhance elements with AI-friendly attributes
	p.addAIFriendlyAttributes(node)

	// We should have new node with only actionable elements so need to change the node to string now
	var buf strings.Builder
	html.Render(&buf, node)

	// Add simplified CSS for clean presentation
	enhancedHTML := p.addSimplifiedStyling(buf.String())

	err = os.WriteFile("logs/parsedHTML.html", []byte(enhancedHTML), 0o666)
	if err != nil {
		fmt.Printf("Unable to write the html data to file. Error: %s\n", err)
	}

	return enhancedHTML, nil
}

// Traverse html node and restructe based on parsing logic
func (p *webParser) traverseNode(node *html.Node) *LLMInfoNode {
	if node == nil {
		return nil
	}

	// TODO: Get all the info for a node to be given back to llm as input
	desc := fmt.Sprintf("Node type: %d, data: %s", node.Type, node.Data)
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
		"h1":     essential,
		"h2":     essential,
		"h3":     essential,
		"h4":     essential,
		"h5":     essential,
		"h6":     essential,
		"nav":    essential,
		"main":   essential,
		"header": essential,
		"footer": essential,
		"section": essential,
		"article": essential,
		"div":    essential,
		"p":      essential,
		"span":   essential,
		"label":  essential,
		"ul":     essential,
		"ol":     essential,
		"li":     essential,
		"table":  essential,
		"tr":     essential,
		"td":     essential,
		"th":     essential,

		"meta":   useless,
		"style":  useless,
		"script": useless,
		"noscript": useless,
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
	return ""
}

// Add AI-friendly attributes to HTML elements for better agent navigation
func (p *webParser) addAIFriendlyAttributes(node *html.Node) {
	if node == nil {
		return
	}

	// Add semantic labels to interactive elements
	if node.Type == html.ElementNode {
		switch node.Data {
		case "button":
			p.addAIAttribute(node, "data-ai-role", "button")
			p.addActionDescription(node)
		case "input":
			p.addAIAttribute(node, "data-ai-role", "input")
			p.addInputDescription(node)
		case "select":
			p.addAIAttribute(node, "data-ai-role", "dropdown")
			p.addSelectDescription(node)
		case "textarea":
			p.addAIAttribute(node, "data-ai-role", "text-area")
			p.addTextareaDescription(node)
		case "a":
			p.addAIAttribute(node, "data-ai-role", "link")
			p.addLinkDescription(node)
		case "form":
			p.addAIAttribute(node, "data-ai-role", "form")
			p.addFormDescription(node)
		case "nav":
			p.addAIAttribute(node, "data-ai-role", "navigation")
		case "header":
			p.addAIAttribute(node, "data-ai-role", "page-header")
		case "footer":
			p.addAIAttribute(node, "data-ai-role", "page-footer")
		case "main":
			p.addAIAttribute(node, "data-ai-role", "main-content")
		}

		// Add visual hierarchy indicators
		if strings.HasPrefix(node.Data, "h") && len(node.Data) == 2 {
			level := node.Data[1:]
			p.addAIAttribute(node, "data-ai-role", "heading")
			p.addAIAttribute(node, "data-ai-level", level)
		}
	}

	// Recursively process children
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		p.addAIFriendlyAttributes(child)
	}
}

// Helper function to add AI attribute to node
func (p *webParser) addAIAttribute(node *html.Node, key, value string) {
	node.Attr = append(node.Attr, html.Attribute{Key: key, Val: value})
}

// Add action description for buttons
func (p *webParser) addActionDescription(node *html.Node) {
	text := p.getElementText(node)
	if text == "" {
		text = p.getAttributeValue(node, "title")
	}
	if text == "" {
		text = p.getAttributeValue(node, "aria-label")
	}
	if text != "" {
		p.addAIAttribute(node, "data-ai-description", "Button: "+text)
	}
}

// Add description for input elements
func (p *webParser) addInputDescription(node *html.Node) {
	inputType := p.getAttributeValue(node, "type")
	placeholder := p.getAttributeValue(node, "placeholder")
	name := p.getAttributeValue(node, "name")
	
	desc := "Input field"
	if inputType != "" {
		desc = fmt.Sprintf("%s input", inputType)
	}
	if name != "" {
		desc += fmt.Sprintf(" for %s", name)
	}
	if placeholder != "" {
		desc += fmt.Sprintf(" (placeholder: %s)", placeholder)
	}
	
	p.addAIAttribute(node, "data-ai-description", desc)
}

// Add description for select elements
func (p *webParser) addSelectDescription(node *html.Node) {
	name := p.getAttributeValue(node, "name")
	desc := "Dropdown selection"
	if name != "" {
		desc += fmt.Sprintf(" for %s", name)
	}
	p.addAIAttribute(node, "data-ai-description", desc)
}

// Add description for textarea elements
func (p *webParser) addTextareaDescription(node *html.Node) {
	name := p.getAttributeValue(node, "name")
	placeholder := p.getAttributeValue(node, "placeholder")
	
	desc := "Text area"
	if name != "" {
		desc += fmt.Sprintf(" for %s", name)
	}
	if placeholder != "" {
		desc += fmt.Sprintf(" (placeholder: %s)", placeholder)
	}
	
	p.addAIAttribute(node, "data-ai-description", desc)
}

// Add description for link elements
func (p *webParser) addLinkDescription(node *html.Node) {
	text := p.getElementText(node)
	href := p.getAttributeValue(node, "href")
	
	desc := "Link"
	if text != "" {
		desc += fmt.Sprintf(": %s", text)
	}
	if href != "" && href != "#" {
		desc += fmt.Sprintf(" (goes to: %s)", href)
	}
	
	p.addAIAttribute(node, "data-ai-description", desc)
}

// Add description for form elements
func (p *webParser) addFormDescription(node *html.Node) {
	action := p.getAttributeValue(node, "action")
	method := p.getAttributeValue(node, "method")
	
	desc := "Form"
	if method != "" {
		desc += fmt.Sprintf(" (%s method)", strings.ToUpper(method))
	}
	if action != "" {
		desc += fmt.Sprintf(" submits to: %s", action)
	}
	
	p.addAIAttribute(node, "data-ai-description", desc)
}

// Get text content from element
func (p *webParser) getElementText(node *html.Node) string {
	if node.Type == html.TextNode {
		return strings.TrimSpace(node.Data)
	}
	
	var text string
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		text += p.getElementText(child)
	}
	
	return strings.TrimSpace(text)
}

// Get attribute value from node
func (p *webParser) getAttributeValue(node *html.Node, attrName string) string {
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}
	return ""
}

// Add simplified CSS styling for clean AI-agent presentation
func (p *webParser) addSimplifiedStyling(htmlContent string) string {
	aiStyleCSS := `
<style>
/* AI-Agent Friendly Styling */
body {
    font-family: Arial, sans-serif;
    line-height: 1.6;
    max-width: 1200px;
    margin: 0 auto;
    padding: 20px;
    background: #ffffff;
    color: #333333;
}

/* Clear visual hierarchy for headings */
h1, h2, h3, h4, h5, h6 {
    margin: 20px 0 10px 0;
    font-weight: bold;
}
h1 { font-size: 2em; color: #2c3e50; border-bottom: 3px solid #3498db; padding-bottom: 10px; }
h2 { font-size: 1.5em; color: #34495e; border-bottom: 2px solid #ecf0f1; padding-bottom: 5px; }
h3 { font-size: 1.3em; color: #34495e; }

/* Interactive elements with clear visual indicators */
[data-ai-role="button"], button {
    background: #3498db !important;
    color: white !important;
    border: none !important;
    padding: 12px 24px !important;
    margin: 8px 4px !important;
    border-radius: 6px !important;
    cursor: pointer !important;
    font-size: 14px !important;
    font-weight: bold !important;
    display: inline-block !important;
    text-decoration: none !important;
    transition: background 0.3s !important;
}

[data-ai-role="button"]:hover, button:hover {
    background: #2980b9 !important;
}

/* Form elements */
[data-ai-role="input"], input, [data-ai-role="text-area"], textarea, [data-ai-role="dropdown"], select {
    border: 2px solid #bdc3c7 !important;
    border-radius: 4px !important;
    padding: 10px !important;
    margin: 5px !important;
    font-size: 14px !important;
    background: #ffffff !important;
}

[data-ai-role="input"]:focus, input:focus, [data-ai-role="text-area"]:focus, textarea:focus, [data-ai-role="dropdown"]:focus, select:focus {
    border-color: #3498db !important;
    outline: none !important;
    box-shadow: 0 0 5px rgba(52, 152, 219, 0.3) !important;
}

/* Links */
[data-ai-role="link"], a {
    color: #3498db !important;
    text-decoration: underline !important;
    font-weight: 500 !important;
}

[data-ai-role="link"]:hover, a:hover {
    color: #2980b9 !important;
    background-color: #ecf0f1 !important;
    padding: 2px 4px !important;
    border-radius: 3px !important;
}

/* Navigation areas */
[data-ai-role="navigation"], nav {
    background: #ecf0f1 !important;
    padding: 15px !important;
    margin: 10px 0 !important;
    border-radius: 6px !important;
    border-left: 4px solid #3498db !important;
}

/* Forms */
[data-ai-role="form"], form {
    background: #f8f9fa !important;
    padding: 20px !important;
    margin: 15px 0 !important;
    border-radius: 8px !important;
    border: 1px solid #dee2e6 !important;
}

/* Lists */
ul, ol {
    padding-left: 20px !important;
}

li {
    margin: 5px 0 !important;
}

/* Tables */
table {
    border-collapse: collapse !important;
    width: 100% !important;
    margin: 15px 0 !important;
}

th, td {
    border: 1px solid #dee2e6 !important;
    padding: 12px !important;
    text-align: left !important;
}

th {
    background: #f8f9fa !important;
    font-weight: bold !important;
}

/* Page structure */
[data-ai-role="page-header"], header {
    background: #34495e !important;
    color: white !important;
    padding: 20px !important;
    margin-bottom: 20px !important;
    border-radius: 6px !important;
}

[data-ai-role="page-footer"], footer {
    background: #ecf0f1 !important;
    padding: 20px !important;
    margin-top: 20px !important;
    border-radius: 6px !important;
    text-align: center !important;
}

[data-ai-role="main-content"], main {
    background: white !important;
    padding: 20px !important;
    margin: 10px 0 !important;
}

/* Add visual indicators for AI attributes */
[data-ai-description]:before {
    content: "🤖 " attr(data-ai-description);
    display: block;
    font-size: 12px;
    color: #7f8c8d;
    margin-bottom: 5px;
    font-style: italic;
}
</style>`

	// Insert the CSS after the opening <head> tag or before closing </head>
	if strings.Contains(htmlContent, "<head>") {
		htmlContent = strings.Replace(htmlContent, "<head>", "<head>"+aiStyleCSS, 1)
	} else if strings.Contains(htmlContent, "</head>") {
		htmlContent = strings.Replace(htmlContent, "</head>", aiStyleCSS+"</head>", 1)
	} else {
		// If no head tag, add it
		htmlContent = strings.Replace(htmlContent, "<html", "<html><head>"+aiStyleCSS+"</head><html", 1)
	}

	return htmlContent
}

