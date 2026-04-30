# CaddyWeb

A modern web management interface for Caddy web server, featuring a Vue 3 frontend and Go/Gin API.

[![TypeScript](https://img.shields.io/badge/TypeScript-blue?style=flat-square&logo=typescript&logoColor=white)](https://www.typescriptlang.org)
[![Vue.js](https://img.shields.io/badge/Vue.js-3.x-42b883?style=flat-square&logo=vue.js&logoColor=white)](https://vuejs.org)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![Caddy](https://img.shields.io/badge/Caddy-v2-8f9554?style=flat-square)](https://caddyserver.com)

[Features](#features) • [Architecture](#architecture) • [Getting Started](#getting-started) • [Project Structure](#project-structure) • [API Documentation](#api-documentation)

## Features

- **Dashboard** - Real-time Caddy server monitoring and statistics
- **Site Management** - Configure static sites and reverse proxy endpoints
- **Domain Management** - Manage domains and DNS settings
- **TLS Certificates** - View and manage SSL/TLS certificates
- **Live Logs** - Real-time server log streaming via SSE
- **Settings** - Configure server preferences

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   CaddyWeb UI   │────▶│    CaddyAPI     │────▶│  Caddy Server   │
│  (Vue 3 + Vite) │     │   (Go + Gin)    │     │   (Config API)  │
│    Port 8080    │     │    Port 8081    │     │                 │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

- **CaddyWeb** - Vue 3 SPA with Element Plus UI and Pinia state management
- **CaddyAPI** - REST API built with Gin framework, manages Caddy configuration
- **Real-time** - Server-Sent Events (SSE) for live log streaming

## Getting Started

### Prerequisites

- Node.js 18+
- Go 1.21+
- Caddy server (for actual web serving)

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd CaddyWeb

# Install frontend dependencies
cd CaddyWeb && pnpm install

# Install backend dependencies
cd ../CaddyAPI && go mod download
```

### Development

Start the backend server first:

```bash
cd CaddyAPI && go run main.go
```

Then start the frontend in a new terminal:

```bash
cd CaddyWeb && pnpm dev
```

Access the application at `http://localhost:8080`. The frontend proxies `/api` requests to `http://localhost:8081`.

### Build

```bash
# Build frontend
cd CaddyWeb && pnpm build

# Build backend
cd CaddyAPI && go build
```

### Type Checking

```bash
# Frontend TypeScript
npx vue-tsc --noEmit

# Backend (compile check)
go build ./...
```

## Project Structure

```
CaddyWeb/                     # Frontend (Vue 3 + TypeScript)
├── src/
│   ├── api/                  # API client layer
│   ├── components/           # Reusable Vue components
│   ├── layout/               # App layout wrapper
│   ├── router/               # Vue Router configuration
│   ├── stores/               # Pinia state stores
│   ├── types/                # Centralized TypeScript types
│   └── views/                # Page components
│       ├── Dashboard.vue     # Dashboard & monitoring
│       ├── Domains.vue       # Domain management
│       ├── Logs.vue          # Live logs viewer
│       ├── Settings.vue      # Server settings
│       ├── Sites.vue         # Site configuration
│       ├── Setup.vue         # Initial setup
│       ├── TLS.vue           # TLS certificate management
│       └── Login.vue         # Authentication
├── package.json
└── vite.config.ts

CaddyAPI/                     # Backend (Go + Gin)
├── internal/
│   ├── handlers/             # HTTP handlers & routes
│   │   ├── sites.go          # Site CRUD operations
│   │   ├── domains.go        # Domain management
│   │   ├── certs.go          # Certificate handling
│   │   ├── settings.go       # Settings management
│   │   ├── sse.go            # Server-Sent Events
│   │   └── router.go         # Route definitions
│   ├── models/               # Data models
│   └── config/               # Configuration utilities
├── main.go
└── go.mod
```

## API Documentation

The API runs on port 8081 (or `PORT` env var). Base URL: `http://localhost:8081/api`

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/config` | Get full Caddy configuration |
| GET | `/sites` | List all configured sites |
| POST | `/sites` | Create a new site |
| PUT | `/sites/:id` | Update a site |
| DELETE | `/sites/:id` | Delete a site |
| GET | `/domains` | List all domains |
| POST | `/domains` | Add a new domain |
| DELETE | `/domains/:id` | Remove a domain |
| GET | `/certs` | List TLS certificates |
| GET | `/settings` | Get server settings |
| PUT | `/settings` | Update settings |
| GET | `/logs/stream` | SSE log stream |

### WebSocket (SSE)

Live log streaming available at `/api/logs/stream` using Server-Sent Events.

## Tech Stack

**Frontend**
- Vue 3 (Composition API)
- TypeScript
- Vite
- Element Plus
- Pinia
- Vue Router
- Axios

**Backend**
- Go 1.21+
- Gin
- gorilla/websocket
- Caddy configuration API