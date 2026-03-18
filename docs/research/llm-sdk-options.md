# LLM SDK Options for Go — Research

The core requirement: call OpenAI, Anthropic, and Gemini with **per-user API keys** and get **typed structured output** back. Evaluated every meaningful option.

---

## Candidates

### 1. GenKit Go SDK (firebase/genkit)

**Stars:** 5,663 | **Status:** Beta (Go), Production (JS)

API key bound at `Init` time — the client is constructed once when the plugin is registered. Re-calling `Init` panics. There is no mechanism to pass a per-request or per-user API key.

```go
// Key baked in at startup — cannot change per user
g := genkit.Init(ctx, genkit.WithPlugins(
    &anthropic.Anthropic{Opts: []option.RequestOption{option.WithAPIKey(staticKey)}},
))
```

**Verdict: eliminated.** Fundamentally incompatible with the per-user key model.

---

### 2. langchaingo (tmc/langchaingo)

**Stars:** 8,872 | **Last push:** Jan 2026 | **Status:** Community-maintained, no clear owner

Key is per-instance (passed at `New()` time), so constructing a new `LLM` per enrichment call with the user's key would work:

```go
llm, err := openai.New(openai.WithToken(userKey), openai.WithModel("gpt-4.1-mini"))
// use llm, then discard
```

**Structured output:** `WithJSONMode()` only — enables JSON object mode but **no JSON schema constraint**. The model is instructed to return JSON but the shape is not enforced. Confirmed by SO question (Jan 2025) that JSON schema (`ResponseFormatJSONSchema`) is not supported through the langchaingo API. You'd have to reach through to the underlying raw client to use it.

**Other issues:**
- Actively moving toward community ownership — "momentum for moving to community effort" in their own README. No clear lead maintainer.
- Abstracts away provider-specific features (Anthropic tool-use tricks, Gemini `response_schema`) — you lose access to the best structured output path for each provider.
- Heavy: pulls in chains, agents, memory, embeddings, vector stores — ~90% unused for our use case.

**Verdict: not recommended.** No JSON schema structured output. Uncertain maintenance trajectory. Leaky abstraction for a narrow use case.

---

### 3. gollm (teilomillet/gollm)

**Stars:** 641 | **Last push:** Feb 2026

Supports OpenAI, Anthropic, Groq, Ollama, Mistral, OpenRouter. Per-instance key. Has structured output with JSON schema validation.

```go
llm, err := gollm.NewLLM(
    gollm.SetProvider("openai"),
    gollm.SetModel("gpt-4.1-mini"),
    gollm.SetAPIKey(userKey),
)
prompt := gollm.NewPrompt("Extract job tags", gollm.WithJSONSchemaValidation())
```

**Issues:**
- 641 stars, small community, unclear long-term support
- JSON schema validation is post-generation (validates output against schema) rather than native provider-level schema enforcement — less reliable than passing schema directly to OpenAI/Gemini
- Adds abstraction that doesn't meaningfully simplify our specific use case (three providers, one structured call type)

**Verdict: not recommended.** Too niche, post-generation validation is weaker than native, not enough traction to trust.

---

### 4. Raw provider SDKs (official)

Three official Go SDKs maintained by the providers themselves:

| SDK | Repo | Stars | Last push | Structured output |
|---|---|---|---|---|
| `openai-go` | `openai/openai-go` | 3,059 | Mar 2026 | Native JSON schema (`ResponseFormatJSONSchema`) |
| `anthropic-sdk-go` | `anthropics/anthropic-sdk-go` | 905 | Mar 2026 | Tool-use pattern for structured output |
| `generative-ai-go` | `google/generative-ai-go` | 853 | Aug 2025 | Native `response_schema` + `response_mime_type` |

All three push the API key at client construction time — trivially per-user:

```go
// OpenAI
client := openai.NewClient(option.WithAPIKey(user.LLMAPIKey))

// Anthropic
client := anthropic.NewClient(option.WithAPIKey(user.LLMAPIKey))

// Gemini
client, _ := genai.NewClient(ctx, option.WithAPIKey(user.LLMAPIKey))
```

**Structured output examples:**

```go
// OpenAI — native JSON schema
type JobTags struct {
    RoleType     string   `json:"role_type"`
    Seniority    string   `json:"seniority"`
    RemotePolicy string   `json:"remote_policy"`
    Country      string   `json:"country"`
    TechStack    []string `json:"tech_stack"`
}
schema, _ := jsonschema.GenerateSchemaForType(JobTags{})
resp, _ := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
    Model: openai.ChatModelGPT4oMini,
    Messages: ...,
    ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
        OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
            JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
                Name:   "job_tags",
                Schema: schema,
                Strict: openai.Bool(true),
            },
        },
    },
})

// Anthropic — tool-use trick (no native JSON schema mode)
resp, _ := client.Messages.New(ctx, anthropic.MessageNewParams{
    Model: anthropic.ModelClaudeHaiku3_5,
    Tools: []anthropic.ToolParam{{
        Name:        "extract_job_tags",
        Description: "Extract structured tags from a job posting",
        InputSchema: jobTagsSchema,  // JSON schema as map
    }},
    ToolChoice: anthropic.ToolChoiceParam{OfToolChoiceTool: &anthropic.ToolChoiceToolParam{Name: "extract_job_tags"}},
    Messages: ...,
})

// Gemini — native response_schema
model := client.GenerativeModel("gemini-2.0-flash")
model.ResponseMIMEType = "application/json"
model.ResponseSchema = &genai.Schema{...}
resp, _ := model.GenerateContent(ctx, ...)
```

**Verdict: recommended.** Official, actively maintained, full access to each provider's structured output primitives, clean per-user client construction. The only downside is three separate integrations — but each is ~50–80 lines of adapter code behind our `JobEnricher` port.

---

## Decision

**Use raw provider SDKs behind the `JobEnricher` port.**

The `JobEnricher` interface stays clean:

```go
type JobEnricher interface {
    Enrich(ctx context.Context, job Job) (JobTags, error)
}
```

The tiered enrichment adapter picks the right SDK based on `user.LLMProvider`, constructs a fresh client with `user.LLMAPIKey` (decrypted), makes the call, and maps the response to `JobTags`. Each provider is its own ~60-line file in `internal/adapters/enrichment/`.

```
internal/adapters/enrichment/
├── enricher.go      -- orchestrates tiers 1→2→3, selects provider
├── ats.go           -- tier 1
├── rules.go         -- tier 2
├── llm_openai.go    -- tier 3: OpenAI
├── llm_anthropic.go -- tier 3: Anthropic
└── llm_gemini.go    -- tier 3: Gemini
```

No framework sits between us and the providers. The `JobEnricher` port is the only abstraction that matters.

---

## Updated architecture implications

- Remove **GenKit** from the tech stack table in `ARCHITECTURE.md`
- Replace with: **openai-go + anthropic-sdk-go + generative-ai-go (raw SDKs)**
- `LLMProvider` enum in `domain` stays unchanged — it now maps to which raw SDK the enrichment adapter instantiates
- No global LLM client in `main.go` — client is constructed per enrichment call from the decrypted user key
