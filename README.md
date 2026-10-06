# DentistChair MVP

This workspace now contains a simple full-stack MVP inspired by the project summary in newclone.md.

## What is included
- A Go backend with SQLite persistence
- Dentist registration, image upload, and camera-log endpoints
- A Python analysis script that produces a lightweight chair-height recommendation result
- A React + Vite frontend to register dentists, analyze images, and log camera events

## Run locally

### 1. Backend
```bash
cd backend
go mod tidy
go run .
```

### 2. Frontend
```bash
cd frontend
npm install
npm run dev
```

Open http://localhost:5173 to use the app.
