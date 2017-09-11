package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func renderJSON(data interface{}, w http.ResponseWriter) error {
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(&data); err != nil {
		log.Printf("Error encoding JSON: %s", err.Error())
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return err
	}

	return nil
}

func callSalesforceAPI(reqURL string, w http.ResponseWriter) (*http.Response, error) {
	u := sfToken.InstanceURL + "/services/data/v37.0" + reqURL
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Add("Authorization", "Bearer "+sfToken.AccessToken)

	client := &http.Client{Timeout: time.Second * 30}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error calling the Salesforce API endpoint: %s", err.Error())
		http.Error(w, "Error calling the Salesforce API endpoint", http.StatusInternalServerError)
		return resp, err
	}

	return resp, nil
}
