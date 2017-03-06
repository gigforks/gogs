package authorization

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gogits/gogs/modules/log"
	"github.com/gogits/gogs/modules/setting"
)

type iyoAuthentication struct {
	url          string
	clientId     string
	clientSecret string
	at           string
}

const iyoURL = "https://itsyou.online"

var auth *iyoAuthentication

// UserIsMemberOfOrg makes an api call to itsyou.online to verify if the user has
// membership access to the organization
func UserIsMemberOfOrg(username string, globalId string) bool {
	if auth == nil {
		loadAuth()
	}
	endpoint := fmt.Sprintf("/api/organizations/%v/users/ismember/%v", globalId, username)

	hc := &http.Client{}
	req, err := http.NewRequest("GET", auth.url+endpoint, nil)

	if err != nil {
		log.Warn("Failed to create request: ", err)
		return false
	}

	resp, err := doRequest(hc, req, true)
	if err != nil || resp == nil {
		log.Warn("Failed to verify if user is a member of org: ", err)
		return false
	}

	response := &struct {
		IsMember bool
	}{}

	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		log.Warn("Failed to read the ismember response: ", err)
		return false
	}

	return response.IsMember
}

// doRequest performs the request with the provided client. If setAuthHeader is true,
// an authentication header gets set. The status code of the response is inspected and
// in case it is a StatusUnauthorized(statuscode = 401), the access token is refreshed,
// the new token is set on the authorization header, and the request is retried.
// Actuall errors must be handled by the caller
func doRequest(hc *http.Client, r *http.Request, setAuthHeader bool) (*http.Response, error) {
	if setAuthHeader {
		r.Header.Set("Authorization", "token "+auth.at)
	}
	resp, err := hc.Do(r)
	if err != nil {
		return nil, err
	}

	// check if the token was expired
	if resp.StatusCode != http.StatusUnauthorized {
		return resp, err
	}

	log.Debug("Retry request")

	// refresh token and try again
	getAccessToken()
	// set the new token in the header
	r.Header.Set("Authorization", "token "+auth.at)
	return hc.Do(r)
}

// getAccessToken makes an API call to Itsyou.Online to pick up an access token
func getAccessToken() {
	if auth == nil {
		loadAuth()
	}

	endpoint := "/v1/oauth/access_token"

	hc := &http.Client{}
	req, err := http.NewRequest("POST", auth.url+endpoint, nil)
	if err != nil {
		return
	}

	q := req.URL.Query()
	q.Add("client_id", auth.clientId)
	q.Add("client_secret", auth.clientSecret)
	q.Add("grant_type", "client_credentials")
	req.URL.RawQuery = q.Encode()

	resp, err := hc.Do(req)
	if err != nil {
		log.Warn("Error while getting the access token: ", err)
		return
	}
	defer resp.Body.Close()

	response := &struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		Scope       string      `json:"scope"`
		ExpiresIn   int64       `json:"expires_in"`
		Info        interface{} `json:"info"`
	}{}

	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		log.Warn("Failed to get access_token: ", err)
	}

	auth.at = response.AccessToken

}

func loadAuth() {
	auth = &iyoAuthentication{
		url:          iyoURL,
		clientId:     setting.OAuthCfg.ClientID,
		clientSecret: setting.OAuthCfg.ClientSecret,
	}
}
