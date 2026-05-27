# Tattoo Consultation Website

AI-powered tattoo design consultation service. Clients upload body photos + design ideas, receive 3 AI-generated tattoo designs with professional art direction.

**Stack:** Go + Chi + PostgreSQL | React + TanStack + Tailwind v4 | Claude Vision + Flux

## Architecture

```
Client uploads body photo + idea
  → Claude Vision analyzes body part, skin tone
  → 3 prompts crafted (vary style/composition)
  → Together AI Flux Schnell (3 sketches)
  → Replicate Flux 1.1 Pro (3 finals)
  → Admin review → Client receives results
```

## Quick Start

```bash
docker compose up -d
# Backend: http://localhost:5200
# Frontend dev: cd frontend && npm run dev  (port 5201)
```

## Project Structure

```
backend/          Go Chi REST API
  cmd/server/     Entry point
  internal/       Auth, DB, handlers, services
  migrations/     PostgreSQL DDL
frontend/         React + TanStack SPA
  src/routes/     TanStack Router pages
  src/components/ Reusable UI components
```
