package enrichment

import (
	"strings"

	"github.com/justestif/go-jobs/internal/core/domain"
)

// applyRules fills any zero-value fields in tags by running keyword and regex
// patterns against job.Title and job.Description. This is tier 2 — free, fast,
// no external calls.
//
// Fields targeted:
//   - Seniority    — keywords in title
//   - RoleType     — keywords in title when not already set by tier 1
//   - RemotePolicy — phrases in title/description when not set by tier 1
//   - TechStack    — keyword list match against description
//   - Country      — common country phrases in location/description
func applyRules(job domain.Job, tags domain.JobTags) domain.JobTags {
	title := strings.ToLower(job.Title)
	desc := strings.ToLower(job.Description)
	combined := title + " " + desc

	// Seniority from title keywords.
	if tags.Seniority == "" {
		tags.Seniority = inferSeniority(title)
	}

	// RoleType from title keywords when ATS department didn't cover it.
	if tags.RoleType == "" || tags.RoleType == domain.RoleOther {
		tags.RoleType = inferRoleType(title)
	}

	// RemotePolicy from description phrases when location string didn't cover it.
	if tags.RemotePolicy == "" {
		tags.RemotePolicy = inferRemotePolicy(combined)
	}

	// TechStack — always augment from description keywords.
	tags.TechStack = inferTechStack(desc)

	// Country — simple phrases in location or description.
	if tags.Country == "" {
		tags.Country = inferCountry(strings.ToLower(job.Location) + " " + desc)
	}

	return tags
}

// inferSeniority returns the seniority level based on keywords in the job title.
// Returns empty string when no keyword matches.
func inferSeniority(title string) domain.Seniority {
	switch {
	case containsAny(title, "intern", "internship", "co-op", "coop"):
		return domain.SeniorityIntern
	case containsAny(title, "junior", "jr.", "jr ", "associate ", "entry level", "entry-level", "new grad", "new-grad"):
		return domain.SeniorityJunior
	case containsAny(title, "staff "):
		return domain.SeniorityStaff
	case containsAny(title, "principal", "lead ", "tech lead", "team lead"):
		return domain.SeniorityLead
	case containsAny(title, "senior", "sr.", "sr "):
		return domain.SenioritySenior
	case containsAny(title, " ii ", " iii ", " iv ", "level 2", "level 3", "l2 ", "l3 ", "l4 ", "l5 "):
		return domain.SenioritySenior
	default:
		return domain.SeniorityMid
	}
}

// inferRoleType infers a RoleType from common keywords in the job title.
func inferRoleType(title string) domain.RoleType {
	switch {
	case containsAny(title, "software engineer", "backend", "frontend", "full stack", "fullstack", "platform engineer", "infrastructure", "devops", "site reliability", "sre", "security engineer", "mobile engineer", "ios engineer", "android engineer", "machine learning engineer", "ml engineer", "ai engineer", "data engineer", "firmware", "embedded"):
		return domain.RoleEngineering
	case containsAny(title, "product designer", "ux designer", "ui designer", "graphic designer", "brand designer", "visual designer"):
		return domain.RoleDesign
	case containsAny(title, "product manager", "program manager", "technical program"):
		return domain.RoleProduct
	case containsAny(title, "data scientist", "data analyst", "business analyst", "analytics engineer", "business intelligence"):
		return domain.RoleData
	case containsAny(title, "marketing", "growth", "content", "seo", "copywriter", "brand manager"):
		return domain.RoleMarketing
	case containsAny(title, "account executive", "sales", "business development", "revenue", "partnerships manager"):
		return domain.RoleSales
	case containsAny(title, "operations", "customer success", "customer support", "hr ", "recruiter", "people ops", "legal", "paralegal"):
		return domain.RoleOperations
	case containsAny(title, "finance", "accounting", "financial analyst", "fp&a", "controller", "treasurer"):
		return domain.RoleFinance
	default:
		return domain.RoleOther
	}
}

