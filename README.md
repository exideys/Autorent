<div align="center">
  <h1>AutoRent</h1>
  <p>
    A full-stack car rental platform for browsing vehicles, creating bookings,
    managing rental orders, and running admin fleet operations.
  </p>

  <p>
    <img alt="React" src="https://img.shields.io/badge/React-20232A?style=for-the-badge&logo=react&logoColor=61DAFB" />
    <img alt="TypeScript" src="https://img.shields.io/badge/TypeScript-3178C6?style=for-the-badge&logo=typescript&logoColor=white" />
    <img alt="Vite" src="https://img.shields.io/badge/Vite-646CFF?style=for-the-badge&logo=vite&logoColor=white" />
    <img alt="Tailwind CSS" src="https://img.shields.io/badge/Tailwind_CSS-06B6D4?style=for-the-badge&logo=tailwindcss&logoColor=white" />
    <img alt="Go" src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" />
    <img alt="Gin" src="https://img.shields.io/badge/Gin-008ECF?style=for-the-badge&logo=gin&logoColor=white" />
    <img alt="MySQL" src="https://img.shields.io/badge/MySQL-4479A1?style=for-the-badge&logo=mysql&logoColor=white" />
    <img alt="Docker" src="https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white" />
    <img alt="Vitest" src="https://img.shields.io/badge/Vitest-6E9F18?style=for-the-badge&logo=vitest&logoColor=white" />
    <img alt="GitHub Actions" src="https://img.shields.io/badge/GitHub_Actions-2088FF?style=for-the-badge&logo=githubactions&logoColor=white" />
  </p>
</div>

## Overview

AutoRent combines a React/Vite frontend with a Go/Gin backend to deliver a complete rental workflow: public vehicle browsing, authenticated booking, user profiles, admin fleet management, news publishing, realtime support chat, AI-assisted vehicle search, and Ukrainian translation support.

The app is designed to run locally during development, ship through Docker, and publish the static frontend through GitHub Pages when configured.

## Features

| Area | What it includes |
| --- | --- |
| Showroom | Public vehicle listings, vehicle details, image browsing, and booking actions. |
| Booking | Pickup/return validation, date/time checks, and protected rental order creation. |
| Accounts | Email/password auth, optional Google sign-in, profile editing, and rental history. |
| Admin | Fleet CRUD, image uploads, customer ratings, rental order visibility, news tools, and support tools. |
| Support | Realtime chat, attachments, Enter-to-send, multiline messages, and admin/customer conversations. |
| Content | Public news page with admin-managed published articles. |
| AI and i18n | Gemini-powered car recommendations and optional DeepL Ukrainian translations. |
| Quality | Go tests, Vitest component tests, linting, build checks, CodeQL, and dependency review. |


## Tech Stack

| Layer | Tools |
| --- | --- |
| Frontend | React, TypeScript, Vite, Tailwind CSS, Framer Motion, lucide-react |
| Backend | Go, Gin, MySQL driver, JWT |
| Database | MySQL-compatible database with TiDB-style TLS defaults |
| Integrations | Google OAuth, Google Drive storage, Gemini, DeepL |
| Testing | Go test, Vitest, React Testing Library, jsdom |
| Delivery | Docker, Docker Compose, Nginx, GitHub Actions, GitHub Pages |

## Project Structure

```text
.
|- Backend/
|  |- internal/          # API handlers, services, repositories, auth, storage
|  |- migrations/        # SQL schema migrations
|  |- Dockerfile
|  `- main.go
|- Frontend/
|  |- src/               # React app, components, data, API client, tests
|  |- Dockerfile
|  |- nginx.conf
|  `- vite.config.js
|- docs/screenshots/     # Public README screenshots
|- .github/workflows/    # CI, CodeQL, dependency review, Pages deploy
`- compose.yml
```

## Environment Variables

Create a root `.env` file for Docker Compose and backend runtime settings. The values below are examples only.

```env
# Backend server
PORT=8080
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000

# Database
DB_USER=root
DB_PASSWORD=
DB_HOST=localhost
DB_PORT=4000
DB_NAME=autorent
DB_TLS=tidb

