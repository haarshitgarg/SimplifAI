package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/haarshitgarg/SimplifAI/handlers"
	"github.com/haarshitgarg/SimplifAI/services"
)

func main() {
	fmt.Println("Starting the SimplifAI Server...")
	r := chi.NewRouter()

	parseServices := services.NewWebParser()
	parseHandlers := handlers.NewParseHandler(parseServices)

	r.Get("/parse", parseHandlers.Parse)

	fmt.Println("Starting server on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

