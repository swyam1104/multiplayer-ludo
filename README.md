# Multiplayer Ludo Backend MVP

This is a real-time, stateful Ludo game backend written in Go.

## Features
- Complete Ludo game rules engine (captures, safe cells, extra turns, specific home logic).
- WebSocket Gateway with JWT authentication.
- In-memory concurrent room state guarded by `sync.RWMutex`.
- MySQL persistence for Users and match metadata.

## Quick Start

### Prerequisites
- Go 1.22+
- MySQL running natively (Ensure you have a `ludo` database created and a `ludo_user` configured. See `start.sh` for exact connection strings).

### Running Locally
- **On Linux / macOS:**
  ```sh
  ./start.sh
  ```
- **On Windows:**
  1. Start MySQL daemon (if not running as a service):
     ```cmd
     start_mysql.bat
     ```
  2. Start the server:
     ```cmd
     run_server.bat
     ```
- Open `http://localhost:8080/` or `demo.html` in your browser.
- Open a second browser window to act as Player 2.
- Use the demo client to register users, create a room, connect the WebSockets, and play the game!

## Architecture
- **Game Engine**: A pure Go state machine. Entirely isolated from networking for robust unit testing.
- **WebSocket Hub**: A concurrent event loop that manages the active connections per room and relays messages between the clients and the Game Engine.
- **Room Manager**: In-memory source of truth for active game lobbies and their attached game engines.

## Roadmap & Next Steps
- Implement Matchmaking Queue
- Implement Redis cache for Multi-instance support
- Friend System, Chat, and Leaderboards
