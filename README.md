# Alcohol Label Verification App

AI-powered prototype for comparing alcohol label images against COLA-style application data. The frontend is a React app deployed on Vercel. The backend is a Go API deployed on Fly.io with Neon Postgres and Neon Auth.

## Table of contents

- [Approach](#approach)
  - [Workflow and data entry](#workflow-and-data-entry)
  - [Why these choices](#why-these-choices)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Local setup](#local-setup)
  - [1. Backend](#1-backend)
  - [2. Frontend](#2-frontend)
  - [3. Full local workflow](#3-full-local-workflow)
- [Deployment](#deployment)
  - [Frontend (Vercel)](#frontend-vercel)
  - [Backend (Fly.io)](#backend-flyio)
- [API endpoints](#api-endpoints)
- [Trade-offs and limitations](#trade-offs-and-limitations)
- [Project layout](#project-layout)
- [Assessment notes](#assessment-notes)

## Approach

This prototype separates **AI extraction** from **compliance decisions**:

1. **Gemini Flash** reads the uploaded label image and returns structured JSON fields plus evidence snippets and confidence scores.
2. **Deterministic Go rules** compare extracted values to the application data and assign `pass`, `needs_review`, or `fail`.
3. Results are stored in Neon Postgres and shown to the reviewer as a plain checklist.

That split keeps the tool auditable. The model helps with OCR-like reading of imperfect photos, but the final status comes from explicit rules the reviewer can understand.

### Workflow and data entry

In the real TTB process described in `[INTERVIEW_NOTES.md](INTERVIEW_NOTES.md)`, a **company or importer** submits a COLA application (brand, ABV, warning text, and so on) together with label artwork. A **compliance agent** then checks that what appears on the label matches what was submitted—not re-entering the whole application from scratch.

This prototype models the **agent verification step** only:


|                | Application data                                                                  | Label image                            |
| -------------- | --------------------------------------------------------------------------------- | -------------------------------------- |
| **This POC**   | Entered manually in the form                                                      | Uploaded by the reviewer               |
| **Production** | Loaded from the submission service (e.g. COLA or an internal case-management API) | Attached to the same submission record |


There is no company-facing submission flow and no COLA integration in this build. Manual entry is a stand-in so reviewers can paste or type sample application fields and upload a label image to see the checklist. In a deployed product, the verify screen would receive `application` data from the upstream service; the UI would focus on image upload (if needed), running verification, and reviewing results.

### Why these choices


| Area     | Choice                                        | Why                                                                              |
| -------- | --------------------------------------------- | -------------------------------------------------------------------------------- |
| AI       | Gemini Flash via backend REST API             | Free tier, fast multimodal JSON extraction, no OCR server in the Docker image    |
| Rules    | Go code, not the LLM                          | Matches assessment needs: exact warning text, ABV tolerance, brand normalization |
| Auth     | Neon Auth in frontend, verification in Go API | Backend owns trust boundaries and database access                                |
| API      | `net/http`, `slog`, `database/sql` + pgx      | Small dependency surface, idiomatic Go                                           |
| DI       | `go.uber.org/dig`                             | Keeps wiring in `cmd/api` without a heavier framework                            |
| Frontend | Vite, TanStack Router/Query/Form, shadcn/ui   | Typed routing, form validation, predictable internal-tool UI                     |
| Deploy   | Vercel + Fly.io                               | Static frontend and containerized API with CORS between them                     |


## Architecture

```text
Browser (Vercel)
  -> Neon Auth sign-in
  -> Go API (Fly.io) with Bearer token
      -> JWKS verification
      -> Gemini Flash extraction
      -> Rules engine
      -> Neon Postgres persistence
```

## Prerequisites

- Node.js `24.16.0`
- npm `11.13.0`
- Go `1.26.4`
- Neon project with Postgres and Neon Auth enabled
- Gemini API key (or use `AI_PROVIDER=fake` locally)
- Vercel CLI and Fly CLI for deployment

## Local setup

### 1. Backend

```bash
cd backend
cp .env.example .env
```

Recommended local `.env` values:

```env
DATABASE_URL=postgres://...
NEON_AUTH_JWKS_URL=https://<project>.neonauth.<region>.aws.neon.tech/<db>/auth/.well-known/jwks.json
AI_PROVIDER=fake
SKIP_AUTH_IN_DEV=true
ALLOWED_ORIGINS=http://localhost:5173
```

Load backend environment variables into your current shell:

```bash
source backend/scripts/load-env.sh
```

From the backend directory:

```bash
source scripts/load-env.sh
```

Alternative one-liner:

```bash
eval "$(backend/scripts/load-env.sh --export)"
```

Run the API:

```bash
cd backend
go run ./cmd/api
```

Run backend tests:

```bash
cd backend
go test ./...
```

Build the backend binary:

```bash
cd backend
go build -o bin/api ./cmd/api
```

### 2. Frontend

```bash
cd frontend
cp .env.example .env.local
npm install
```

Load frontend environment variables into your current shell:

```bash
source frontend/scripts/load-env.sh
```

From the frontend directory:

```bash
source scripts/load-env.sh
```

Load both backend and frontend env vars at once from the repo root:

```bash
source scripts/load-env.sh
```

Then start Vite (it will also read `.env` / `.env.local` from disk when npm runs):

```bash
npm run dev
```

Recommended local `.env.local`:

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_DEMO_MODE=true
```

Frontend commands:

```bash
cd frontend
npm run lint
npm run format
npm run test
npm run build
```

### 3. Full local workflow

1. Start the Go API on port `8080`.
2. Start the Vite dev server on port `5173`.
3. Open `http://localhost:5173/login` and continue in demo mode.
4. Use **Verify** to submit a label image plus application fields (entered manually in this POC; see [Workflow and data entry](#workflow-and-data-entry)).
5. Review the checklist. Failures are listed first.

With `AI_PROVIDER=fake`, the backend returns predictable sample extraction data without calling Gemini.

## Deployment

### Frontend (Vercel)

Deploy the `frontend` directory as a Vite project.

Set environment variables:

- `VITE_API_BASE_URL=https://<your-fly-app>.fly.dev`
- `VITE_NEON_AUTH_URL=https://<project>.neonauth.<region>.aws.neon.tech/<db>/auth`
- `VITE_DEMO_MODE=false`

### Backend (Fly.io)

From the repository root:

```bash
fly launch --no-deploy
fly secrets set DATABASE_URL=... NEON_AUTH_JWKS_URL=... GEMINI_API_KEY=... ALLOWED_ORIGINS=https://<your-vercel-app>.vercel.app
fly deploy
```

The root `[fly.toml](fly.toml)` builds `[backend/Dockerfile](backend/Dockerfile)` with a multi-stage Alpine image (`CGO_ENABLED=0`, non-root user).

## API endpoints

- `GET /healthz`
- `GET /api/v1/me`
- `POST /api/v1/verifications` — image + `application` JSON field
- `GET /api/v1/verifications`
- `GET /api/v1/verifications/{id}`

## Trade-offs and limitations

- **5-second target**: Gemini Flash is used for speed, but network latency and image size still matter.
- **Government warning boldness**: The prototype checks required wording and `GOVERNMENT WARNING:` casing. Bold formatting is routed to `needs_review` when confidence is low.
- **Free tier**: Gemini free quotas are suitable for demos and assessment usage, not high-volume production.
- **No COLA integration**: This is a standalone proof of concept, not connected to TTB systems. Application data is typed in for demos; production would pull it from the submission service.
- **Demo mode**: `VITE_DEMO_MODE=true` and `SKIP_AUTH_IN_DEV=true` simplify local testing but must be disabled in production.

## Project layout

```text
frontend/   React + Vite UI
backend/    Go API, rules engine, Gemini extractor, Postgres store
fly.toml    Fly.io deployment config
INTERVIEW_NOTES.md
```

## Assessment notes

See `[INTERVIEW_NOTES.md](INTERVIEW_NOTES.md)` for stakeholder context from the take-home brief.