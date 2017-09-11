package main

var sfToken *SalesforceToken

// SalesforceToken contains the data returned from the callback upon
// successful authentication
type SalesforceToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    string `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	State        string `json:"state"`
	InstanceURL  string `json:"instance_url"`
	ID           string `json:"id"`
	IssuedAt     string `json:"issued_at"`
	Signature    string `json:"signature"`
}
