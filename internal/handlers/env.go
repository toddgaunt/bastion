package handlers

import (
	"github.com/toddgaunt/bastion/internal/auth"
	"github.com/toddgaunt/bastion/internal/clock"
	"github.com/toddgaunt/bastion/internal/content"
	"github.com/toddgaunt/bastion/internal/log"
)

// Env stores all application state that isn't request specific. All fields
// must be safe for concurrent use.
type Env struct {
	Store  content.Store
	Logger log.Logger
	Clock  clock.Provider

	// TODO: split these fields into a separate environment for auth-only endpoints.
	// type AuthEnv struct
	Auth    auth.Authenticator
	SignKey auth.SymmetricKey
}
