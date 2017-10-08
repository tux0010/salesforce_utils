package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type NewToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    string `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	State        string `json:"state"`
	InstanceURL  string `json:"instance_url"`
	ID           string `json:"id"`
	IssuedAt     string `json:"issued_at"`
	Signature    string `json:"signature"`
}

type RefreshToken struct {
	AccessToken string `json:"access_token"`
	InstanceURL string `json:"instance_url"`
	IssuedAt    string `json:"issued_at"`
	ID          string `json:"id"`
	Signature   string `json:"signature"`
}

type Salesforce struct {
	consumerKey    string
	consumerSecret string
	redirectURL    string
	loginBaseURL   string
	refreshToken   string
}

func NewSalesforce(consumerKey, consumerSecret, redirectURL, loginBaseURL, refreshToken string) (*Salesforce, error) {
	if consumerKey == "" {
		return nil, errors.New("Empty consumer key")
	}

	if consumerSecret == "" {
		return nil, errors.New("Empty consumer key")
	}

	if redirectURL == "" {
		return nil, errors.New("Empty redirect URL")
	}

	if loginBaseURL == "" {
		return nil, errors.New("Empty login base URL")
	}

	return &Salesforce{
		consumerKey:    consumerKey,
		consumerSecret: consumerSecret,
		redirectURL:    redirectURL,
		loginBaseURL:   loginBaseURL,
		refreshToken:   refreshToken,
	}, nil
}

func (s *Salesforce) loginHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	v := url.Values{}
	v.Set("response_type", "token")
	v.Set("client_id", s.consumerKey)
	v.Set("redirect_uri", s.redirectURL)
	v.Add("scope", "api refresh_token")
	loginURL := fmt.Sprintf("%s/authorize?%s", s.loginBaseURL, v.Encode())

	log.WithFields(log.Fields{"url": loginURL}).Println("Re-directing to login page")
	http.Redirect(w, r, loginURL, 301)
}

func (s *Salesforce) receiveTokenHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// IMPORTANT: Since Salesforce returns a fragment URL, the "#..." is stripped the browser. We will use
	// Javascript to obtain this for us and will POST with "#" converted to "?" so that it can be seen and
	// parsed properly as a query URL
	log.WithFields(log.Fields{"url": r.URL}).Println("Received response from Salesforce with token info")

	t, err := template.ParseFiles(fmt.Sprintf("templates%sreceive_oauth_token.html", string(filepath.Separator)))
	if err != nil {
		log.WithError(err).Println("Unable to parse template")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	t.Execute(w, struct{ CallbackURL string }{"/token/parse"})
}

func (s *Salesforce) parseTokenHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	log.WithFields(log.Fields{"url": r.URL}).Println("Received response from JS with token fragment data")

	v := r.URL.Query()
	token := &NewToken{
		AccessToken:  v.Get("access_token"),
		ExpiresIn:    v.Get("expires_in"),
		RefreshToken: v.Get("refresh_token"),
		State:        v.Get("state"),
		InstanceURL:  v.Get("instance_url"),
		ID:           v.Get("id"),
		IssuedAt:     v.Get("issued_at"),
		Signature:    v.Get("signature"),
	}

	encoder := json.NewEncoder(w)
	err := encoder.Encode(token)
	if err != nil {
		log.WithError(err).Println("Unable to render JSON")
		http.Error(w, fmt.Sprintf("Unable to render token JSON: %s", err.Error()), http.StatusInternalServerError)
	}
}

func (s *Salesforce) refreshTokenHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if s.refreshToken == "" {
		http.Error(w, "Missing refresh token", http.StatusForbidden)
		return
	}

	log.WithFields(log.Fields{"refresh_token": s.refreshToken}).Println("Requesting new token")

	v := url.Values{}
	v.Set("grant_type", "refresh_token")
	v.Set("client_id", s.consumerKey)
	v.Set("client_secret", s.consumerSecret)
	v.Set("refresh_token", s.refreshToken)

	refreshURL := fmt.Sprintf("%s/token?%s", s.loginBaseURL, v.Encode())
	log.WithFields(log.Fields{"url": refreshURL}).Println("Requesting refresh token")

	resp, err := http.Post(refreshURL, "", nil)
	if err != nil {
		log.WithError(err).Println("Error from Salesforce API")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var sr RefreshToken

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&sr)
	if err != nil {
		log.WithError(err).Println("Error decoding JSON from Salesforce API response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(sr)
}
