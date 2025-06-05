package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/julienschmidt/httprouter"
)

type envelope map[string]interface{}

func (app *application) writeJson(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")

	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {

	maxBytes := 1_048_576

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)

	if err != nil {

		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSOn type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unkown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")

			return fmt.Errorf("body must contains unkown key %s", fieldName)

		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err

		}

		return err
	}

	err = dec.Decode(&struct{}{})

	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

func (app *application) background(fn func()) {

	app.wg.Add(1)

	go func() {

		defer app.wg.Done()
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("%s", err)
				app.logger.Error("Something went wrong in this ")
			}
		}()

		fn()

	}()

}

func (app *application) createJWTToken(userID, expDate int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userID,
		"exp":    expDate,
	})

	tokenString, err := token.SignedString([]byte(app.config.jwt.secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GetUserByID godoc
// @Summary Get user by ID
// @Description Returns a single user
// @Tags users
// @Accept  json
// @Produce  json
// @Param   id  path  int  true  "User ID"
// @Success 200 {string} string "User found"
// @Failure 404 {string} string "User not found"
// @Router /api/users/{id} [get]
func (app *application) wrapHandler(h func(http.ResponseWriter, *http.Request, httprouter.Params)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ps := httprouter.ParamsFromContext(r.Context())
		h(w, r, ps)
	}
}
