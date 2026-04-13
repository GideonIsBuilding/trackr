# Trackr — Job Application Tracker

A full-stack job application tracker with a Go backend, React frontend, and PostgreSQL database.

## Structure

```
job-tracker/
├── backend/       Go API server (chi router, pgx, JWT auth)
├── frontend/      React + TypeScript (Vite, TanStack Query)
├── migrations/    PostgreSQL migrations (golang-migrate)
└── README.md
```

## Quick start

### 1. Database
```bash
createdb job_tracker
cd backend && cp .env.example .env   # fill in DATABASE_URL, JWT_SECRET
make migrate-up
```

### 2. Backend
```bash
cd backend
go mod tidy
make run        # starts on :8080
```

### 3. Frontend
```bash
cd frontend
npm install
npm run dev     # starts on :5173, proxies /api → :8080
```

## API endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | /api/auth/register | Create account |
| POST | /api/auth/login | Sign in, get JWT |
| GET | /api/applications | List all applications |
| POST | /api/applications | Log new application |
| GET | /api/applications/:id | Get single application |
| PATCH | /api/applications/:id/status | Update status |
| GET | /api/applications/:id/history | Status timeline |
| PUT | /api/applications/:id/reminder | Configure reminder |
| POST | /api/applications/:id/reminder/snooze | Snooze reminder |
