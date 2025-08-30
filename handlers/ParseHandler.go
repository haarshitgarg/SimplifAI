package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/haarshitgarg/SimplifAI/services"
)

type ParseReq struct {
	URL string `json:"url"`
}

type ParseHandler struct {
	Service services.WebParser
}

func NewParseHandler(s services.WebParser) *ParseHandler {
	return &ParseHandler{Service: s}
}

// Parse handler to parse http body to a more llm friendly format
func (h *ParseHandler) Parse(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Calling Parse in ParseHandler.go")
	// TODO: get the exact url from the body. So need to make sure the header is application/json
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(400)
		w.Write([]byte("The request content type was bad"))
		return
	}
	var req ParseReq
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&req)
	if err != nil || req.URL == "" {
		w.WriteHeader(400)
		w.Write([]byte("The request does not have a url body in json format"))
		return
	}

	res, err := h.Service.Parse(req.URL)

	if err != nil {
		w.WriteHeader(200)
		w.Write([]byte("Parse error"))
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(res))
}

// Test Handler to test the html parsing capability.
func (h *ParseHandler) TestParser(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("TestParser handler function called\n")
	filename := chi.URLParam(r, "filename")
	parsedHTML, err := h.Service.ParseHTML(filename)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	fmt.Println(parsedHTML)
	w.WriteHeader(200)
	w.Write([]byte(parsedHTML))
}

