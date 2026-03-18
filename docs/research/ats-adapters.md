# ATS Adapter Research

## Company Discovery

The Simplify GitHub org (`github.com/SimplifyJobs`) maintains two community-curated job repos with direct ATS apply links:

- `SimplifyJobs/Summer2026-Internships` (dev branch)
- `SimplifyJobs/New-Grad-Positions` (dev branch)

The README markdown tables embed direct ATS URLs. Parsing them yields company slugs/tokens for all three platforms with no scraping or authentication:

| Platform   | URL pattern                                      | Extraction regex                                        |
| ---------- | ------------------------------------------------ | ------------------------------------------------------- |
| Greenhouse | `job-boards.greenhouse.io/{board_token}/jobs/…`  | `job-boards\.greenhouse\.io/([^/]+)/jobs`               |
| Lever      | `jobs.lever.co/{company_slug}/…`                 | `jobs\.lever\.co/([^/]+)/`                              |
| Ashby      | `jobs.ashbyhq.com/{board_name}/…`                | `jobs\.ashbyhq\.com/([^/]+)/`                           |

**Counts after merging + deduplicating both READMEs (verified):**

| Platform   | Unique companies |
| ---------- | ---------------- |
| Greenhouse | 154              |
| Lever      | 49               |
| Ashby      | 93               |
| **Total**  | **296**          |

These are all startup/growth-stage companies — exactly the target. Zero Workday. Zero enterprise ATS noise.

**Seed strategy for MVP:** On first run, fetch both READMEs, parse the URLs, upsert companies into the DB. Subsequent scrape cycles keep the list fresh by re-parsing (new companies appear as the repos update daily). The raw README URLs are stable — no API key required.

Raw README URLs:
- `https://raw.githubusercontent.com/SimplifyJobs/Summer2026-Internships/dev/README.md`
- `https://raw.githubusercontent.com/SimplifyJobs/New-Grad-Positions/dev/README.md`

---

Live-verified against real API responses. Each section covers: endpoint, response shape, what's useful for enrichment tier 1, and implementation notes.

---

## Greenhouse

### Discovery
Companies are discovered by scraping the Greenhouse platform directory (post-MVP). For MVP, the board token is the company's slug (e.g. `greenhouse`, `stripe`).

### Endpoint

```
GET https://boards-api.greenhouse.io/v1/boards/{board_token}/jobs?content=true
```

No authentication required. `content=true` includes description HTML, departments, and offices in the list response — avoids a second per-job request.

### Response shape (verified)

```json
{
  "jobs": [
    {
      "id": 7609008,
      "internal_job_id": 3360516,
      "title": "Brand Designer, EMEA",
      "location": { "name": "Anywhere in Ireland" },
      "absolute_url": "https://job-boards.greenhouse.io/greenhouse/jobs/7609008",
      "updated_at": "2026-02-24T11:10:17-05:00",
      "first_published": "2026-02-24T11:10:17-05:00",
      "content": "<div>...</div>",        // HTML description
      "departments": [
        { "id": 229071, "name": "Brand & Creative" }
      ],
      "offices": [
        { "id": 88404, "name": "Republic of Ireland", "location": null }
      ],
      "metadata": null                    // custom fields — company-dependent
    }
  ],
  "meta": { "total": 42 }
}
```

### Enrichment tier 1 — fields available from ATS

| `job_tags` field  | Source field              | Notes                                                  |
| ----------------- | ------------------------- | ------------------------------------------------------ |
| `role_type`       | `departments[].name`      | "Engineering", "Product", "Design", "Sales", etc.      |
| `location_norm`   | `location.name`           | Raw string — needs normalisation                       |
| `country`         | `offices[].location`      | Sometimes present as "New York, NY, United States"     |
| `remote_policy`   | `location.name`           | Look for "Remote" / "Hybrid" in string — unreliable    |
| `seniority`       | —                         | Not in API — title parsing only                        |
| `tech_stack`      | —                         | Not in API — description parsing only                  |

### Dedup key
`external_id` = `jobs[].id` (integer, cast to string). Stable across scrapes.

