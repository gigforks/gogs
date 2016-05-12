// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package user

import (
	"net/url"

	"github.com/go-macaron/captcha"
	"golang.org/x/oauth2"
	"net/http"
	"encoding/json"
	"strings"
	"io/ioutil"
	"fmt"
	// "github.com/dchest/uniuri"

	"github.com/gigforks/gogs/models"
	"github.com/gigforks/gogs/modules/auth"
	"github.com/gigforks/gogs/modules/base"
	"github.com/gigforks/gogs/modules/log"
	"github.com/gigforks/gogs/modules/mailer"
	"github.com/gigforks/gogs/modules/middleware"
	"github.com/gigforks/gogs/modules/setting"
)

const (
	SIGNIN          base.TplName = "user/auth/signin"
	SIGNUP          base.TplName = "user/auth/signup"
	ACTIVATE        base.TplName = "user/auth/activate"
	FORGOT_PASSWORD base.TplName = "user/auth/forgot_passwd"
	RESET_PASSWORD  base.TplName = "user/auth/reset_passwd"
)

func SignIn(ctx *middleware.Context) {
	ctx.Data["Title"] = ctx.Tr("sign_in")

	// Check auto-login.
	isSucceed, err := middleware.AutoSignIn(ctx)
	if err != nil {
		ctx.Handle(500, "AutoSignIn", err)
		return
	}

	if isSucceed {
		if redirectTo, _ := url.QueryUnescape(ctx.GetCookie("redirect_to")); len(redirectTo) > 0 {
			ctx.SetCookie("redirect_to", "", -1, setting.AppSubUrl)
			ctx.Redirect(redirectTo)
		} else {
			ctx.Redirect(setting.AppSubUrl + "/")
		}
		return
	}

	ctx.HTML(200, SIGNIN)
}

func SignInPost(ctx *middleware.Context, form auth.SignInForm) {
	ctx.Data["Title"] = ctx.Tr("sign_in")

	if ctx.HasError() {
		ctx.HTML(200, SIGNIN)
		return
	}

	u, err := models.UserSignIn(form.UserName, form.Password)
	if err != nil {
		if models.IsErrUserNotExist(err) {
			ctx.RenderWithErr(ctx.Tr("form.username_password_incorrect"), SIGNIN, &form)
		} else {
			ctx.Handle(500, "UserSignIn", err)
		}
		return
	}

	if form.Remember {
		days := 86400 * setting.LogInRememberDays
		ctx.SetCookie(setting.CookieUserName, u.Name, days, setting.AppSubUrl)
		ctx.SetSuperSecureCookie(base.EncodeMD5(u.Rands+u.Passwd),
			setting.CookieRememberName, u.Name, days, setting.AppSubUrl)
	}

	ctx.Session.Set("uid", u.Id)
	ctx.Session.Set("uname", u.Name)
	if redirectTo, _ := url.QueryUnescape(ctx.GetCookie("redirect_to")); len(redirectTo) > 0 {
		ctx.SetCookie("redirect_to", "", -1, setting.AppSubUrl)
		ctx.Redirect(redirectTo)
		return
	}

	ctx.Redirect(setting.AppSubUrl + "/")
}

func SignOut(ctx *middleware.Context) {
	ctx.Session.Delete("uid")
	ctx.Session.Delete("uname")
	ctx.Session.Delete("socialId")
	ctx.Session.Delete("socialName")
	ctx.Session.Delete("socialEmail")
	ctx.SetCookie(setting.CookieUserName, "", -1, setting.AppSubUrl)
	ctx.SetCookie(setting.CookieRememberName, "", -1, setting.AppSubUrl)
	ctx.Redirect(setting.AppSubUrl + "/")
}

func SignUp(ctx *middleware.Context) {
	ctx.Data["Title"] = ctx.Tr("sign_up")

	ctx.Data["EnableCaptcha"] = setting.Service.EnableCaptcha

	if setting.Service.DisableRegistration {
		ctx.Data["DisableRegistration"] = true
		ctx.HTML(200, SIGNUP)
		return
	}

	ctx.HTML(200, SIGNUP)
}

