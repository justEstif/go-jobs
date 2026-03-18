# Enrichment Adapter Research

Covers: GenKit Go SDK, LLM provider structured output, per-user API key handling, encryption, and enrichment approaches from the wild.

---

## GenKit Go SDK

### Status
**Beta as of April 2025** — rich feature set, approaching production readiness. Go was announced stable (1.0) in the same cycle as the JS SDK. The `compat_oai` plugin layer (used by Anthropic and OpenAI) is production-ready.

### Structured output
`GenerateData[T]` is the idiomatic path — pass a Go struct, get a typed result back:

```go
type JobTags struct {
    RoleType      string   `json:"role_type"`
    Seniority     string   `json:"seniority"`
    RemotePolicy  string   `json:"remote_policy"`
    Country       string   `json:"country"`
    TechStack     []string `json:"tech_stack"`
}

tags, _, err := genkit.GenerateData[JobTags](ctx, g,
    ai.WithModel(model),
    ai.WithPrompt(prompt),
)
```

GenKit automatically converts the struct to a JSON schema and passes it to the model as a constrained output schema. If the model returns malformed JSON, GenKit returns an error — no silent bad data.

Alternatively use `Generate` + `WithOutputType` if you need access to token usage / finish reason:

```go
resp, err := genkit.Generate(ctx, g,
    ai.WithModel(model),
    ai.WithPrompt(prompt),
    ai.WithOutputType(JobTags{}),
)
var tags JobTags
resp.Output(&tags)
```

### Plugin model
Providers are registered as plugins at `genkit.Init` time:

```go
g := genkit.Init(ctx, genkit.WithPlugins(
    &openai.OpenAI{APIKey: "..."},
    &anthropic.Anthropic{Opts: []option.RequestOption{option.WithAPIKey("...")}},
    &googlegenai.GoogleAI{},
))
```

Models are then referenced by name string: `"openai/gpt-4.1"`, `"anthropic/claude-3-5-haiku"`, `"googleai/gemini-2.0-flash"`.

### Critical constraint: API key is bound at Init time

**The API key is baked into the client when the plugin is initialised.** There is no per-request key injection in the GenKit API. Source confirmed in `compat_oai.go`:

```go
func (o *OpenAICompatible) Init(ctx context.Context) []api.Action {
    client := openai.NewClient(o.Opts...)  // key is in Opts, set once
    o.client = &client
```

**This breaks the user-provided API key model** if we use GenKit directly — we cannot pass a different user's key per enrichment call without re-initialising the plugin, which panics on double-init.

### Consequence: GenKit is not viable for per-user API keys

Two options:

**Option A — Skip GenKit, call provider APIs directly**  
Use the raw OpenAI Go SDK (`github.com/openai/openai-go`), Anthropic SDK, or Google Gen AI SDK directly. Pass the user's decrypted API key as a client option per call. More code, but full control. The `JobEnricher` adapter wraps the raw SDK.

**Option B — GenKit for server-owned key only, raw SDK for user keys**  
If we ever add a server-level LLM key (e.g. for a free enrichment tier), use GenKit for that. For user-provided keys, always call raw SDKs. The `JobEnricher` interface hides this split.

**Recommendation: Option A for MVP.** The per-user key model is core to the product. GenKit adds complexity without benefit when we can't use its key-management. Use raw provider SDKs behind the `JobEnricher` port. GenKit can be revisited post-MVP if we want its observability/tracing features.

---

## LLM Provider Structured Output

All three providers support JSON-mode / structured output natively. No need for prompt hacking or output parsing.

### OpenAI
```
POST https://api.openai.com/v1/chat/completions
```
- `response_format: { type: "json_schema", json_schema: { schema: {...} } }` — strict structured output, guaranteed valid JSON matching schema
- Go SDK: `github.com/openai/openai-go` — official, maintained by OpenAI
- Key passed per-client: `openai.NewClient(option.WithAPIKey(userKey))`

### Anthropic
```
POST https://api.anthropic.com/v1/messages
```
- No native JSON schema mode — use system prompt with JSON instruction + tool use trick
- Best practice: define a tool with the desired schema, force tool use. Response comes back as structured tool call JSON.
- Go SDK: `github.com/anthropics/anthropic-sdk-go` — official
- Key passed per-client: `anthropic.NewClient(option.WithAPIKey(userKey))`

### Google Gemini
```
POST https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent
```
- `generationConfig.response_mime_type: "application/json"` + `response_schema` — native structured output
- Go SDK: `github.com/google/generative-ai-go`
- Key passed per-client at construction

### Comparison for our use case

| Provider | Structured output | Go SDK | Cost (cheapest model) | Notes |
|---|---|---|---|---|
| OpenAI | Native JSON schema | Official | ~$0.15/1M input (gpt-4.1-mini) | Cleanest structured output support |
| Anthropic | Tool-use trick | Official | ~$0.25/1M input (haiku-3.5) | Works well, slightly more setup |
| Gemini | Native JSON schema | Official | Free tier available (flash) | Best price; flash is fast |

