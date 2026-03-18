# Go Web Template

Modern Go web application template with Chi, Templ, Tailwind CSS + DaisyUI, and HTMX.

## Features

- **Chi** - Lightweight, idiomatic HTTP router
- **Templ** - Type-safe Go templating
- **Tailwind CSS + DaisyUI** - Utility-first CSS with a component library (standalone, no Node required)
- **HTMX** - Dynamic interactions without heavy JavaScript
- **PostgreSQL** - Robust relational database
- **sqlc** - Compile-time type-safe SQL
- **golang-migrate** - Database migrations
- **mise** - Unified tool and task management

## Prerequisites

- [mise](https://mise.jdx.dev/) installed
- [Docker](https://www.docker.com/) for PostgreSQL

## Quick Start

### 1. Clone with gonew

```bash
gonew github.com/justestif/go-web-template github.com/yourname/myproject
cd myproject
```

### 2. Install Tools

```bash
mise install
```

This installs:

- Go (latest)
- templ (latest)
- sqlc (latest)
- golang-migrate (latest)
- air (latest)

### 3. Download Tailwind CSS + DaisyUI

Download the [Tailwind CSS standalone binary](https://github.com/tailwindlabs/tailwindcss/releases/latest) and the [DaisyUI bundle](https://github.com/saadeghi/daisyui/releases/latest) into the project root:

```bash
curl -sLo tailwindcss https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64
chmod +x tailwindcss
curl -sLO https://github.com/saadeghi/daisyui/releases/latest/download/daisyui.mjs
curl -sLO https://github.com/saadeghi/daisyui/releases/latest/download/daisyui-theme.mjs
```

> These files are gitignored. Each developer needs to download them once.
> For macOS, replace `tailwindcss-linux-x64` with `tailwindcss-macos-arm64` or `tailwindcss-macos-x64`.

### 4. Start PostgreSQL

```bash
docker-compose up -d
```

### 5. Setup Project

```bash
mise run setup
```

This will:

- Run database migrations
- Generate templ components
- Generate type-safe SQL code

### 6. Start Development

```bash
mise run dev
```

Air handles everything on each file change: builds CSS, regenerates templ files, and recompiles Go.

### 7. Visit Application

Open [http://localhost:3000](http://localhost:3000)

## Available Tasks

Run `mise tasks` to see all available tasks:

- `mise run dev` - Start development server with live reload (CSS + templ + Go)
- `mise run build` - Build production binary (CSS + templ + Go)
- `mise run db-migrate` - Run database migrations
- `mise run db-rollback` - Rollback last migration
- `mise run sqlc` - Generate type-safe SQL code
- `mise run setup` - Complete project setup

## Project Structure

```
.
├── cmd/web/              # Application entry point
├── components/           # Templ templates
├── internal/
│   ├── handlers/         # HTTP handlers
│   ├── middleware/        # Custom middleware
│   └── database/         # Database queries & connection
├── migrations/           # Database migrations
├── static/css/           # Generated CSS output (gitignored)
├── styles/               # CSS source
│   └── input.css         # Tailwind + DaisyUI entry point
├── mise.toml             # Tool & task configuration
└── docker-compose.yml    # PostgreSQL setup
```

## Development Notes

### CSS

Tailwind CSS is built by the standalone CLI (no Node.js required). The binary and generated CSS are gitignored — run the download step above once per machine. Air rebuilds CSS automatically on every file change.

To customize the theme, add a `@plugin "../daisyui-theme.mjs"` block to `styles/input.css`. See the [DaisyUI theme docs](https://daisyui.com/docs/themes/).

### CSRF

Forms use Gorilla CSRF middleware:

- Token field name: `gorilla.csrf.Token`
- Access in templates: `csrf.Token(r)`
- Automatically validated on POST/PUT/DELETE
- Set `secure=true` in production (HTTPS only)

### Database

Sample migration creates a `users` table. See `internal/database/queries.sql` for example queries.

## Production Deployment

1. Set environment variables:

   ```bash
   export DATABASE_URL="postgres://..."
   export CSRF_KEY="your-32-byte-secret-key"
   export PORT="8080"
   ```

2. Update CSRF middleware to use `secure=true` in `cmd/web/main.go`

3. Build production binary (generates CSS, templ, and Go binary):

   ```bash
   mise run build
   ```

4. Run migrations:

   ```bash
   mise run db-migrate
   ```

5. Start server:

   ```bash
   ./bin/app
   ```

## License

MIT
