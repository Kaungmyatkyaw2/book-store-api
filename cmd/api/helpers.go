package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Kaungmyatkyaw2/book-store-api/internal/validator"
	"github.com/julienschmidt/httprouter"
)

type envelope map[string]interface{}

type GoogleOauthResponse struct {
	ID         string `json:"sub"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Link       string `json:"link"`
	Picture    string `json:"picture"`
	Gender     string `json:"gender"`
	Locale     string `json:"locale"`
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
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

func (app *application) readParamInt(r *http.Request, key string) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	val, err := strconv.ParseInt(params.ByName(key), 10, 64)

	if err != nil {
		return -1, errors.New("invalid paramerter")
	}

	return val, nil
}

func (app *application) readIDParam(r *http.Request) (int64, error) {
	id, err := app.readParamInt(r, "id")
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

func (app *application) readString(qs url.Values, key string, defaultValue string) string {

	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return s

}

func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)

	app.logger.Debug("S : ", s)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)

	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return i

}
