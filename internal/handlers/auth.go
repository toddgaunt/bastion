package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/toddgaunt/bastion/internal/auth"
	"github.com/toddgaunt/bastion/internal/errors"
	"github.com/toddgaunt/bastion/internal/log"
)

const claimsKey contextKey = "claims"

var tokensMutex = sync.Mutex{}
var tokens = make(map[auth.BearerToken]auth.Claims)

var duration = time.Duration(time.Second * 5)

// GetClaims retrieves the claims associated with a bearer token.
func GetClaims(token auth.BearerToken) (auth.Claims, bool) {
	tokensMutex.Lock()
	claims, ok := tokens[token]
	delete(tokens, token)
	tokensMutex.Unlock()

	return claims, ok
}

// AddClaims associates a bearer token with authentication claims.
func AddClaims(token auth.BearerToken, claims auth.Claims) {
	tokensMutex.Lock()
	tokens[token] = claims
	tokensMutex.Unlock()
}

var invalidCredentials = errors.Note{
	StatusCode: http.StatusUnauthorized,
	Detail:     "enter a valid username and password",
}

// authResponse is the payload containing the tokens that the authentication endpoints return.
type authResponse struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

// Authorize is a middleware that verifies the authorization token provided by
// a request, and continues to the next handler if successful.
func (env Env) Authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		claims, err := env.SignKey.Verify(auth.JWT(token))
		if err != nil {
			prob := errors.Note{StatusCode: http.StatusUnauthorized, Detail: "couldn't verify token"}.Wrap(err)
			handleError(w, prob, env.Logger)

			return
		}

		if !claims.IsValid(env.Clock.Now()) {
			prob := errors.Note{}.Wrap(errors.New("expired claims"))
			handleError(w, prob, env.Logger)

			return
		}

		ctx := context.WithValue(r.Context(), claimsKey, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Login authenticates a user and returns to them an access token and a refresh token.
func (env Env) Login(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) errors.Problem {
		username, password, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("Www-Authenticate", `Basic realm="restricted"`)
			return invalidCredentials.Wrap(errors.New("user must enter basic auth"))
		}

		claims, err := env.Auth.Authenticate(username, password)
		if err != nil {
			w.Header().Set("Www-Authenticate", `Basic realm="restricted"`)
			return invalidCredentials.Wrap(err)
		}

		// Log successful authentications.
		env.Logger.With("username", username).Print(log.Info, "Authenticated user")

		refreshToken, err := auth.NewBearerToken()
		if err != nil {
			return statusInternal.Wrap(err)
		}

		AddClaims(refreshToken, claims)

		accessToken, err := env.SignKey.Sign(claims, env.Clock.Now(), duration)
		if err != nil {
			return statusInternal.Wrap(err)
		}

		resp := authResponse{
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
			return errors.Note{
				StatusCode: http.StatusBadRequest,
				Detail:     "failed to decode token",
			}.Wrap(err)
		}

		claims, ok := GetClaims(refreshToken)

		if !ok {
			return errors.Note{
				Op:         "Auth",
				StatusCode: http.StatusUnauthorized,
				Detail:     "invalid refresh token",
			}.Wrap(errors.Errorf("refresh token %s not found", refreshToken))
			//w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
			//w.Header().Add("Location", "/auth/login")
			//w.WriteHeader(http.StatusTemporaryRedirect)
			//return nil
		}

		if !claims.IsValid(env.Clock.Now()) {
			return errors.Note{
				Op:         "Auth",
				StatusCode: http.StatusUnauthorized,
				Detail:     "invalid refresh token",
			}.Wrap(errors.New("refresh token claims are expired"))
		}

		// Log successful authentications.
		env.Logger.With("token", refreshToken, "claims", claims).Print(log.Info, "Refreshed authentication")

		resp, prob := env.generateTokens(claims)
		if prob != nil {
			return prob
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(w)
		return errors.Note{
			StatusCode: http.StatusInternalServerError,
		}.Wrap(enc.Encode(resp))
	}

	err := fn(w, r)
	handleError(w, err, env.Logger)
}

func (env Env) generateTokens(claims auth.Claims) (*authResponse, errors.Error) {
	refreshToken, err := auth.NewBearerToken()
	if err != nil {
		return nil, statusInternal.Wrap(err)
	}

	AddClaims(refreshToken, claims)

	accessToken, err := env.SignKey.Sign(claims, env.Clock.Now(), duration)
	if err != nil {
		return nil, statusInternal.Wrap(err)
	}

	resp := authResponse{
		RefreshToken: string(refreshToken),
		AccessToken:  string(accessToken),
	}

	return &resp, nil
}
