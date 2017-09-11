package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	// Hard-coded port since it's defined in the Salesforce connected-app as part of the
	// callback URL
	port := 10000

	r := mux.NewRouter()

	// Authentication handlers
	r.HandleFunc("/salesforce/login", indexPageHandler)
	r.HandleFunc("/salesforce/oauth/receive_token", receiveOAuthTokenHandler)

	// API Definitions
	r.HandleFunc("/api/v1/salesforce/reports/run/{id}", runReportAPI).Methods("GET")
	r.HandleFunc("/api/v1/salesforce/accounts", getAccountsAPI).Methods("GET")

	loggedRouter := handlers.LoggingHandler(os.Stdout, r)

	log.Printf("Starting server on :%d", port)
	srv := &http.Server{
		Handler:      loggedRouter,
		Addr:         fmt.Sprintf(":%d", port),
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
