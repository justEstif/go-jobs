# @justestif/go-jobs

CLI for [go-jobs](https://github.com/justEstif/go-jobs) — a self-hosted job aggregator that pulls from 1000+ companies on Greenhouse, Lever, and Ashby.

**Live site:** [jobs.estifanos.cc](https://jobs.estifanos.cc)

## Install

```sh
npm install -g @justestif/go-jobs
```

This downloads the prebuilt `jobs` binary for your platform (macOS, Linux, Windows).

## Quick start

```sh
# Target the hosted site
jobs --base-url https://jobs.estifanos.cc search --query "backend engineer"

# Register / login
jobs --base-url https://jobs.estifanos.cc register
jobs --base-url https://jobs.estifanos.cc login

# Search with filters
jobs search --query "frontend react" --role engineering --seniority senior

# Track applications
jobs interested <job-id>
jobs apply <job-id>
jobs pipeline
```

If you're running your own instance, set `--base-url` to your server or export `BASE_URL`.

## Job Coach

Analyze job postings against your resume with AI:

```sh
# Set your resume
jobs resume set --file ~/resume.md

# Analyze a job (requires LLM provider configured in web UI)
jobs analyze <job-id>

# Export the raw prompt to pipe to your own LLM
jobs prompt <job-id> | llm "analyze my fit for this role"
```

`jobs prompt` outputs the full analysis prompt without calling any LLM — no API key needed, just a resume on file.

## Use with AI agents

The CLI is designed to work as a tool for AI coding agents. Give your agent a prompt like:

> Search go-jobs for backend engineering roles, find one that matches my resume, analyze the fit, and draft a case study for my strongest matching project.

The agent can drive the full workflow:

```sh
jobs search --query "backend engineer" --role engineering --json
jobs resume set --file ~/resume.md
jobs analyze <job-id>
jobs prompt <job-id> | llm "optimize my resume for this role"
```

**MCP server coming soon.**

## Commands

| Command | Description |
|---|---|
| `jobs search` | Search jobs with filters |
| `jobs login` | Authenticate with email/password |
| `jobs register` | Create an account |
| `jobs logout` | Clear session |
| `jobs interested <id>` | Mark a job as interested |
| `jobs apply <id>` | Mark a job as applied |
| `jobs status <id> <status>` | Set job status |
| `jobs notes <id> <text>` | Add notes to a job |
| `jobs applied` | List applied jobs |
| `jobs pipeline` | View full pipeline |
| `jobs resume set` | Set resume from file or stdin |
| `jobs resume show` | Print stored resume |
| `jobs resume clear` | Remove stored resume |
| `jobs analyze <id>` | AI analysis of job vs resume |
| `jobs prompt <id>` | Export raw LLM prompt |
| `jobs scrape` | Run the scrape pipeline |
| `jobs enrich` | Run enrichment on untagged jobs |
| `jobs serve` | Start the web server |

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
