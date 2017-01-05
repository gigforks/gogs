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

To disable the standard registration flow, set the `DISABLE_REGISRATION` to `TRUE`
in the same custom/conf/app.ini file. This setting is found in the `[service]` section.
For example:

```
[service]
REGISTER_EMAIL_CONFIRM = false
ENABLE_NOTIFY_MAIL     = false
DISABLE_REGISTRATION   = true  <-----
ENABLE_CAPTCHA         = false
REQUIRE_SIGNIN_VIEW    = false
```

## ItsYou.Online login user handling

What happens when a user logs in with ItsYou.Online for the first time:

1. The username doesn't exist yet. A new user is made and the user can only log in
through his ItsYou.Online account.
2. The username already exists. The system will not make a new user, and give the
ItsYou.Online user access to the user stored in the system with the matchin username.
If someone has a 'regular' account, and doesn't have a matching ItsYou.Online account,
an attacker could possibly make an ItsYou.Online account with said username, and
assume control of the 'regular' account.


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
