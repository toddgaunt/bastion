package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/toddgaunt/bastion/internal/auth"
	"github.com/toddgaunt/bastion/internal/errors"
)

var key, _ = auth.GenerateSymmetricKey()

var tokensMutex = sync.Mutex{}
var tokens = make(map[auth.BearerToken]auth.Claims)

var duration = time.Duration(time.Second * 5)

func GetClaims(token auth.BearerToken) (auth.Claims, bool) {
	tokensMutex.Lock()
	claims, ok := tokens[token]
	delete(tokens, token)
	tokensMutex.Unlock()

	return claims, ok
}

func AddClaims(token auth.BearerToken, claims auth.Claims) {
	tokensMutex.Lock()
	tokens[token] = claims
	tokensMutex.Unlock()
}

func ErrUnauthorized(err error) error {
	return errors.Annotation{
		WithStatus: http.StatusUnauthorized,
	}.Wrap(err)
}

func ErrAuthenticate(err error) error {
	return errors.Annotation{
		WithStatus: http.StatusInternalServerError,
		WithDetail: "failed to authenticate",
	}.Wrap(err)
}

// AuthResponse is the payload containing the tokens that the authentication endpoints return.
type AuthResponse struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

// Refresh authenticates a user and returns to them an access token and a refresh token.
func Refresh(authenticator auth.Authenticator) func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) error {
		username, password, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
			return ErrUnauthorized(errors.New("enter username and password"))
		}

		claims, err := authenticator.Authenticate(username, password)
		if err != nil {
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
			return ErrUnauthorized(err)
		}

		refreshToken, err := auth.NewBearerToken()
		if err != nil {
			return ErrAuthenticate(err)
		}

		AddClaims(refreshToken, claims)

		accessToken, err := claims.Sign(time.Now(), duration, key)
		if err != nil {
			return ErrAuthenticate(err)
		}

		resp := AuthResponse{
			RefreshToken: string(refreshToken),
			AccessToken:  string(accessToken),
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		return enc.Encode(resp)
	}

	return wrapper(fn)
}

// Token accepts a refresh Token from the request and sends back a new refresh and access Token.
func Token(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) error {
		var refreshToken auth.BearerToken

		dec := json.NewDecoder(r.Body)
		dec.Decode(&refreshToken)

		claims, ok := GetClaims(refreshToken)

		if !ok {
			return errors.Annotation{
				WithOp:     "Auth",
				WithStatus: http.StatusUnauthorized,
				WithDetail: "invalid refresh token",
			}.Wrap(errors.Errorf("refresh token %s not found", refreshToken))
			//w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
			//w.Header().Add("Location", "/auth/refresh")
			//w.WriteHeader(http.StatusTemporaryRedirect)
			//return nil
		}

		if claims.IsExpired(time.Now()) {
			return errors.Annotation{
				WithOp:     "Auth",
				WithStatus: http.StatusUnauthorized,
				WithDetail: "invalid refresh token",
			}.Wrap(errors.New("refresh token claims are expired"))
		}

		resp, err := generateTokens(claims)
		if err != nil {
			return err
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		return enc.Encode(resp)
	}
	wrapper(fn)
}

func generateTokens(claims auth.Claims) (*AuthResponse, error) {
	refreshToken, err := auth.NewBearerToken()
	if err != nil {
		return nil, ErrAuthenticate(err)
	}

	AddClaims(refreshToken, claims)

	accessToken, err := claims.Sign(time.Now(), duration, key)
	if err != nil {
		return nil, ErrAuthenticate(err)
	}

	resp := AuthResponse{
		RefreshToken: string(refreshToken),
		AccessToken:  string(accessToken),
	}

	return &resp, nil
}
