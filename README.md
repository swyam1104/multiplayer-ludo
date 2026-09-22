# 🎲 Multiplayer Ludo Backend

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Database](https://img.shields.io/badge/Database-MySQL%208.4-4479A1?style=flat&logo=mysql)](https://www.mysql.com/)
[![Deployment](https://img.shields.io/badge/Deployed%20on-Railway-0B0D0E?style=flat&logo=railway)](https://railway.app)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A high-performance, real-time, stateful multiplayer Ludo game backend written in Go. Features an isolated deterministic game state engine, concurrent WebSocket communication with JWT authentication, and MySQL persistence.

---

## 🌐 Live Demo

The production application is live and accessible online:

🔗 **[https://multiplayer-ludo-production-fe0b.up.railway.app/](https://multiplayer-ludo-production-fe0b.up.railway.app/)**

> **How to test multiplayer live:**
> 1. Open the live URL in your browser.
> 2. Open an Incognito window (or a second browser tab) with the same link for Player 2.
> 3. Register and log in both players.
> 4. Player 1 creates a room, Player 2 joins with the generated **Room Code**.
> 5. Connect WebSockets, start the game, roll dice, and play!

---

## 🚀 Key Features

- **Stateful Game Engine**: Complete deterministic implementation of classic Ludo rules (captures, safe star tiles, spawn roll of 6, home column progression, and win conditions).
- **Concurrent Game Gateway**: Full-duplex WebSocket communication supporting real-time event broadcasting and synchronized turns.
- **Thread-Safe Architecture**: State synchronization guarded by granular `sync.RWMutex` read/write locks.
- **Authentication & Security**: Secure user registration and login with bcrypt password hashing and stateless JWT bearer tokens.
- **Persistence**: Relational MySQL schema for users, historical matches, and audit trails.
- **Dockerized Deployment**: Multi-stage lightweight Alpine Linux container optimized for minimal cloud footprint.

---

## 🏗️ Architecture

```
                      +-------------------+
                      |   Client (Web)    |
                      +---------+---------+
                                |
             +------------------+------------------+
             | REST API (HTTP)                     | WebSocket (WSS)
             v                                     v
    +-----------------+                   +-----------------+
    |   Auth & Room   |                   |  WebSocket Hub  |
    |   Controllers   |                   |   & Client      |
    +--------+--------+                   +--------+--------+
             |                                     |
             | MySQL                               | In-Memory Events
             v                                     v
    +-----------------+                   +-----------------+
    |  MySQL Storage  |                   |   Game Engine   |
    | (Users & Match) |                   | (State Machine) |
    +-----------------+                   +-----------------+
```

### Module Breakdown
* `internal/game`: Pure Go game state machine completely decoupled from networking. Fully unit tested.
* `internal/websocket`: Real-time connection management, message serialization, and room event broadcasting.
* `internal/room`: Lobby lifecycle (creation, joining, player capacity, status tracking).
* `internal/auth`: User registration, authentication, password verification, and JWT generation.
* `cmd/server/main.go`: Application entry point, router initialization, and database connection pooling.

---

## 📡 API Reference

### Authentication
| Method | Endpoint | Description | Auth Required |
|---|---|---|---|
| `POST` | `/api/auth/register` | Register a new user (`username`, `email`, `password`) | No |
| `POST` | `/api/auth/login` | Authenticate and obtain JWT token (`email`, `password`) | No |

### Room Management
| Method | Endpoint | Description | Auth Required |
|---|---|---|---|
| `POST` | `/api/rooms` | Create a new multiplayer room lobby | Bearer JWT |
| `POST` | `/api/rooms/{code}/join`| Join an existing room via 6-character room code | Bearer JWT |
| `GET`  | `/api/rooms/{code}` | Fetch current room lobby status and players | Bearer JWT |

### WebSocket Gateway
* **Endpoint**: `/ws?room={roomCode}&token={jwtToken}`
* **Protocols**: `ws://` (Local) / `wss://` (Production TLS)

#### WebSocket Events
| Event Action | Direction | Payload Example |
|---|---|---|
| `dice_roll` | Client ➡️ Server | `{"action": "dice_roll"}` |
| `move_token` | Client ➡️ Server | `{"action": "move_token", "token_id": 0}` |
| `start_game` | Client ➡️ Server | `{"action": "start_game"}` |
| `game_state` | Server ➡️ Client | Broadcasts updated board positions, active turn, and dice result |
| `player_joined`| Server ➡️ Client | Notifies room members of a new player |

---

## 🛠️ Local Development & Quick Start

### Prerequisites
* **Go**: Version `1.22+` (or `1.25+`)
* **MySQL**: Version `8.0+`
* **Docker & Docker Compose** *(Optional)*

### Option A: Running with Docker Compose (Recommended)
```bash
docker-compose up --build
```
This spins up both the MySQL database and the Go server container automatically on `http://localhost:8080`.

---

### Option B: Running Natively on Windows

1. **Start MySQL Server:**
   ```cmd
   start_mysql.bat
   ```
2. **Start the Go Application:**
   ```cmd
   run_server.bat
   ```
3. Open `http://localhost:8080` in your browser.

---

### Option C: Running Natively on Linux / macOS

1. **Run the startup script:**
   ```bash
   chmod +x start.sh
   ./start.sh
   ```

---

## ⚙️ Environment Variables

| Variable | Description | Default / Example |
|---|---|---|
| `PORT` | Listening HTTP/WS port | `8080` |
| `JWT_SECRET` | Secret key for signing authorization tokens | `supersecretjwtkey` |
| `DB_DSN` | MySQL database connection string | `root:password@tcp(127.0.0.1:3306)/ludo?parseTime=true` |

---

## 🧪 Testing

Execute the comprehensive unit test suite covering game rules, token capture mechanics, star/safe positions, and extra turns:

```bash
go test -v ./internal/game/...
```

---

## 🗺️ Roadmap
- [ ] Matchmaking Queue system
- [ ] Redis caching layer for distributed multi-node scaling
- [ ] In-game real-time chat and emoticons
- [ ] Global Leaderboard and Win/Loss player analytics
- [ ] Mobile Flutter client integration

---

## 📄 License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
