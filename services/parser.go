package services

import (
	"log"
)

// TODO: Learn how to implement the business logic of writing a parser

// The reason we have a WebParser interface is because we can swap the constructor to return mock web parser if we require it for testing
type WebParser interface {
	Parse(r *string) (string, error)
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
