package ports

import "errors"

// ErrNotFound is returned by repository methods when the requested record
// does not exist. Callers use errors.Is(err, ports.ErrNotFound) to distinguish
// a missing record from an infrastructure failure.
var ErrNotFound = errors.New("not found")
