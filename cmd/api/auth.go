package main

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Kaungmyatkyaw2/book-store-api/internal/data"
	"github.com/Kaungmyatkyaw2/book-store-api/internal/validator"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

// LoginAccount godoc
// @Summary Log in to an account
// @Description Login to an account
// @Tags Authentication
// @Param request body LoginRequestBody true "Login data"
// @Produce  json
// @Success 200 {object} LoginResponse "Login success"
// @Failure 500 {object} InternalServerErrorResponse "Internal Server Error"
// @Failure 422 {object} GeneralErrorResponse "Validation Error"
// @Failure 401 {object} GeneralErrorResponse "Invalid Credential Error"
// @Router /v1/auth/login [post]
func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlainText(v, input.Password)

	if !v.IsValid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return

	}

	if user.AuthProvider != data.CredentialAuthProvider {
		app.badRequestResponse(w, r, errors.New("your account is not registerd with credentials"))
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	accessToken, err := app.createJWTToken(user.ID, time.Now().Add(time.Hour*24).Unix())

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.setRefreshTokenCookie(w, user)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJson(w, http.StatusOK, envelope{"accessToken": accessToken}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
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

	err := app.writeJson(w, http.StatusOK, envelope{"url": url}, nil)

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
			app.badRequestResponse(w, r, errors.New("your account is created with credentials"))
			return
		}

		token, err := app.createJWTToken(existingUser.ID, time.Now().Add(time.Hour*24).Unix())

		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		err = app.writeJson(w, http.StatusOK, envelope{"token": token}, nil)

		if err != nil {
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	user := &data.User{
		Name:      userInfo.Name,
		Email:     userInfo.Email,
		Picture:   userInfo.Picture,
		Activated: true,
	}

	err = user.Password.Set(userInfo.ID)

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

	token, err := app.createJWTToken(user.ID, time.Now().Add(time.Hour*24).Unix())

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJson(w, http.StatusOK, envelope{"token": token}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	rToken, err := r.Cookie("jwt")

	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			app.badRequestResponse(w, r, errors.New("refresh token isn't found in cookies."))
		default:
			app.serverErrorResponse(w, r, err)

		}
		return
	}

	verifiedToken, err := app.verifyJWTToken(rToken.Value)

	if err != nil {
		switch {
		case strings.HasPrefix(err.Error(), "invalid jwt"):
			app.badRequestResponse(w, r, err)
		default:
			app.serverErrorResponse(w, r, err)

		}
		return
	}

	var userID int64

	if claims, ok := verifiedToken.Claims.(jwt.MapClaims); ok {
		userID = int64(claims["userID"].(float64))
	}

	token, err := app.createJWTToken(userID, time.Now().Add(time.Hour*24).Unix())

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJson(w, http.StatusOK, envelope{"token": token}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
