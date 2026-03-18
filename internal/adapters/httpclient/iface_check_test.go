package httpclient

import "github.com/justestif/go-jobs/internal/core/ports"

// Compile-time interface checks.
var _ ports.AuthService        = (*AuthClient)(nil)
var _ ports.SessionRepository  = (*AuthClient)(nil)
var _ ports.JobSearchService   = (*SearchClient)(nil)
var _ ports.ApplicationService = (*ApplicationClient)(nil)
