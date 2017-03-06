package user

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/gogits/gogs/models"
	"github.com/gogits/gogs/modules/base"
	"github.com/gogits/gogs/modules/context"
	"github.com/gogits/gogs/modules/setting"
	"golang.org/x/oauth2"
)

func OAuthAuthorize(ctx *context.Context) {

	cnf := &oauth2.Config{
		ClientID:     setting.OAuthCfg.ClientID,
		ClientSecret: setting.OAuthCfg.ClientSecret,
		RedirectURL:  extractRedirectUrl(ctx),
		Scopes:       []string{setting.OAuthCfg.Scope},
		Endpoint: oauth2.Endpoint{
			AuthURL:  setting.OAuthCfg.AuthURL,
			TokenURL: setting.OAuthCfg.TokenURL,
		},
	}
	rnd := base.GetRandomString(10)

	ctx.Session.Set("state", rnd)
	codeURL := cnf.AuthCodeURL(rnd)
	ctx.Redirect(codeURL)

}

func OAuthRedirect(ctx *context.Context) {
	//EXCHANGE CODE FOR ACCESS TOKEN.
	code := ctx.Query("code")
	v := url.Values{}
	v.Add("client_id", setting.OAuthCfg.ClientID)
	v.Add("client_secret", setting.OAuthCfg.ClientSecret)
	v.Add("redirect_uri", extractRedirectUrl(ctx))
	v.Add("code", code)
	v.Add("state", ctx.Session.Get("state").(string))

	tokresp, err := http.Post(setting.OAuthCfg.TokenURL, "application/x-www-form-urlencoded", strings.NewReader(v.Encode()))

	if err != nil {
		ctx.Handle(500, "error in post tokenURL", nil)
	}
	type resp struct {
		AccessToken string            `json:"access_token"`
		TokenType   string            `json:"token_type"`
		Scope       string            `json:"scope"`
		Info        map[string]string `json:"info"`
	}
	data := &resp{}
	decoder := json.NewDecoder(tokresp.Body)
	decoder.Decode(&data)

	//get info and set proper variables
	if data.AccessToken == "" {
		ctx.HandleText(500, "client error")
	}

	// CALL TO USERS BACKEND API TO GET EMAIL
	username := data.Info["username"]
	userapi := "https://itsyou.online/api/users/" + username + "/info"

	type EmailEntry struct {
		EmailAddress string `json:"emailaddress"`
		Label        string `json:"label"`
	}

	type InfoResponse struct {
		EmailAddresses []EmailEntry `json:"emailaddresses"`
	}
	var infoRes InfoResponse

	req, _ := http.NewRequest("GET", userapi, nil)
	req.Header.Set("Authorization", "token "+data.AccessToken)

	client := &http.Client{}
	reqresp, err := client.Do(req)
	txt, _ := ioutil.ReadAll(reqresp.Body)

	if err := json.Unmarshal(txt, &infoRes); err != nil {
		ctx.Handle(500, "CreateUser", err)
	}
	email := infoRes.EmailAddresses[0].EmailAddress

	// Now test if user exists in database or not if doesn't create a new one
	u, err := models.GetUserByName(username)

	if err != nil {
		if models.IsErrUserNotExist(err) {
			// CREATE USER
			u := &models.User{
				Name:     username,
				Email:    email,
				Passwd:   base.GetRandomString(10),
				IsActive: true,
			}

			if err := models.CreateUser(u); err != nil {
				ctx.Handle(500, "CreateUser", err)

			}
		}
	}
	// USER EXISTS IN DB AT THIS SECOND.
	u, err = models.GetUserByName(username)
	if err == nil {
		ctx.Session.Set("uid", u.ID)
		ctx.Session.Set("uname", u.Name)
		ctx.Redirect("/")

	} else {
		ctx.Handle(500, "CreateUser", err)

	}

	// Now that the user is logged add the organizations
	orgs := make([]string, 0)
	scopes := strings.Split(data.Scope, ",")
	for _, scope := range scopes {
		if strings.HasPrefix(scope, "user:memberof:") {
			orgs = append(orgs, strings.TrimPrefix(scope, "user:memberof:"))
		}
	}
	ctx.Session.Set("organizations", orgs)
}

func extractRedirectUrl(ctx *context.Context) string {
	url := ctx.Context.Req.Request.URL
	redirectUrl := url.Scheme
	// use http:// scheme if there isn't one, usefull for dev...
	if len(redirectUrl) == 0 {
		redirectUrl = "http://"
	}
	redirectUrl += ctx.Context.Req.Request.Host
	redirectUrl += "/oauth/redirect"

	return redirectUrl
}
