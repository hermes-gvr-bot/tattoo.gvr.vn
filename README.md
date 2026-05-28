# tattoo.gvr.vn — AI Tattoo Design Consultation

Dịch vụ tư vấn thiết kế hình xăm bằng AI. Khách upload ảnh cơ thể + ý tưởng → AI phân tích và sinh 3 phương án thiết kế.

**Live:** [https://tattoo.gvr.vn](https://tattoo.gvr.vn)

## Stack

| Layer | Tech |
|-------|------|
| Backend | Go + Chi Router + PostgreSQL |
| Frontend | React + TanStack Router + Query + Tailwind v4 |
| AI Vision | Claude Vision (OpenRouter) — body/skin/tattoo forensic analysis |
| AI Sketch | Together AI Flux Schnell ($0.001/ảnh) |
| AI Final | Replicate Flux 1.1 Pro ($0.06/ảnh) |
| AI LoRA | Replicate Flux Dev LoRA Trainer — body preview |
| Storage | MinIO S3 (self-hosted) |
| Deploy | Docker Compose + Traefik + Cloudflare SSL |

## Features

### 🎨 Hai chế độ tư vấn

| Mode | Input | Output |
|------|-------|--------|
| **New Tattoo** | Ảnh body part + ý tưởng | 3 thiết kế mới (Faithful / Bold / Artistic) |
| **Makeup Enhance** | Ảnh hình xăm hiện tại + mong muốn | 3 phương án: Touch-up / Enhance / Cover-up |

### 🧠 AI Pipeline (tự động, không cần admin trigger)

```
Upload ảnh → Claude Vision forensic analysis
           → Craft 3 prompts (vary style/composition)
           → Together AI sketch × 3 (3-5s mỗi ảnh)
           → Replicate Flux 1.1 Pro final × 3 (7s mỗi ảnh)
           → LoRA training từ ảnh khách (5-10 phút)
           → Body preview: thiết kế "dán" lên cơ thể khách
```

### 🔬 Forensic Analysis (Makeup mode)

Phân tích 8 điểm: tattoo style, line quality (có rating), hex color palette, composition, condition (aging/scarring), problem areas, cover-up difficulty (1-10), full forensic report.

### 🖼️ Body Preview (LoRA)

Train LoRA từ nhiều ảnh cơ thể khách → sinh thiết kế trực tiếp lên cơ thể, cho khách thấy trước hình xăm trông như thế nào trên người mình.

### 👤 Auth & Roles

- JWT auth (GoTrue-compatible)
- Client: tạo consultation, xem kết quả, gửi feedback
- Admin: dashboard, review, stats

## Quick Start

```bash
# Dev (server-local)
cd backend
DATABASE_URL=postgres://tattoo:tattoo_dev@localhost:5437/tattoo_consultation?sslmode=disable \
OPENROUTER_API_KEY=sk-or-... TOGETHER_API_KEY=... REPLICATE_API_KEY=r8_... \
S3_ENDPOINT=localhost:9010 S3_ACCESS_KEY=tattoo-backend S3_SECRET_KEY=tattoo-backend-secret \
S3_BUCKET=tattoo-consultation S3_PUBLIC_URL=http://100.86.223.10:9010 \
go run ./cmd/server

# Frontend
cd frontend && npm run dev  # → http://localhost:5201
```

## Production Deploy

```bash
# Build backend binary
cd backend && CGO_ENABLED=0 GOOS=linux go build -o ../deploy/server ./cmd/server

# Build frontend
cd frontend && npm run build  # → dist/

# Deploy to VPS
scp deploy/server deploy/docker-compose.yml deploy/nginx.conf root@tmds-vps-prod.tail3840e.ts.net:/root/domains/tattoo.gvr.vn/
scp -r dist/ root@tmds-vps-prod.tail3840e.ts.net:/root/domains/tattoo.gvr.vn/

# Start on VPS
ssh root@tmds-vps-prod.tail3840e.ts.net "cd /root/domains/tattoo.gvr.vn && docker compose up -d"

# Apply DB migrations (if new)
cat backend/migrations/003_add_lora.sql | ssh root@tmds-vps-prod.tail3840e.ts.net "docker exec -i tattoo-postgres psql -U tattoo -d tattoo_consultation"
```

## Project Structure

```
backend/
  cmd/server/main.go          Entry point
  internal/
    auth/jwt.go               JWT verify + claims
    config/config.go           Env config loader
    db/postgres.go             Connection + migrations
    handler/
      auth.go                 Register, login, me
      consultation.go         CRUD + auto-trigger AI
      generate.go             Admin manual trigger (fallback)
      photo.go                Multi-photo upload for LoRA
    middleware/auth.go         JWT middleware
    service/
      generator.go            AI pipeline orchestrator
      vision.go               Claude Vision analysis
      prompt.go               Variant prompt crafting
      lora.go                 Replicate LoRA training
    storage/s3.go              MinIO S3 client
  migrations/
    001_init.sql              Base schema
    002_add_consultation_type.sql  Dual consultation modes
    003_add_lora.sql          LoRA + multi-photo support
frontend/
  src/routes/                 TanStack Router pages
  src/components/             Reusable UI (Tailscale dark theme)
deploy/
  docker-compose.yml          Prod stack (PG + backend + nginx + Traefik)
  nginx.conf                  Reverse proxy + SPA fallback
```

## Database Schema

```
users             — Auth (email, bcrypt hash, role)
consultations     — Body photo, idea, type, status, analysis
variants          — 3 per consultation (sketch + final + body_preview)
loras             — LoRA training tracker (pending→training→ready)
consultation_photos — Additional reference photos for LoRA
feedbacks         — Client revision requests
payments          — Crypto payment tracking
```

## API

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/auth/register` | ✗ | Register |
| POST | `/api/auth/login` | ✗ | Login → JWT |
| GET | `/api/auth/me` | ✓ | Current user |
| POST | `/api/consultations` | ✓ | Create (auto-triggers AI) |
| GET | `/api/consultations` | ✓ | List my consultations |
| GET | `/api/consultations/{id}` | ✓ | Detail + variants |
| POST | `/api/consultations/{id}/photos` | ✓ | Upload extra photos (LoRA) |
| GET | `/api/consultations/{id}/lora-status` | ✓ | LoRA training progress |
| GET | `/api/admin/consultations` | admin | All consultations |
| GET | `/api/admin/stats` | admin | Dashboard stats |
| POST | `/api/admin/consultations/{id}/generate` | admin | Manual retrigger |
| GET | `/health` | ✗ | Health check |

## AI Providers

| Provider | Model | Use | Price |
|----------|-------|-----|-------|
| OpenRouter | Claude Sonnet 4 | Vision + prompt crafting | ~$0.005/req |
| Together AI | FLUX.1-schnell | Sketch (fast) | $0.001/img |
| Replicate | Flux 1.1 Pro | Final (quality) | $0.06/img |
| Replicate | Flux Dev LoRA Trainer | Body LoRA | ~$0.50/train |

**Monthly estimate (300 consultations):** ~$20-30 (sketches + finals + vision)

## Migrations on Deploy

⚠️ **DB migrations are NOT auto-applied on VPS.** After any deploy with new SQL migrations, run manually:

```bash
ssh root@tmds-vps-prod.tail3840e.ts.net "docker exec -i tattoo-postgres psql -U tattoo -d tattoo_consultation" < backend/migrations/XXX.sql
```

## Dev Notes

- Dev server: `tmds-server-local.tail3840e.ts.net:5201`
- Prod VPS: `tmds-vps-prod.tail3840e.ts.net`
- MinIO: `100.86.223.10:9010` (browser: `tat`, creds in env)
- CORS must include Tailscale IP + FQDN origins
- S3_PUBLIC_URL must be set to public IP (not localhost)
- Background goroutines must use `context.Background()`, not `r.Context()`
