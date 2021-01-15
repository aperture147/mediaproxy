package middleware

import (
	"context"
	"errors"
	"mediaproxy/util"
	"net/http"
	"os"
	"strings"
)

const SpecialRequestKey = "special"

type TokenAuthenticator struct {
	// Normal token
	Token string

	// Special token, use it to grant some special actions
	// like preserve image quality or override some default settings
	SpecialToken string
}

func NewTokenAuthenticator() TokenAuthenticator {
	token := os.Getenv("AUTH_TOKEN")
	if token == "" {
		token = "5FpA8Ad9uHCmCdPuf8sj49SpeyCrDTLAw4xAGUGH85Rf2phvQ77wDATjA2M4w8CD"
	}
	specialToken := os.Getenv("HQ_TOKEN")
	if specialToken == "" {
		specialToken = "8rzpd9ZMeQnnYQrCnQe924QeLsRwczzkZ6K6THnKU39fLAM2ZSbLKXdEBHKF934e"
	}
	return TokenAuthenticator{
		Token:        token,
		SpecialToken: specialToken,
	}
}

func (t TokenAuthenticator) Verify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			util.WriteForbiddenResponse(w, errors.New("no token provided"))
			return
		}

		rawAuth := strings.Split(r.Header.Get("Authorization"), " ")
		if len(rawAuth) == 2 && rawAuth[0] == "Bearer" && rawAuth[1] == t.Token {
			r.WithContext(context.WithValue(context.Background(), SpecialRequestKey, true))
			next.ServeHTTP(w, r)
			return
		}
		util.WriteUnauthorizedResponse(w, errors.New("wrong token"))
	})
}
