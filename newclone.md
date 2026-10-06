# DentistChair Project Summary

## Project Purpose
DentistChair is a full-stack intelligent surveillance and body-recognition system for dental clinic workflows. It allows users to:
- register dentists by uploading a full-body image and assigning a name
- analyze the uploaded image and estimate a recommended chair height
- receive simulated camera frame events and store interpretation logs
- inspect the registered dentists and recent analysis history from a simple dashboard

In short, this project combines a Go backend, a React frontend, and a lightweight AI-style analysis pipeline to create a monitoring and recognition workflow.

## Overall Architecture
The project is divided into three main parts:
1. a simple frontend web application for registering dentists and reviewing logs
2. a backend API server and SQLite database
3. a lightweight Python-based analysis pipeline that can later be upgraded to MediaPipe/OpenCV

## Project Structure

```text
Finalproject_dentist_ai_simulate/
├── backend/
│   ├── main.go
│   ├── go.mod
│   ├── go.sum
│   ├── model/
│   │   └── main_pipeline.py
│   ├── data/
│   ├── uploads/
│   └── README.md
├── frontend/
│   ├── src/
│   │   ├── App.jsx
│   │   ├── main.jsx
│   │   └── index.css
│   ├── package.json
│   ├── vite.config.js
│   └── index.html
├── newclone.md
└── README.md
```

## Main Components

### 1. Backend
The backend is built with Go and the standard net/http package.
It provides APIs for:
- dentist registration
- image analysis submission
- camera event logging
- retrieval of stored dentists and logs
It uses SQLite for storage.

### 2. Frontend
The frontend is built with React and Vite.
It provides a simple web dashboard for:
- registering dentists
- uploading images for analysis
- simulating live camera events
- reviewing recent logs

### 3. AI/Computer Vision Pipeline
The Python model script implements a lightweight analysis workflow that can be expanded into a richer computer-vision feature set later.
The current version estimates a chair height and produces a simple label/confidence summary.

## Typical Workflow
1. A user registers a dentist with a name and optional image.
2. The backend stores the dentist and associated chair height recommendation.
3. The user uploads an image for analysis.
4. The backend runs the analysis pipeline and stores the result in the log history.
5. Results are shown in the dashboard.

## Technology Stack
- Backend: Go, SQLite
- Frontend: React, Vite
- AI: Python, with a simple analysis stub that can evolve into MediaPipe/OpenCV

## Notes for Cloning or Reuse
- The backend uses a local SQLite file under backend/data.
- Uploaded media is stored under backend/uploads.

## Summary
This repository is a lightweight MVP for an AI-assisted dental chair monitoring workflow with a Go backend, a React frontend, and a simple analysis pipeline.
