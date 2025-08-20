package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/haarshitgarg/SimplifAI/services"
)

// Setup a ParseHandler struct
type ParseHandler struct {
	Service services.WebParser
}

func NewParseHandler(s services.WebParser) *ParseHandler {
	return &ParseHandler{Service: s}
}

func (h *ParseHandler) Parse(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request reached the Parse handler")
	body := "This is a temporary Response"

	nres, err := h.Service.Parse(&body)

	if err != nil {
		return
	}

	log.Println(nres)
}