**All three work fine for tag extraction.** The enrichment adapter needs a thin abstraction per provider — roughly 50–80 lines each.

---

## Enrichment Approaches in the Wild

### What others do

**Matt-Kobzik/job-scraper (Python, Anthropic)**
- Scores jobs 1–5 against a candidate profile using Claude Haiku
- Sends full job description in the prompt
- Simple, effective, cheap (~$0.001/job with Haiku)
- No tiered approach — LLM-only

**adgramigna/job-board-scraper (Python, no LLM)**
- Classifies jobs using dbt transformations — pure SQL/string matching on department names from ATS
- No LLM at all — maps `departments.name` → role category
- Works well because Greenhouse/Lever department names are fairly consistent
- Confirms our Tier 1 (ATS metadata) approach is sufficient for basic classification

**EchoJobs / standard aggregator approach**
- Typically: department/category from ATS → role type; title keywords → seniority; location field → remote policy
- LLM used for tech stack extraction and description summarisation, not basic classification
- Confirms: only tech stack extraction truly needs LLM

### What this means for our tier design

The research validates the tiered approach and sharpens where each tier earns its cost:

| Tag field      | Tier 1 (ATS) | Tier 2 (rules) | Tier 3 (LLM) | Confidence |
|---|---|---|---|---|
| `role_type`    | department name | title keywords | fallback | High from Tier 1+2 |
| `seniority`    | — | title keywords ("senior", "staff", "intern") | rarely needed | High from Tier 2 |
| `remote_policy`| workplaceType enum (Lever/Ashby) | location string | — | High from Tier 1 |
| `country`      | country/address fields | location string parse | — | High from Tier 1 |
| `tech_stack`   | — | keyword list against description | best coverage | LLM earns its cost here |
| `location_norm`| location string | string normalisation | — | Medium |

**Key insight:** Tech stack is the only field where LLM meaningfully beats rules. A keyword list catches `"Go"`, `"Postgres"`, `"React"` but misses contextual mentions (`"experience with the Go ecosystem"`, `"Python or equivalent"`). LLM is worth it here and nowhere else for MVP.

---

## API Key Encryption

For encrypting user LLM API keys at rest in PostgreSQL.

### Approach: AES-256-GCM with a server-side master key

Standard approach for encrypting short secrets in a DB column:

```go
import "crypto/aes"
import "crypto/cipher"
import "crypto/rand"

// Encrypt returns base64(nonce + ciphertext)
func Encrypt(plaintext, masterKey []byte) (string, error) {
    block, _ := aes.NewCipher(masterKey)  // masterKey must be 32 bytes for AES-256
    gcm, _ := cipher.NewGCM(block)
    
    nonce := make([]byte, gcm.NonceSize())
    rand.Read(nonce)
    
    ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt reverses the above
func Decrypt(encoded string, masterKey []byte) ([]byte, error) {
    data, _ := base64.StdEncoding.DecodeString(encoded)
    block, _ := aes.NewCipher(masterKey)
    gcm, _ := cipher.NewGCM(block)
    
    nonceSize := gcm.NonceSize()
    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    return gcm.Open(nil, nonce, ciphertext, nil)
}
```

### Master key management
- Stored as an environment variable (`APP_ENCRYPTION_KEY`) — never in the DB or source
- 32 random bytes, base64-encoded for easy env var storage
- Dokku: `dokku config:set APP_ENCRYPTION_KEY=$(openssl rand -base64 32)`
- Rotation: decrypt all rows with old key, re-encrypt with new key (offline migration)

### What this protects against
- DB dump / backup exposure — encrypted at rest
- Direct DB access — key not in DB

### What this does NOT protect against
- Compromise of the application process (key is in memory)
- Compromise of the env/config store
- This is standard "encryption at rest" — not end-to-end. Sufficient for MVP.

### No external KMS for MVP
AWS KMS / GCP KMS adds operational complexity for marginal security gain at this scale. Single-server Dokku deployment with env-var key is the right tradeoff for MVP. Note in the doc for post-MVP.

---

## Summary & Decisions

| Topic | Decision | Rationale |
|---|---|---|
| GenKit for enrichment | **No** | API key bound at Init — breaks per-user key model |
| LLM call implementation | **Raw provider SDKs** | Per-user key passed per-client construction |
| Provider priority for MVP | **Gemini Flash first** | Free tier, fast, native JSON schema |
| Enrichment tier 3 scope | **Tech stack extraction only** | Only field where LLM beats rules; all others covered by Tiers 1+2 |
| API key encryption | **AES-256-GCM, env var master key** | Standard approach, no external dependencies |
| GenKit revisit | **Post-MVP** | Worth revisiting for observability/tracing if we add server-level key |