### Implementation notes
- `internal_job_id` is the job entity; `id` is the posting. A job can have multiple postings (e.g. different locations). Use `id` (posting) as dedup key since we're tracking postings, not jobs.
- Filter out rows where `internal_job_id` is null — those are prospect/general application posts, not real job postings.
- `content` field contains HTML-encoded entities (double-escaped). Unescape before storing.
- No pagination parameter in v1 — returns all active postings in one response. Large boards may be slow; add a timeout.
- `metadata` is company-configurable custom fields. Mostly null. Don't rely on it.

---

## Lever

### Discovery
Company slugs are the account name used on `jobs.lever.co` (e.g. `jobs.lever.co/welocalize` → slug is `welocalize`).

### Endpoint

```
GET https://api.lever.co/v0/postings/{company_slug}?mode=json&limit=100
```

This is the **public Postings API (v0)** — no authentication required. Distinct from the authenticated v1 API which is for internal HR use only.

### Response shape (verified)

```json
[
  {
    "id": "6e0ac589-3fba-429a-8ce4-5c48a2cf0e2d",
    "text": "Junior Quality Coordinator - Portuguese (Brazil)",
    "createdAt": 1675993015566,
    "hostedUrl": "https://jobs.lever.co/welocalize/6e0ac589-3fba-429a-8ce4-5c48a2cf0e2d",
    "applyUrl": "https://jobs.lever.co/welocalize/6e0ac589-3fba-429a-8ce4-5c48a2cf0e2d/apply",
    "categories": {
      "commitment": "Freelance-Remote",
      "department": "Welo Data - AI Services",
      "location": "Brazil",
      "team": "Data Validation",
      "allLocations": ["Brazil"]
    },
    "description": "<div>...</div>",
    "descriptionPlain": "...",
    "country": "BR",
    "workplaceType": "remote"
  }
]
```

### Enrichment tier 1 — fields available from ATS

| `job_tags` field  | Source field                | Notes                                              |
| ----------------- | --------------------------- | -------------------------------------------------- |
| `role_type`       | `categories.department`     | Free text — "Engineering", "Product", etc.         |
| `location_norm`   | `categories.location`       | City/country string                                |
| `country`         | `country`                   | ISO 2-letter code — reliable                       |
| `remote_policy`   | `workplaceType`             | Enum: `"remote"` / `"hybrid"` / `"on-site"` — reliable |
| `seniority`       | —                           | Not in API — title parsing only                    |
| `tech_stack`      | —                           | Not in API — description parsing only              |

### Dedup key
`external_id` = `id` (UUID string). Stable.

### Implementation notes
- `workplaceType` is the best structured remote-policy signal of any of the three platforms — use it directly.
- `country` is ISO 2-letter — map to our normalised country format.
- `categories.commitment` maps loosely to employment type ("Full-time", "Part-time", "Contract", "Freelance-Remote").
- Response is a JSON array (not wrapped in an object). Empty array `[]` means no postings or invalid slug.
- `limit` param supported. Default appears to be all. Use `limit=100&offset=...` for pagination if needed — verify behaviour on large boards.
- `createdAt` is Unix ms timestamp.

---

## Ashby

### Discovery
Company board names come from `jobs.ashbyhq.com/{name}`. The name is the path segment.

### Endpoint

```
GET https://api.ashbyhq.com/posting-api/job-board/{job_board_name}
```

No authentication required. Single request returns all listed postings.

### Response shape (verified)

```json
{
  "apiVersion": "1",
  "jobs": [
    {
      "id": "145ff46b-1441-4773-bcd3-c8c90baa598a",
      "title": "Engineer Who Can Design, Americas",
      "department": "Engineering",
      "team": "Americas Engineering",
      "employmentType": "FullTime",
      "location": "Remote - North to South America",
      "isRemote": true,
      "workplaceType": "Remote",
      "publishedAt": "2025-11-14T00:46:58.989+00:00",
      "isListed": true,
      "address": {
        "postalAddress": {
          "addressLocality": "San Francisco",
          "addressRegion": "California",
          "addressCountry": "USA"
        }
      },
      "secondaryLocations": [
        {
          "location": "Remote - US",
          "address": {
            "postalAddress": {
              "addressLocality": "San Francisco",
              "addressRegion": "California",
              "addressCountry": "United States"
            }
          }
        }
      ],
      "jobUrl": "https://jobs.ashbyhq.com/ashby/145ff46b-...",
      "applyUrl": "https://jobs.ashbyhq.com/ashby/145ff46b-.../application",
      "descriptionHtml": "<p>...</p>",
      "descriptionPlain": "..."
    }
  ]
}
```

