# Multiplayer Ludo Backend MVP

This is a real-time, stateful Ludo game backend written in Go.

## Features
- Complete Ludo game rules engine (captures, safe cells, extra turns, specific home logic).
- WebSocket Gateway with JWT authentication.
- In-memory concurrent room state guarded by `sync.RWMutex`.
- MySQL persistence for Users and match metadata.

## Quick Start
1. Start the DB and Backend:
   ```sh
   docker compose up --build
   ```
2. Open `demo.html` in your browser.
3. Use the demo client to register a user, create a room, connect the WebSocket, and start a game.

## Architecture
- **Game Engine**: A pure Go state machine. Entirely isolated from networking for robust unit testing.
- **WebSocket Hub**: A concurrent event loop that manages the active connections per room and relays messages between the clients and the Game Engine.
- **Room Manager**: In-memory source of truth for active game lobbies and their attached game engines.

## Roadmap & Next Steps
- Implement Matchmaking Queue
- Implement Redis cache for Multi-instance support
- Friend System, Chat, and Leaderboards
