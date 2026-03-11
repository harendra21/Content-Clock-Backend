# Content Clock Backend

Backend service for Content Clock.  
Built on Go + PocketBase with custom REST routes for social OAuth, scheduled publishing, and analytics.

## Stack

- Go
- PocketBase
- Goth/OAuth integrations
- Cron jobs (inside PocketBase app lifecycle)

## Prerequisites

- Go `1.24+`
- A `.env` file in project root

## Setup

```bash
go mod download
```

## Environment Variables

Create `.env` in this directory.

```env
# General
API_HOST="http://localhost:8080"
REDIRECT_HOST="http://localhost:4200"
JWT_KEY="change-me"

# Facebook + Instagram
FACEBOOK_APP_ID=""
FACEBOOK_SECRET=""

# Twitter/X
TWITTER_KEY=""
TWITTER_SECRET=""
# Optional explicit gotwi keys (if omitted, backend uses TWITTER_KEY/TWITTER_SECRET)
GOTWI_API_KEY=""
GOTWI_API_KEY_SECRET=""

# LinkedIn
LINKEDIN_APP_ID=""
LINKEDIN_SECRET=""

# Pinterest
PINTEREST_APP_ID=""
PINTEREST_SECRET=""

# Mastodon
MASTODON_CLIENT_KEY=""
MASTODON_CLIENT_SECRET=""
MASTODON_BASE_URL="https://mastodon.social"

# Reddit
REDDIT_CLIENT_ID=""
REDDIT_SECRET=""

# Threads
THREADS_APP_ID=""
THREADS_SECRET_KEY=""

# AI (used by /api/v1/post-with-ai; key currently sent to OpenRouter)
CHAT_GPT_KEY=""
OPENROUTER_APP_NAME="Content Clock"
OPENROUTER_SITE_URL="http://localhost:4200"
```

Optional DB flag used in code:

```env
DB_MIGRATE="false"

# Set true once when you want backend to create/update required PocketBase collections.
# Keep false for normal runtime after schema is in place.
```

## Run Locally

```bash
go run . serve --http=0.0.0.0:8080
```

After first run:

- API base: `http://localhost:8080/api/`
- Admin UI: `http://localhost:8080/_/`
- Create initial PocketBase superuser from the Admin UI install link.

## Runtime Behavior

- Custom routes are mounted under `/api/v1/*` (OAuth start/callback, add connections, AI helper).
- Scheduled publisher cron runs every minute.
- Analytics fetch cron runs every 3 hours.
- Root `/` redirects to frontend.

## Docker

Build and run:

```bash
docker build -t content-clock-backend .
docker run -p 8080:8080 --env-file .env content-clock-backend
```

Or with compose:

```bash
docker compose up --build
```

Note:

- `docker-compose.yml` maps `./pbData:/pb/pb_data` for PocketBase data persistence.

## Frontend Pairing

Frontend should point to this backend URL in its environment config:

- `v1Api`: `http://localhost:8080/api/v1`
- `apiHost`: `http://localhost:8080/api`
- `pocketbaseUrl`: `http://localhost:8080`

## Deploy (Fly.io)

This repo includes:

- `fly.toml` for Fly app/runtime config

### One-time setup

```bash
fly auth login
fly launch --no-deploy
fly volumes create pb_data --size 3 --region bom
```

Enable Fly GitHub integration (Auto Deploy) from Fly dashboard for this app/repo.

### Fly secrets required

Set all app env secrets in Fly (`fly secrets set ...`), and at minimum:

- `API_HOST` = your Fly URL (for example `https://content-clock-backend.fly.dev`)
- `REDIRECT_HOST` = `https://content-clock.vercel.app`
- `OPENROUTER_SITE_URL` = `https://content-clock.vercel.app`

### OAuth callback base

Use:

`https://<your-fly-app>.fly.dev/api/v1/auth/<provider>/callback`