### Enrichment tier 1 — fields available from ATS

| `job_tags` field  | Source field                          | Notes                                                         |
| ----------------- | ------------------------------------- | ------------------------------------------------------------- |
| `role_type`       | `department`                          | Clean string — "Engineering", "Product", "Design", etc.       |
| `location_norm`   | `location`                            | Human-readable string                                         |
| `country`         | `address.postalAddress.addressCountry`| Full country name ("United States", "USA") — needs mapping    |
| `remote_policy`   | `workplaceType`                       | Enum: `"OnSite"` / `"Remote"` / `"Hybrid"` — reliable        |
| `seniority`       | `employmentType`                      | `"FullTime"` / `"PartTime"` / `"Intern"` / `"Contract"` — partial signal for intern |
| `tech_stack`      | —                                     | Not in API — description parsing only                         |

### Dedup key
`external_id` = `id` (UUID string). Stable.

### Implementation notes
- `isListed` — filter to `true` only. Unlisted postings exist in the API but should not be displayed publicly.
- `workplaceType` is an enum with consistent casing (`"OnSite"`, `"Remote"`, `"Hybrid"`). Map directly.
- `employmentType` gives a partial seniority signal: `"Intern"` is reliable; `"FullTime"` covers everything else.
- `secondaryLocations` is useful for multi-location postings — store in `location` field, consider comma-joining for display.
- `addressCountry` inconsistently uses "USA" vs "United States" — normalise on ingest.
- `publishedAt` is ISO 8601 with timezone. Parse carefully.

---

## Enrichment tier 1 — summary across platforms

| `job_tags` field  | Greenhouse         | Lever                   | Ashby                    | Tier 2 fallback       |
| ----------------- | ------------------ | ----------------------- | ------------------------ | --------------------- |
| `role_type`       | `departments.name` | `categories.department` | `department`             | title keyword match   |
| `location_norm`   | `location.name`    | `categories.location`   | `location`               | description text      |
| `country`         | `offices.location` | `country` (ISO)         | `address.addressCountry` | location string parse |
| `remote_policy`   | location string    | `workplaceType` (enum)  | `workplaceType` (enum)   | keyword match         |
| `seniority`       | —                  | —                       | `employmentType` (Intern only) | title keywords  |
| `tech_stack`      | —                  | —                       | —                        | description keywords  |

**Key insight:** `remote_policy` is well-structured in Lever and Ashby (enums). Greenhouse is string-only — parse the location name. `tech_stack` and `seniority` (beyond intern) always require tier 2 or 3.

---

## `RawJob` shape — what scrapers return to the core

The `JobScraper` port returns `[]RawJob`. This is the normalised intermediate type — platform differences are resolved inside each adapter before returning.

```go
// core/domain/raw_job.go
type RawJob struct {
    ExternalID  string    // dedup key — platform posting ID
    Title       string
    URL         string    // direct link to apply
    Location    string    // raw location string
    Description string    // plain text
    RawHTML     string    // original HTML — stored for re-enrichment
    FirstSeen   time.Time

    // ATS metadata — populated where available; zero value = unknown
    Department  string    // role_type signal
    Country     string    // ISO 2-letter if available, full name otherwise
    WorkplaceType string  // "remote" | "hybrid" | "onsite" | "" (empty = unknown)
    EmploymentType string // "fulltime" | "parttime" | "intern" | "contract" | ""
}
```

Each adapter maps its platform response to `RawJob`. Country normalisation (ISO code lookup) happens in the adapter. Workplace type normalisation (lower-cased enum) happens in the adapter. The core receives a clean, platform-agnostic struct.