# Auth
JWT_SECRET=change-this-secret
JWT_TOKEN_TTL=24h
ADMIN_SETUP_TOKEN=change-this-admin-setup-token

# Optional AI and translation
GEMINI_API_KEY=
GEMINI_MODEL=gemini-2.5-flash
DEEPL_API_KEY=
DEEPL_API_URL=https://api-free.deepl.com

# Optional Google auth
GOOGLE_AUTH_CLIENT_ID=
VITE_GOOGLE_AUTH_CLIENT_ID=

# Optional Google Drive storage
GOOGLE_DRIVE_OAUTH_CLIENT_ID=
GOOGLE_DRIVE_OAUTH_CLIENT_SECRET=
GOOGLE_DRIVE_OAUTH_REFRESH_TOKEN=
GOOGLE_DRIVE_CARS_FOLDER_ID=
GOOGLE_DRIVE_NEWS_FOLDER_ID=
GOOGLE_DRIVE_SUPPORT_FOLDER_ID=

# Upload limits
IMAGE_UPLOAD_MAX_BYTES=10485760
SUPPORT_ATTACHMENT_MAX_BYTES=10485760

# Frontend build/runtime
VITE_API_BASE_URL=http://localhost:8080
VITE_BASE_PATH=/
FRONTEND_PORT=3000
```

Security notes:

- Do not commit real API keys, JWT secrets, OAuth client secrets, database passwords, refresh tokens, or admin setup tokens.
- `.env`, `Backend/.env`, `Frontend/.env`, credential folders, service account JSON files, and Google Drive credential JSON files are ignored by git.
- `DB_TLS=tidb` is the backend default. Use `DB_TLS=false` for a local MySQL server without TLS.
- `GEMINI_API_KEY`, `DEEPL_API_KEY`, Google auth, and Google Drive settings are optional. The app disables those integrations when they are missing.

## Database Setup

Apply migrations in order against your MySQL-compatible database:

```bash
mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p "$DB_NAME" < Backend/migrations/001_create_users.sql
mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p "$DB_NAME" < Backend/migrations/002_create_cars_and_images.sql
mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p "$DB_NAME" < Backend/migrations/003_create_rental_orders.sql
mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p "$DB_NAME" < Backend/migrations/004_create_news.sql
mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p "$DB_NAME" < Backend/migrations/005_create_support_messages.sql
```

## Local Development

Start the backend:

```bash
cd Backend
go mod download
go run .
```

Start the frontend in another terminal:

```bash
cd Frontend
npm ci
npm run dev
```

The Vite dev server proxies `/api` requests to `http://localhost:8080`.

## Docker

Run both services with Docker Compose:

```bash
docker compose up --build
```

The frontend is served by Nginx on `http://localhost:3000` by default, and the backend listens on `http://localhost:8080`.

This Compose file expects an external MySQL-compatible database configured through `.env`; it does not start a database container.

## Testing and Checks

Backend:

```bash
cd Backend
go test ./...
```

Frontend:

```bash
cd Frontend
npm run test
npm run lint
npm run build
```

On Windows PowerShell, if `npm` is blocked by execution policy, use `npm.cmd`:

```powershell
npm.cmd run test
npm.cmd run build
```

## API Overview

| Area | Endpoint |
| --- | --- |
| Health | `GET /health` |
| Auth | `/api/auth` |
| Cars | `/api/cars` |
| Rental orders | `/api/rental-orders` |
| News | `/api/news` |
| Support | `/api/support` |
| AI recommendation | `/api/ai/car-recommendation` |
| Translation | `/api/translate` |
| Admin | `/api/admin` |

Admin APIs require an admin JWT.

## Deployment Notes

- The frontend Docker image builds static assets and serves them through Nginx.
- The backend Docker image runs the compiled Go API on port `8080`.
- GitHub Actions include CI checks, Docker build smoke tests, CodeQL, dependency review, and GitHub Pages deployment for the frontend.
- For GitHub Pages deployments, configure `VITE_API_BASE_URL` so the static frontend can reach the deployed backend API.
