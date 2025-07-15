package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
)

type authReq struct {
	Password string `json:"password"`
}

type authResp struct {
	Token string `json:"token"`
}

func SigninHandler(w http.ResponseWriter, req *http.Request) {

	var authR authReq
	var buf bytes.Buffer

	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		writeJson(w, errResp{Error: err.Error()}, http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(buf.Bytes(), &authR); err != nil {
		writeJson(w, errResp{Error: err.Error()}, http.StatusBadRequest)
		return
	}

	if authR.Password == Pass {

		signedToken, err := signedToken()
		if err != nil {
			writeJson(w, errResp{Error: err.Error()}, http.StatusInternalServerError)
			return
		}

		writeJson(w, authResp{Token: signedToken}, http.StatusOK)

	} else {
		writeJson(w, errResp{Error: "wrong password"}, http.StatusUnauthorized)
		return
	}

}

func auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(Pass) > 0 {
			var jwt string
			cookie, err := r.Cookie("token")
			if err == nil {
				jwt = cookie.Value
			}
			var valid bool
			signedToken, err := signedToken()

			if err != nil {
				writeJson(w, errResp{Error: err.Error()}, http.StatusInternalServerError)
				return
			}

			valid = strings.EqualFold(signedToken, jwt)
			if !valid {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}

		}
		next(w, r)
	})
}

func signedToken() (string, error) {

	secret := []byte(Pass)

	jwtToken := jwt.New(jwt.SigningMethodHS256)
	return jwtToken.SignedString(secret)

}
