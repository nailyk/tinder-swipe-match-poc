# Tinder Swipe Match Simulator with Redis & Go

A proof-of-concept simulation of Tinder-style swiping and matching, built using Go and Redis with embedded Lua functions to ensure atomic and race-free operations. This project demonstrates how to leverage Redis server-side functions for consistency under concurrent workloads.

---

## 🚀 Features

- ✅ **Atomic swipe and match logic** using [Redis Lua functions](https://redis.io/docs/latest/develop/programmability/functions-intro/)
- 🧪 **Concurrent user simulation** with goroutines and a lock-safe random generator
- 🧱 **Correctness ensured** with Lua-level synchronization (no race conditions)
- 🐳 **Docker/Docker Compose** setup with Redis 7 Alpine and Go app
- ⚙️ **Configurable user and swipe volume**
- 🧪 **Unit tests** with validation of match logic and atomicity

---

## 📦 Prerequisites

- [Docker](https://docs.docker.com/get-docker) & Docker Compose (recommended for isolated setup)
- Go 1.23+ (for local build and test runs)
- Redis 7+ (required for `FUNCTION LOAD` if not using Docker)

---

## 🧰 Setup & Running Locally

1. **Start Redis**  
   If you're not using Docker, install and run Redis 7+ locally:

   ```bash
   redis-server
   ```

2. **Build & Run the Simulation**

   ```bash
   go mod tidy
   go build -o tinder
   ./tinder -redis localhost:6379 -users 100 -swipes 1000
   ```

   ### CLI Flags

   - `-redis` – Redis address (default: `localhost:6379`)
   - `-users` – Number of users to simulate (default: `10`)
   - `-swipes` – Total number of swipes to simulate (default: `100`)

---

## 🐳 Running with Docker Compose

To bring up both Redis and the simulation environment:

```bash
docker-compose up --build
```

This will:

- Start Redis 7 in one container
- Build and run the Go app in another
- Automatically run the simulation with default values

---

## 🧠 Redis Lua Function

The Lua script `swipe_and_check_match.lua` is loaded into Redis via `FUNCTION LOAD` and implements:

- ✅ Atomic swipe recording for user pairs
- 💘 Match detection when both users like each other
- 📥 Match persistence into `matches:<user>` Redis sets

**Key Benefits:**

- No need for external locking or retries
- Resilient under concurrent load (backed by `FCall` and `register_function`)

---

## 🧪 Testing

Tests are defined in `tinder_service_test.go`. They validate:

- ✅ Correct match detection
- ❌ No match on mismatched or one-sided swipes
- 🔄 Atomic behavior under concurrency

Run tests:

```bash
go test -v
```

Ensure Redis is running locally (`localhost:6379`) before executing tests.

---

## 📁 Project Structure

```
.
├── main.go                    # CLI entry point and service setup
├── swipe_simulation.go        # Swipe simulation logic with goroutines
├── tinder_service.go          # Redis interaction and match logic
├── swipe_and_check_match.lua  # Atomic Lua function registered in Redis
├── tinder_service_test.go     # Unit tests for swipe + match
├── Dockerfile                 # Go build container
├── docker-compose.yml         # Redis + app orchestration
└── README.md                  # This file
```

---

## 🙌 Inspiration

This project is inspired by the system design challenges explored in [HelloInterview's Tinder Architecture Breakdown](https://www.hellointerview.com/learn/system-design/problem-breakdowns/tinder).

---
