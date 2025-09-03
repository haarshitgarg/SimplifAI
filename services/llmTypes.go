package services

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

type LLMInfoNode struct {
	parent, firstchild, lastchild, nextSibling, prevSibling *LLMInfoNode

	selectorID string
	elemDesc   string
}

// Create a new LLM info node based on the selector id
func NewLLMInfoNode(selectorID string) *LLMInfoNode {
	return &LLMInfoNode{
		parent:      nil,
		firstchild:  nil,
		lastchild:   nil,
		nextSibling: nil,
		prevSibling: nil,

		selectorID: selectorID,
		elemDesc:   "",
	}
}

// Update the selector id for a node
func (n *LLMInfoNode) UpdateSelectorID(id string) {
	n.selectorID = id
}

// Update the node element description
func (n *LLMInfoNode) UpdateElementDesc(desc string) {
	// TODO: Make it more robust. Maybe add more functions to decide how to chnage this description
	n.elemDesc = n.elemDesc + " " + desc
}

// Append child to a node
func (n *LLMInfoNode) AppendChild(child *LLMInfoNode) {
	if n.firstchild == nil {
		n.firstchild = child
		n.lastchild = child
		child.parent = n
	} else {
		child.parent = n
		child.prevSibling = n.lastchild
		n.lastchild.nextSibling = child
		n.lastchild = child
	}
}

// Print recursively prints the node and its children in a tree format
func (n *LLMInfoNode) Print(indent int) {
	if n == nil {
		return
	}

	// Indentation for tree visualization
	prefix := strings.Repeat("  ", indent)

	// Print this node
	fmt.Printf("%s- selectorID: %s, elemDesc: %s\n", prefix, n.selectorID, n.elemDesc)

	// Print children recursively
	for child := n.firstchild; child != nil; child = child.nextSibling {
		child.Print(indent + 1)
	}
}

// Gets the description of an HTML node
// TODO: I need to get the description based on the parent into consideration
func GetDescription(node *html.Node) string {
	if node == nil {
		return ""
	}

	if node.Type == html.TextNode {
		return strings.TrimSpace(node.Data)
	}

	desc := ""
	switch node.Data {
	case "body":
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			childDesc := GetDescription(c)
			if childDesc != "" {
				desc += childDesc + "\n\n"
			}
		}
	case "nav":
		desc = "We have the following navigation links:\n"
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			childDesc := GetDescription(c)
			if childDesc != "" {
				desc += "- " + " " + childDesc + "\n" // Get list of links
			}
		}
	case "a":
		// Check the attributes here. Right now just the href but maybe target and other important stuff
		for _, attr := range node.Attr {
			if attr.Key == "href" {
				desc = "Link: " + attr.Val
				break
			}
		}

		childDesc := ""
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			nDec := GetDescription(c)
			if nDec != "" {
				childDesc += nDec
			}
		}
		desc += ", Description: " + childDesc

	case "i", "p", "h1", "h2", "h3":
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			childDesc := GetDescription(c)
			if childDesc != "" {
				desc += " " + childDesc
			}
		}

	case "div", "main":
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			childDec := GetDescription(c)
			if childDec != "" {
				desc += childDec + "\n";
			}
		}

	default:

	}

	return desc
}
