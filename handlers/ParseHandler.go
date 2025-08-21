package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/haarshitgarg/SimplifAI/services"
)

// Setup a ParseHandler struct
type ParseHandler struct {
	Service services.WebParser
}

// request body format for GetHTML handler
type getHTMLRequest struct {
	URL string `json:"url"`
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

func (h *ParseHandler) GetHTML(w http.ResponseWriter, r *http.Request) {
	log.Println("GetHTML handler function")
	body := r.Body

	var url getHTMLRequest
	if json.NewDecoder(body).Decode(&url) != nil {
		log.Println("Could not decode the body for GetHTML handler")
		return
	}

	_, err := h.Service.GetHTML(url.URL)
	if err != nil {
		log.Println(err)
		return
	}

	//log.Println(html)
}

func (h *ParseHandler) ShowPage(w http.ResponseWriter, r *http.Request) {
	htmlpage, err := os.ReadFile("html/generated.html")
	if err != nil {
		log.Println("Could not read the generated.html")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(htmlpage)
}

