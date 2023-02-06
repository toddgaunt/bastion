package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/toddgaunt/bastion/internal/auth"
	"github.com/toddgaunt/bastion/internal/errors"
)

const claimsKey contextKey = "claims"

var signKey, _ = auth.GenerateSymmetricKey()

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

var ErrUnauthorized = errors.Annotation{
	WithStatus: http.StatusUnauthorized,
	WithDetail: "enter username and password",
}

// AuthResponse is the payload containing the tokens that the authentication endpoints return.
type AuthResponse struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

// Authorize is a middleware that verifies the authorization token provided by
// a request, and continues to the next handler if successful.
func (env Env) Authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		claims, err := signKey.Verify(auth.JWT(token))
		if err != nil {
			prob := errors.Annotation{WithStatus: http.StatusUnauthorized, WithDetail: "couldn't verify token"}.Wrap(err)
			handleError(w, prob, env.Logger)

			return
		}

		fmt.Printf("try\n")

		ctx := context.WithValue(r.Context(), claimsKey, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Login authenticates a user and returns to them an access token and a refresh token.
func (env Env) Login(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) errors.Problem {
		username, password, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
			return ErrUnauthorized.Wrap(errors.New("user must enter basic auth"))
		}

		claims, err := env.Auth.Authenticate(username, password)
		if err != nil {
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
			return ErrUnauthorized.Wrap(err)
		}

		refreshToken, err := auth.NewBearerToken()
		if err != nil {
			return statusInternal.Wrap(err)
		}

		AddClaims(refreshToken, claims)

		accessToken, err := signKey.Sign(claims, time.Now(), duration)
		if err != nil {
			return statusInternal.Wrap(err)
		}

		resp := AuthResponse{
			RefreshToken: string(refreshToken),
			AccessToken:  string(accessToken),
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		return statusInternal.Wrap(enc.Encode(resp))
	}

	err := fn(w, r)
	handleError(w, err, env.Logger)
}

// Token accepts a refresh Token from the request and sends back a new refresh and access Token.
func (env Env) Token(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) errors.Problem {
		var refreshToken auth.BearerToken

		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&refreshToken)
		if err != nil {
			return errors.Annotation{
				WithStatus: http.StatusBadRequest,
				WithDetail: "failed to decode token",
			}.Wrap(err)
		}

		claims, ok := GetClaims(refreshToken)

		if !ok {
			return errors.Annotation{
				WithOp:     "Auth",
				WithStatus: http.StatusUnauthorized,
				WithDetail: "invalid refresh token",
			}.Wrap(errors.Errorf("refresh token %s not found", refreshToken))
			//w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
			//w.Header().Add("Location", "/auth/login")
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

		resp, prob := generateTokens(claims)
		if prob != nil {
			return prob
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		return errors.Annotation{
			WithStatus: http.StatusInternalServerError,
		}.Wrap(enc.Encode(resp))
	}

	err := fn(w, r)
	handleError(w, err, env.Logger)
}

func generateTokens(claims auth.Claims) (*AuthResponse, errors.Error) {
	refreshToken, err := auth.NewBearerToken()
	if err != nil {
		return nil, statusInternal.Wrap(err)
	}

	AddClaims(refreshToken, claims)

	accessToken, err := signKey.Sign(claims, time.Now(), duration)
	if err != nil {
		return nil, statusInternal.Wrap(err)
	}

	resp := AuthResponse{
		RefreshToken: string(refreshToken),
		AccessToken:  string(accessToken),
	}

	return &resp, nil
}
