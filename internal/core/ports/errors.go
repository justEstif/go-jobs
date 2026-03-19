package ports

import "errors"

// ErrNotFound is returned by repository methods when the requested record
// does not exist. Callers use errors.Is(err, ports.ErrNotFound) to distinguish
// a missing record from an infrastructure failure.
var ErrNotFound = errors.New("not found")

// ErrInvalidCredentials is returned by AuthService methods when the supplied
// password does not match the stored hash. Used by Login, ChangePassword, and
// DeleteAccount. HTTP adapters map this to a user-facing error message.
var ErrInvalidCredentials = errors.New("invalid email or password")
