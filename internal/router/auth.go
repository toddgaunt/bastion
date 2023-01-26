package router

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/toddgaunt/bastion/internal/auth"
	"github.com/toddgaunt/bastion/internal/errors"
)

var key, _ = auth.GenerateSymmetricKey()

var RefreshTokensMutex = sync.Mutex{}
var RefreshTokens = make(map[auth.BearerToken]auth.Claims)

func ErrUnauthorized(err error) error {
	return errors.Annotate{
		WithOp:     "Auth",
		WithStatus: http.StatusUnauthorized,
	}.Wrap(err)
}

func ErrAuthenticate(err error) error {
	return errors.Annotate{
		WithOp:     "Auth",
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
func Refresh(w http.ResponseWriter, r *http.Request) error {
	username, password, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
		return ErrUnauthorized(errors.New("enter username and password"))
	}

	//TODO actual authentication
	if username != "foo" && password != "bar" {
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
		return ErrUnauthorized(errors.New("invalid username and password"))
	}

	refreshToken, err := auth.NewBearerToken()
	if err != nil {
		return ErrAuthenticate(err)
	}

	func() {
		RefreshTokensMutex.Lock()
		defer RefreshTokensMutex.Unlock()

		RefreshTokens[refreshToken] = auth.NewClaims(username, time.Now(), time.Hour*24)
	}()

	accessClaims := auth.NewClaims(username, time.Now(), time.Minute*5)
	accessToken, err := auth.Sign(accessClaims, key)
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

// Token accepts a refresh Token from the request and sends back a new refresh and access Token.
func Token(w http.ResponseWriter, r *http.Request) error {
	var refreshToken auth.BearerToken

	dec := json.NewDecoder(r.Body)
	dec.Decode(&refreshToken)

	claims, ok := func() (auth.Claims, bool) {
		RefreshTokensMutex.Lock()
		defer RefreshTokensMutex.Unlock()

		claims, ok := RefreshTokens[refreshToken]
		delete(RefreshTokens, refreshToken)
		return claims, ok
	}()

	if !ok {
		return errors.Annotate{
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
		return errors.Annotate{
			WithOp:     "Auth",
			WithStatus: http.StatusUnauthorized,
			WithDetail: "invalid refresh token",
		}.Wrap(errors.New("refresh token claims are expired"))
	}

	resp, err := generateTokens(claims.Username)
	if err != nil {
		return err
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	return enc.Encode(resp)
}

func generateTokens(username string) (*AuthResponse, error) {
	refreshToken, err := auth.NewBearerToken()
	if err != nil {
		return nil, ErrAuthenticate(err)
	}

	RefreshTokens[refreshToken] = auth.NewClaims(username, time.Now(), time.Hour*24)

	accessClaims := auth.NewClaims(username, time.Now(), time.Minute*5)
	accessToken, err := auth.Sign(accessClaims, key)
	if err != nil {
		return nil, ErrAuthenticate(err)
	}

	resp := AuthResponse{
		RefreshToken: string(refreshToken),
		AccessToken:  string(accessToken),
	}

	return &resp, nil
}
