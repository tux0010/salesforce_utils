package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/gorilla/mux"
)

func receiveOAuthTokenHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// IMPORTANT: Since Salesforce returns a fragment URL, the "#..." is stripped
		// by the browser. We will use Javascript to obtain this for us and will POST
		// to the same URL with "#" converted to "?" so that it can be seen and parsed
		// properly as a query URL
		log.Println("Received response from Salesforce with token info")

		t, err := template.ParseFiles(fmt.Sprintf("templates%sreceive_oauth_token.html", string(filepath.Separator)))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		log.Println("Rendering form to callback with fragment URL token info via POST")
		t.Execute(w, nil)
	case "POST":
		log.Println("Received response from Javascript POST with token info")

		v := r.URL.Query()
		token := &SalesforceToken{
			AccessToken:  v.Get("access_token"),
			ExpiresIn:    v.Get("expires_in"),
			RefreshToken: v.Get("refresh_token"),
			State:        v.Get("state"),
			InstanceURL:  v.Get("instance_url"),
			ID:           v.Get("id"),
			IssuedAt:     v.Get("issued_at"),
			Signature:    v.Get("signature"),
		}

		// TODO: Store token in Database
		sfToken = token

		// Render as JSON for now
		encoder := json.NewEncoder(w)
		err := encoder.Encode(sfToken)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to render token JSON: %s", err.Error()), http.StatusInternalServerError)
		}
	}
}

func indexPageHandler(w http.ResponseWriter, r *http.Request) {
	// Using the user-agent flow, so that we can request a refresh token (automatically renew)
	// See the documentation on the Salesforce Rest API page for more information
	consumerKey := ""
	callbackURL := "http://localhost:10000/salesforce/oauth/receive_token"
	//loginURL := "https://login.salesforce.com/services/oauth2/authorize?"
	loginURL := "https://test.salesforce.com/services/oauth2/authorize?"

	v := url.Values{}
	v.Set("response_type", "token")
	v.Set("client_id", consumerKey)
	v.Set("redirect_uri", callbackURL)
	v.Add("scope", "api refresh_token")
	loginURL += v.Encode()

	log.Printf("Redirecting to the Salesforce Login page: %s", loginURL)
	http.Redirect(w, r, loginURL, 301)
}

func runReportAPI(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	url := "/analytics/reports/" + id + "?includeDetails=true"
	resp, err := callSalesforceAPI(url, w)
	if err != nil {
		return
	}

	var js interface{}
	defer r.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&js)
	if err != nil {
		log.Printf("Unable to decode JSON from body: %s", err.Error())
		http.Error(w, "Unable to decode JSON from body", 422)
		return
	}

	renderJSON(js, w)
}

func getAccountsAPI(w http.ResponseWriter, r *http.Request) {
	v := url.Values{}
	v.Set("q", "SELECT name from Account")
	queryURL := "/query/?" + v.Encode()

	resp, err := callSalesforceAPI(queryURL, w)
	if err != nil {
		return
	}

	var js interface{}
	defer r.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&js)
	if err != nil {
		log.Printf("Unable to decode JSON from body: %s", err.Error())
		http.Error(w, "Unable to decode JSON from body", 422)
		return
	}

	renderJSON(js, w)
}
