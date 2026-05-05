package models

import "net/url"

type OAuthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func (r OAuthRequest) ToValues() url.Values {
	form := url.Values{}
	form.Set("client_id", r.ClientID)
	form.Set("client_secret", r.ClientSecret)
	form.Set("grant_type", r.GrantType)
	form.Set("scope", r.Scope)
	return form
}

type OAuthRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
	Scope        string `json:"scope"`
}
