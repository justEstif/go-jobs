package enrichment

import (
	"strings"

	"github.com/justestif/go-jobs/internal/core/domain"
)

// extractFromATS populates JobTags from fields the ATS already provided in
// the raw job payload. This is tier 1 — free, instant, no external calls.
//
// Fields extracted here:
//   - RemotePolicy  — from job.Location keywords ("remote", "hybrid", "onsite")
//   - Country       — ISO 2-letter code when job.Location is exactly 2 uppercase chars
//   - LocationNorm  — raw location string as-is; later tiers may normalise further
//
// Returns the partially-filled tags; zero values indicate the field was not
// available from ATS data and should be filled by a later tier.
func extractFromATS(job domain.Job) domain.JobTags {
	tags := domain.JobTags{
		JobID: job.ID,
	}

	// Country — scrapers capture ISO codes when the ATS provides them.
	// Stored in plain text; upper-case and length-check for safety.
	if country := strings.TrimSpace(strings.ToUpper(job.Location)); len(country) == 2 {
		tags.Country = country
	}

	// RemotePolicy — derive from raw location string keywords.
	loc := strings.ToLower(job.Location)
	switch {
	case strings.Contains(loc, "remote"):
		tags.RemotePolicy = domain.WorkplaceRemote
	case strings.Contains(loc, "hybrid"):
		tags.RemotePolicy = domain.WorkplaceHybrid
	case strings.Contains(loc, "on-site") || strings.Contains(loc, "onsite") || strings.Contains(loc, "in-office"):
		tags.RemotePolicy = domain.WorkplaceOnsite
	}

	// LocationNorm — use the raw location as-is for tier 1; LLM/rules can
	// normalise further. Non-empty only when we have something meaningful.
	if loc := strings.TrimSpace(job.Location); loc != "" {
		tags.LocationNorm = loc
	}

	return tags
}

// containsAny returns true if s contains any of the given substrings.
func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
