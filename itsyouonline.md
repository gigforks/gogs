## To add Its you online application info.
Create app.ini in custom/conf directory with section like this

```
[itsyouonline]
CLIENT_ID     = YOUR CLIENT ID GOES HERE. (organization name)
CLIENT_SECRET = YOUR CLIENT SECRET.
REDIRECT_URL  = GOGS_ROOT URL /oauth/redirect
AUTH_URL      = https://itsyou.online/v1/oauth/authorize
TOKEN_URL     = https://itsyou.online/v1/oauth/access_token
SCOPE        = user:email

```

## To extend locales of the application
Create your custom locale under custom/conf ( for example you will have `custom/conf/locale/locale_en-US.ini`)
Add required words

```
sign_in_itsyouonline = Sign in using ItsyouOnline

```
Then in templates you can access it using
```
{{.i18n.Tr "sign_in_itsyouonline" }}
```
