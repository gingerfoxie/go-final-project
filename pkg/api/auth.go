package api

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
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

	if authR.Password == os.Getenv("TODO_PASSWORD") {

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
		pass := os.Getenv("TODO_PASSWORD")
		if len(pass) > 0 {
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
			//log.Println("old " + jwt)
			valid = strings.EqualFold(signedToken, jwt)
			if !valid {
				log.Println("wrong token")
				log.Println("new " + signedToken)
				log.Println("old " + jwt)
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}
			log.Println("auth OK")
		}
		next(w, r)
	})
}

func signedToken() (string, error) {

	envSecret := os.Getenv("TODO_PASSWORD")
	secret := []byte(envSecret)

	jwtToken := jwt.New(jwt.SigningMethodHS256)
	return jwtToken.SignedString(secret)

}
