package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Kaungmyatkyaw2/book-store-api/internal/data"
	"golang.org/x/oauth2"
)

func (app *application) generateStateOauthCookie(w http.ResponseWriter) string {
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	return state
}

func (app *application) getOauthUserInfo(code string) (*GoogleOauthResponse, error) {
	token, err := app.googleOauth.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}

	client := app.googleOauth.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result GoogleOauthResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (app *application) handleGoogleExistingUserLogin(w http.ResponseWriter, r *http.Request, user *data.User) {

	err := app.issueAccessToken(w, user.ID)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) handleGoogleNewUserLogin(w http.ResponseWriter, r *http.Request, userInfo *GoogleOauthResponse) {
	user := &data.User{
		Name:      userInfo.Name,
		Email:     userInfo.Email,
		Picture:   userInfo.Picture,
		Activated: true,
	}

	err := user.Password.Set(userInfo.ID)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	user.AuthProvider = data.GoogleOauthProvider

	err = app.models.Users.Insert(user)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return

	}

	err = app.issueAccessToken(w, user.ID)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// GoogleLogin godoc
// @Summary Log in with google
// @Description Login to an account using google oauth
// @Tags Authentication
// @Produce  json
// @Success 200 {object} GoogleLoginResponse "Return Redirect URL to continue Login with google"
// @Failure 500 {object} InternalServerErrorResponse "Internal Server Error"
// @Router /v1/auth/google [get]
func (app *application) googleLoginHandler(w http.ResponseWriter, r *http.Request) {

	oauthState := app.generateStateOauthCookie(w)

	url := app.googleOauth.AuthCodeURL(oauthState, oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	err := app.writeJSON(w, http.StatusOK, envelope{"url": url}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// GoogleLoginCallback godoc
// @Summary Callback for Google Login
// @Description Callback for google successful login
// @Tags Authentication
// @Produce  json
// @Success 200 {object} LoginResponse "Login success"
// @Failure 500 {object} InternalServerErrorResponse "Internal Server Error"
// @Failure 400 {object} GeneralErrorResponse "Bad Request Error"
// @Router /v1/auth/google/callback [get]
func (app *application) googleCallbackHandler(w http.ResponseWriter, r *http.Request) {

	code := r.FormValue("code")

	userInfo, err := app.getOauthUserInfo(code)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	existingUser, err := app.models.Users.GetByEmail(userInfo.Email)

	if err != nil && !errors.Is(err, data.ErrRecordNotFound) {
		app.serverErrorResponse(w, r, err)
		return
	}

	if existingUser != nil {

		if existingUser.AuthProvider == data.CredentialAuthProvider {
			app.badRequestResponse(w, r, errors.New("your account is registered by credentials"))
			return
		}

		app.handleGoogleExistingUserLogin(w, r, existingUser)
		return
	}

	app.handleGoogleNewUserLogin(w, r, userInfo)

}