// inferRemotePolicy scans combined title+description text for remote-work phrases.
// Returns empty string when no phrase matches.
func inferRemotePolicy(text string) domain.WorkplaceType {
	switch {
	case containsAny(text, "fully remote", "100% remote", "remote-first", "remote first", "work from anywhere", "work from home", "wfh"):
		return domain.WorkplaceRemote
	case containsAny(text, "remote"):
		return domain.WorkplaceRemote
	case containsAny(text, "hybrid"):
		return domain.WorkplaceHybrid
	case containsAny(text, "on-site", "onsite", "in-office", "in office", "office-based"):
		return domain.WorkplaceOnsite
	default:
		return ""
	}
}

// techKeywords is the known tech-stack term list matched against job descriptions.
// Terms are lowercase; matched as whole words or substrings depending on context.
var techKeywords = []string{
	// Languages
	"go", "golang", "python", "rust", "java", "kotlin", "scala", "c++", "c#", "ruby",
	"typescript", "javascript", "swift", "elixir", "haskell", "clojure", "r ",
	// Frameworks / runtimes
	"react", "vue", "angular", "next.js", "nuxt", "svelte", "node.js", "express",
	"django", "flask", "fastapi", "rails", "spring", "gin", "fiber", "actix",
	// Data / ML
	"pytorch", "tensorflow", "spark", "kafka", "airflow", "dbt", "flink",
	"pandas", "numpy", "scikit-learn", "hugging face", "langchain",
	// Infrastructure / cloud
	"kubernetes", "k8s", "docker", "terraform", "ansible", "helm",
	"aws", "gcp", "azure", "cloudflare", "datadog", "prometheus", "grafana",
	// Databases
	"postgresql", "postgres", "mysql", "sqlite", "mongodb", "redis", "cassandra",
	"dynamodb", "bigquery", "snowflake", "clickhouse", "elasticsearch",
	// Other
	"graphql", "grpc", "rest api", "openapi", "protobuf", "kafka", "rabbitmq",
	"git", "github", "gitlab", "ci/cd", "llm", "openai", "anthropic",
}

// inferTechStack returns the subset of techKeywords found in the description.
func inferTechStack(desc string) []string {
	var found []string
	seen := make(map[string]bool)
	for _, kw := range techKeywords {
		if seen[kw] {
			continue
		}
		if strings.Contains(desc, kw) {
			found = append(found, kw)
			seen[kw] = true
		}
	}
	return found
}

// countryPhrases maps common location phrases to ISO 2-letter codes.
var countryPhrases = map[string]string{
	"united states":  "US",
	"u.s.":           "US",
	"usa":            "US",
	"new york":       "US",
	"san francisco":  "US",
	"bay area":       "US",
	"los angeles":    "US",
	"seattle":        "US",
	"austin":         "US",
	"boston":         "US",
	"chicago":        "US",
	"denver":         "US",
	"united kingdom": "GB",
	"u.k.":           "GB",
	"london":         "GB",
	"canada":         "CA",
	"toronto":        "CA",
	"vancouver":      "CA",
	"germany":        "DE",
	"berlin":         "DE",
	"munich":         "DE",
	"france":         "FR",
	"paris":          "FR",
	"netherlands":    "NL",
	"amsterdam":      "NL",
	"sweden":         "SE",
	"stockholm":      "SE",
	"australia":      "AU",
	"sydney":         "AU",
	"melbourne":      "AU",
	"india":          "IN",
	"bangalore":      "IN",
	"hyderabad":      "IN",
	"singapore":      "SG",
	"brazil":         "BR",
	"sao paulo":      "BR",
	"israel":         "IL",
	"tel aviv":       "IL",
}

// inferCountry scans text for known location phrases and returns an ISO code.
// Returns empty string when nothing matches.
func inferCountry(text string) string {
	for phrase, code := range countryPhrases {
		if strings.Contains(text, phrase) {
			return code
		}
	}
	return ""
}