func SignUpPost(ctx *middleware.Context, cpt *captcha.Captcha, form auth.RegisterForm) {
	ctx.Data["Title"] = ctx.Tr("sign_up")

	ctx.Data["EnableCaptcha"] = setting.Service.EnableCaptcha

	if setting.Service.DisableRegistration {
		ctx.Error(403)
		return
	}

	if ctx.HasError() {
		ctx.HTML(200, SIGNUP)
		return
	}

	if setting.Service.EnableCaptcha && !cpt.VerifyReq(ctx.Req) {
		ctx.Data["Err_Captcha"] = true
		ctx.RenderWithErr(ctx.Tr("form.captcha_incorrect"), SIGNUP, &form)
		return
	}

	if form.Password != form.Retype {
		ctx.Data["Err_Password"] = true
		ctx.RenderWithErr(ctx.Tr("form.password_not_match"), SIGNUP, &form)
		return
	}

	u := &models.User{
		Name:     form.UserName,
		Email:    form.Email,
		Passwd:   form.Password,
		IsActive: !setting.Service.RegisterEmailConfirm,
	}
	if err := models.CreateUser(u); err != nil {
		switch {
		case models.IsErrUserAlreadyExist(err):
			ctx.Data["Err_UserName"] = true
			ctx.RenderWithErr(ctx.Tr("form.username_been_taken"), SIGNUP, &form)
		case models.IsErrEmailAlreadyUsed(err):
			ctx.Data["Err_Email"] = true
			ctx.RenderWithErr(ctx.Tr("form.email_been_used"), SIGNUP, &form)
		case models.IsErrNameReserved(err):
			ctx.Data["Err_UserName"] = true
			ctx.RenderWithErr(ctx.Tr("user.form.name_reserved", err.(models.ErrNameReserved).Name), SIGNUP, &form)
		case models.IsErrNamePatternNotAllowed(err):
			ctx.Data["Err_UserName"] = true
			ctx.RenderWithErr(ctx.Tr("user.form.name_pattern_not_allowed", err.(models.ErrNamePatternNotAllowed).Pattern), SIGNUP, &form)
		default:
			ctx.Handle(500, "CreateUser", err)
		}
		return
	}
	log.Trace("Account created: %s", u.Name)

	// Auto-set admin for the only user.
	if models.CountUsers() == 1 {
		u.IsAdmin = true
		u.IsActive = true
		if err := models.UpdateUser(u); err != nil {
			ctx.Handle(500, "UpdateUser", err)
			return
		}
	}

	// Send confirmation e-mail, no need for social account.
	if setting.Service.RegisterEmailConfirm && u.Id > 1 {
		mailer.SendActivateAccountMail(ctx.Context, u)
		ctx.Data["IsSendRegisterMail"] = true
		ctx.Data["Email"] = u.Email
		ctx.Data["Hours"] = setting.Service.ActiveCodeLives / 60
		ctx.HTML(200, ACTIVATE)

		if err := ctx.Cache.Put("MailResendLimit_"+u.LowerName, u.LowerName, 180); err != nil {
			log.Error(4, "Set cache(MailResendLimit) fail: %v", err)
		}
		return
	}

	ctx.Redirect(setting.AppSubUrl + "/user/login")
}

func OauthAuthorize(ctx *middleware.Context)  {
	// Cfg, err := ini.Load(bindata.MustAsset("conf/app.ini"))
	// if err != nil {
	// 	log.Fatal(4, "Fail to parse 'conf/app.ini': %v", err)
	// }	
	ctx.Data["Title"] = ctx.Tr("oauth/authorize") 
	
	conf := &oauth2.Config{
		ClientID:     setting.Cfg.Section("oauth").Key("CLIENTID").String(),
		ClientSecret: setting.Cfg.Section("oauth").Key("CLIENTSECRET").String(),
		RedirectURL: setting.Cfg.Section("oauth").Key("REDIRECTURL").String(),
		Scopes:       []string{setting.Cfg.Section("oauth").Key("SCOPE").String()},
		Endpoint: oauth2.Endpoint{
			AuthURL:  setting.Cfg.Section("oauth").Key("AUTHURL").String(),
			TokenURL: setting.Cfg.Section("oauth").Key("TOKENURL").String(),
		},
	}
	
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOnline)
	ctx.Redirect(setting.AppSubUrl + url)
}

