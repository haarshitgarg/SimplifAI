package services

import (
	"fmt"
	"strings"
)

type LLMInfoNode struct {
	parent, firstchild, lastchild, nextSibling, prevSibling *LLMInfoNode

	selectorID string
	elemDesc string
}

// Create a new LLM info node based on the selector id 
func NewLLMInfoNode(selectorID string) *LLMInfoNode {
	return &LLMInfoNode{
		parent: nil, 
		firstchild: nil, 
		lastchild: nil,
		nextSibling: nil,
		prevSibling: nil,

		selectorID: selectorID,
		elemDesc: "",
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
