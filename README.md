# URL Shortener

A production-style URL shortening service built in **Go**, using **MongoDB** for storage, **JWT** for authentication, and **Prometheus** for metrics.

Anonymous users can shorten URLs with random aliases. Registered users get custom aliases, ownership of their links, and access to per-link analytics via Prometheus.

---

## Table of Contents

- [Features](#features)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Architecture](#architecture)
- [Setup & Run](#setup--run)
- [API Reference](#api-reference)
- [Data Models](#data-models)
- [Authentication Flow](#authentication-flow)
- [Metrics](#metrics)
- [Code Walkthrough](#code-walkthrough)
- [Testing the API](#testing-the-api)

---

## Features

- Shorten any URL into a 6-character random alias
- Optional JWT-based user accounts
- Custom aliases for registered users
- Owner-only update / delete of shortened URLs
- Paginated listing of a user's URLs
- Click metrics exposed in Prometheus format (per-alias, per-referrer, per-user-agent)
- TTL support for expiring links
- Layered architecture (handler → service → repository) for clean separation of concerns

---

## Tech Stack

| Layer            | Library / Tool                                    |
|------------------|---------------------------------------------------|
| Language         | Go 1.22+                                          |
| HTTP framework   | [Gin](https://github.com/gin-gonic/gin)           |
| Database         | MongoDB (via official Go driver)                  |
| Authentication   | JWT via [golang-jwt/jwt](https://github.com/golang-jwt/jwt) |
| Password hashing | bcrypt (`golang.org/x/crypto`)                    |
| Metrics          | [prometheus/client_golang](https://github.com/prometheus/client_golang) |
| Env config       | [godotenv](https://github.com/joho/godotenv)      |

---

## Project Structure

```
url-shortener/
├── cmd/
│   └── api/
│       └── main.go              # entry point — wires everything together
├── config/
│   └── config.go                # loads .env into typed Config struct
├── internal/
│   ├── handler/                 # HTTP layer — receives requests, sends responses
│   │   ├── auth_handler.go
│   │   └── url_handler.go
│   ├── service/                 # business logic
│   │   ├── auth_service.go
│   │   └── url_service.go
│   ├── repository/              # database access layer
│   │   ├── user_repository.go
│   │   └── url_repository.go
│   ├── model/                   # data structs
│   │   ├── user.go
│   │   └── url.go
│   ├── middleware/              # JWT auth middleware
│   │   └── auth.go
│   └── metrics/                 # Prometheus metric definitions
│       ├── metrics.go
│       └── helpers.go
├── .env                         # environment variables (not committed)
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

### Why this layout?

The project follows the **standard Go layout**:

- `cmd/` — entry points. If you add a CLI tool later, it goes here too.
- `internal/` — Go enforces that nothing outside this module can import packages under `internal/`. This keeps your business logic private by default.
- **Layered architecture** inside `internal/` — handler → service → repository. Each layer only talks to the one directly below it. This makes the code testable, swappable, and easy to reason about.

---

## Architecture

```
Client
  │
  ↓ HTTP request
┌─────────────────────────────────────────────────────────────┐
│ Gin Router (cmd/api/main.go)                                │
│  ├── Public routes      → auth                              │
│  ├── OptionalAuth group → /urls (POST, GET)                 │
│  ├── RequireAuth group  → /users/:id/urls                   │
│  └── /metrics           → Prometheus scrape endpoint        │
└─────────────────────────────────────────────────────────────┘
  │
  ↓
┌─────────────────────────────────────────────────────────────┐
│ Middleware (internal/middleware)                            │
│  ├── RequireAuth   — rejects without valid JWT              │
│  └── OptionalAuth  — extracts user_id if token present      │
└─────────────────────────────────────────────────────────────┘
  │
  ↓
┌─────────────────────────────────────────────────────────────┐
│ Handler (internal/handler)                                  │
│   • Parse + validate request                                │
│   • Read user_id from context (set by middleware)           │
│   • Call service                                            │
│   • Format JSON response                                    │
└─────────────────────────────────────────────────────────────┘
  │
  ↓
┌─────────────────────────────────────────────────────────────┐
│ Service (internal/service)                                  │
│   • Business logic (alias generation, bcrypt, JWT signing)  │
│   • Coordinates multiple repository calls if needed         │
└─────────────────────────────────────────────────────────────┘
  │
  ↓
┌─────────────────────────────────────────────────────────────┐
│ Repository (internal/repository)                            │
│   • Only layer that talks to MongoDB                        │
│   • CRUD operations on collections                          │
└─────────────────────────────────────────────────────────────┘
  │
  ↓
MongoDB
```

Meanwhile, the redirect handler updates Prometheus counters in memory. Prometheus scrapes `/metrics` every 15s to collect those values.

---
