package enrichment

import (
	"strings"

	"github.com/justestif/go-jobs/internal/core/domain"
)

// extractFromATS populates JobTags from fields the ATS already provided in
// the raw job payload. This is tier 1 — free, instant, no external calls.
//
// Fields extracted here:
//   - RemotePolicy  — from job.Location and scraped WorkplaceType (stored in Description prefix by scraper)
//   - Country       — ISO 2-letter code if the scraper captured it
//   - RoleType      — mapped from Department string
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

// departmentToRoleType maps an ATS department string to a domain.RoleType.
// Returns domain.RoleOther when no mapping matches.
func departmentToRoleType(department string) domain.RoleType {
	d := strings.ToLower(department)

	switch {
	case containsAny(d, "engineer", "software", "backend", "frontend", "platform", "infrastructure", "devops", "sre", "security", "mobile", "ios", "android", "ml", "machine learning", "ai", "data engineer"):
		return domain.RoleEngineering
	case containsAny(d, "design", "ux", "ui", "product design", "brand", "visual"):
		return domain.RoleDesign
	case containsAny(d, "product", "pm ", "program manager"):
		return domain.RoleProduct
	case containsAny(d, "data science", "analytics", "analyst", "business intelligence", "bi "):
		return domain.RoleData
	case containsAny(d, "marketing", "growth", "content", "seo", "brand"):
		return domain.RoleMarketing
	case containsAny(d, "sales", "account executive", "business development", "revenue", "partnerships"):
		return domain.RoleSales
	case containsAny(d, "operations", "ops", "support", "customer success", "hr", "people", "recruiting", "legal", "finance", "accounting"):
		return domain.RoleOperations
	case containsAny(d, "finance", "accounting", "fp&a", "treasury"):
		return domain.RoleFinance
	default:
		return domain.RoleOther
	}
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