func OauthRedirect(ctx *middleware.Context) {	
	
	//request token using code 
	code := ctx.Query("code")	
	v := url.Values{}
	v.Add("code", code)
	v.Add("client_id", setting.Cfg.Section("oauth").Key("CLIENTID").String() )
	v.Add("client_secret",setting.Cfg.Section("oauth").Key("CLIENTSECRET").String() )
	v.Add("redirect_uri", setting.Cfg.Section("oauth").Key("REDIRECTURL").String() )
	v.Add("state", "state")	
	tokenResponse, _ := http.Post(setting.Cfg.Section("oauth").Key("TOKENURL").String(), "application/x-www-form-urlencoded", strings.NewReader(v.Encode()))
	
	//parse tokenResponse 
	tokenbd, _ := ioutil.ReadAll(tokenResponse.Body)
	type bdy struct {
		AccessToken string `json:"access_token"`
		TokenType string  `json:"token_type"`
		Scope string `json:"scope"`
		Info map[string]string `json:"info"`
	}
	data := bdy{}
	json.Unmarshal(tokenbd, &data)
	
	//get info and set proper variables
	accessToken := data.AccessToken
	username := data.Info["username"]
	client := http.Client{}
	req , err := http.NewRequest("GET", fmt.Sprintf("https://itsyou.online/users/%s/info", username), nil)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorize", fmt.Sprintf("token %s", accessToken))
	infoResponse, err := client.Do(req)	
	infobd, _ := ioutil.ReadAll(infoResponse.Body)
	// passwd := uniuri.NewLen(20)
	
	
	// //add to cookies and check user existence
	// ctx.SetCookie(setting.CookieUserName, username)
	// u := &models.User{
	// 	Name:     username,
	// 	Email:    form.Email,
	// 	Passwd:   passwd,
	// 	IsActive: !setting.Service.RegisterEmailConfirm,
	// }
	// if err := models.CreateUser(u); err != nil {
	// 	switch {
	// 	case models.IsErrUserAlreadyExist(err):
	// 		middleware.AutoSignIn(ctx)
	// 	case models.IsErrEmailAlreadyUsed(err):
	// 		ctx.Data["Err_Email"] = true
	// 		ctx.Handle(500, "email_been_used", err)
	// 	case models.IsErrNamePatternNotAllowed(err):
	// 		ctx.Data["Err_UserName"] = true
	// 		ctx.Handle(500, "user_pattern_not_allowed", err)
	// 	default:
	// 		ctx.Handle(500, "CreateUser", err)
	// 	}
	// 	return
	// }
	
	// log.Trace("Account created: %s", u.Name)

	// // Auto-set admin for the only user.
	// if models.CountUsers() == 1 {
	// 	u.IsAdmin = true
	// 	u.IsActive = true
	// 	if err := models.UpdateUser(u); err != nil {
	// 		ctx.Handle(500, "UpdateUser", err)
	// 		return
	// 	}
	// }

	// // Send confirmation e-mail, no need for social account.
	// if setting.Service.RegisterEmailConfirm && u.Id > 1 {
	// 	mailer.SendActivateAccountMail(ctx.Context, u)
	// 	ctx.Data["IsSendRegisterMail"] = true
	// 	ctx.Data["Email"] = u.Email
	// 	ctx.Data["Hours"] = setting.Service.ActiveCodeLives / 60
	// 	ctx.HTML(200, ACTIVATE)

	// 	if err := ctx.Cache.Put("MailResendLimit_"+u.LowerName, u.LowerName, 180); err != nil {
	// 		log.Error(4, "Set cache(MailResendLimit) fail: %v", err)
	// 	}
	// 	return
	// }	 
	
	//debugging 
	log.Debug("status code  ::: %v ",  infoResponse.StatusCode)
	log.Debug("body  ::: %v ",  string(infobd))	
	log.Debug("url sent ::: %s", req.URL.String())
	log.Debug("user name  :::" + username)
	if err == nil {
		panic(err)
	}

	// ctx.Data["Title"] = ctx.Tr("oauth/redirect")
	ctx.Redirect(setting.AppUrl + "/user/login")
}

func Activate(ctx *middleware.Context) {
	code := ctx.Query("code")
	if len(code) == 0 {
		ctx.Data["IsActivatePage"] = true
		if ctx.User.IsActive {
			ctx.Error(404)
			return
		}
		// Resend confirmation e-mail.
		if setting.Service.RegisterEmailConfirm {
			if ctx.Cache.IsExist("MailResendLimit_" + ctx.User.LowerName) {
				ctx.Data["ResendLimited"] = true
			} else {
				ctx.Data["Hours"] = setting.Service.ActiveCodeLives / 60
				mailer.SendActivateAccountMail(ctx.Context, ctx.User)

				if err := ctx.Cache.Put("MailResendLimit_"+ctx.User.LowerName, ctx.User.LowerName, 180); err != nil {
					log.Error(4, "Set cache(MailResendLimit) fail: %v", err)
				}
			}
		} else {
			ctx.Data["ServiceNotEnabled"] = true
		}
		ctx.HTML(200, ACTIVATE)
		return
	}

	// Verify code.
	if user := models.VerifyUserActiveCode(code); user != nil {
		user.IsActive = true
		user.Rands = models.GetUserSalt()
		if err := models.UpdateUser(user); err != nil {
			if models.IsErrUserNotExist(err) {
				ctx.Error(404)
			} else {
				ctx.Handle(500, "UpdateUser", err)
			}
			return
		}

		log.Trace("User activated: %s", user.Name)

		ctx.Session.Set("uid", user.Id)
		ctx.Session.Set("uname", user.Name)
		ctx.Redirect(setting.AppSubUrl + "/")
		return
	}

	ctx.Data["IsActivateFailed"] = true
	ctx.HTML(200, ACTIVATE)
}

