# @justestif/go-jobs

CLI for [go-jobs](https://github.com/justEstif/go-jobs) — a self-hosted job aggregator that pulls from 1000+ companies on Greenhouse, Lever, and Ashby.

**Live site:** [jobs.estifanos.cc](https://jobs.estifanos.cc)

## Install

```sh
npm install -g @justestif/go-jobs
```

This downloads the prebuilt `go-jobs` binary for your platform (macOS, Linux, Windows).

## Quick start

```sh
# Target the hosted site
go-jobs --base-url https://jobs.estifanos.cc search --query "backend engineer"

# Register / login
go-jobs --base-url https://jobs.estifanos.cc register
go-jobs --base-url https://jobs.estifanos.cc login

# Search with filters
go-jobs search --query "frontend react" --role engineering --seniority senior

# Track applications
go-jobs interested <job-id>
go-jobs apply <job-id>
go-jobs pipeline
```

If you're running your own instance, set `--base-url` to your server or export `BASE_URL`.

## Job Coach

Analyze job postings against your resume with AI:

```sh
# Set your resume
go-jobs resume set --file ~/resume.md

# Analyze a job (requires LLM provider configured in web UI)
go-jobs analyze <job-id>

# Export the raw prompt to pipe to your own LLM
go-jobs prompt <job-id> | llm "analyze my fit for this role"
```

`go-jobs prompt` outputs the full analysis prompt without calling any LLM — no API key needed, just a resume on file.

## Use with AI agents

The CLI is designed to work as a tool for AI coding agents. Give your agent a prompt like:

> Search go-jobs for backend engineering roles, find one that matches my resume, analyze the fit, and draft a case study for my strongest matching project.

The agent can drive the full workflow:

```sh
go-jobs search --query "backend engineer" --role engineering --json
go-jobs resume set --file ~/resume.md
go-jobs analyze <job-id>
go-jobs prompt <job-id> | llm "optimize my resume for this role"
```

**MCP server coming soon.**

## Commands

| Command | Description |
|---|---|
| `go-jobs search` | Search jobs with filters |
| `go-jobs login` | Authenticate with email/password |
| `go-jobs register` | Create an account |
| `go-jobs logout` | Clear session |
| `go-jobs interested <id>` | Mark a job as interested |
| `go-jobs apply <id>` | Mark a job as applied |
| `go-jobs status <id> <status>` | Set job status |
| `go-jobs notes <id> <text>` | Add notes to a job |
| `go-jobs applied` | List applied jobs |
| `go-jobs pipeline` | View full pipeline |
| `go-jobs resume set` | Set resume from file or stdin |
| `go-jobs resume show` | Print stored resume |
| `go-jobs resume clear` | Remove stored resume |
| `go-jobs analyze <id>` | AI analysis of job vs resume |
| `go-jobs prompt <id>` | Export raw LLM prompt |
| `go-jobs scrape` | Run the scrape pipeline |
| `go-jobs enrich` | Run enrichment on untagged jobs |
| `go-jobs serve` | Start the web server |

## Self-hosting

```sh
docker pull ghcr.io/justestif/go-jobs:latest
```

See the [self-hosting guide](https://github.com/justEstif/go-jobs/blob/main/docs/self-hosting.md) for Docker Compose setup, environment variables, and deployment.

## Links

- [GitHub](https://github.com/justEstif/go-jobs)
- [Live site](https://jobs.estifanos.cc)
- [Architecture docs](https://github.com/justEstif/go-jobs/blob/main/docs/ARCHITECTURE.md)

## License

MIT