func ActivateEmail(ctx *middleware.Context) {
	code := ctx.Query("code")
	email_string := ctx.Query("email")

	// Verify code.
	if email := models.VerifyActiveEmailCode(code, email_string); email != nil {
		if err := email.Activate(); err != nil {
			ctx.Handle(500, "ActivateEmail", err)
		}

		log.Trace("Email activated: %s", email.Email)
		ctx.Flash.Success(ctx.Tr("settings.add_email_success"))
	}

	ctx.Redirect(setting.AppSubUrl + "/user/settings/email")
	return
}

func ForgotPasswd(ctx *middleware.Context) {
	ctx.Data["Title"] = ctx.Tr("auth.forgot_password")

	if setting.MailService == nil {
		ctx.Data["IsResetDisable"] = true
		ctx.HTML(200, FORGOT_PASSWORD)
		return
	}

	ctx.Data["IsResetRequest"] = true
	ctx.HTML(200, FORGOT_PASSWORD)
}

func ForgotPasswdPost(ctx *middleware.Context) {
	ctx.Data["Title"] = ctx.Tr("auth.forgot_password")

	if setting.MailService == nil {
		ctx.Handle(403, "ForgotPasswdPost", nil)
		return
	}
	ctx.Data["IsResetRequest"] = true

	email := ctx.Query("email")
	ctx.Data["Email"] = email

	u, err := models.GetUserByEmail(email)
	if err != nil {
		if models.IsErrUserNotExist(err) {
			ctx.Data["Err_Email"] = true
			ctx.RenderWithErr(ctx.Tr("auth.email_not_associate"), FORGOT_PASSWORD, nil)
		} else {
			ctx.Handle(500, "user.ResetPasswd(check existence)", err)
		}
		return
	}

	if ctx.Cache.IsExist("MailResendLimit_" + u.LowerName) {
		ctx.Data["ResendLimited"] = true
		ctx.HTML(200, FORGOT_PASSWORD)
		return
	}

	mailer.SendResetPasswordMail(ctx.Context, u)
	if err = ctx.Cache.Put("MailResendLimit_"+u.LowerName, u.LowerName, 180); err != nil {
		log.Error(4, "Set cache(MailResendLimit) fail: %v", err)
	}

	ctx.Data["Hours"] = setting.Service.ActiveCodeLives / 60
	ctx.Data["IsResetSent"] = true
	ctx.HTML(200, FORGOT_PASSWORD)
}

func ResetPasswd(ctx *middleware.Context) {
	ctx.Data["Title"] = ctx.Tr("auth.reset_password")

	code := ctx.Query("code")
	if len(code) == 0 {
		ctx.Error(404)
		return
	}
	ctx.Data["Code"] = code
	ctx.Data["IsResetForm"] = true
	ctx.HTML(200, RESET_PASSWORD)
}

func ResetPasswdPost(ctx *middleware.Context) {
	ctx.Data["Title"] = ctx.Tr("auth.reset_password")

	code := ctx.Query("code")
	if len(code) == 0 {
		ctx.Error(404)
		return
	}
	ctx.Data["Code"] = code

	if u := models.VerifyUserActiveCode(code); u != nil {
		// Validate password length.
		passwd := ctx.Query("password")
		if len(passwd) < 6 {
			ctx.Data["IsResetForm"] = true
			ctx.Data["Err_Password"] = true
			ctx.RenderWithErr(ctx.Tr("auth.password_too_short"), RESET_PASSWORD, nil)
			return
		}

		u.Passwd = passwd
		u.Rands = models.GetUserSalt()
		u.Salt = models.GetUserSalt()
		u.EncodePasswd()
		if err := models.UpdateUser(u); err != nil {
			ctx.Handle(500, "UpdateUser", err)
			return
		}

		log.Trace("User password reset: %s", u.Name)
		ctx.Redirect(setting.AppSubUrl + "/user/login")
		return
	}

	ctx.Data["IsResetFailed"] = true
	ctx.HTML(200, RESET_PASSWORD)
}
