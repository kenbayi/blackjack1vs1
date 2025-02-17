<img src="./docs/logo.png" alt="Logo" width=200> 

# Dueljack

Dueljack is an online multiplayer card game built using **Go** with **PostgreSQL** for data storage, **Redis** for managing game sessions, and **WebSockets** for real-time player interactions.

## Features
- User authentication and session management.
- Real-time multiplayer gameplay using WebSockets.
- Game rooms with betting system.
- Database management with PostgreSQL.
- Redis for storing active game sessions.

## Installation
### Requirements
- Go 1.20+
- PostgreSQL
- Redis

### Steps
1. Clone the repository:
   ```sh
   git clone https://github.com/your-repo/dueljack.git
   cd dueljack
   ```
2. Install dependencies:
   ```sh
   go mod tidy
   ```
3. Configure database settings in `config/`.
4. Run the server:
   ```sh
   go run server.go
   ```

## Back-End Documentation
For full API reference and WebSocket implementation details, check the Dueljack Documentation on [Russian](/docs/DOCS_RU.md) or [English](/docs/DOCS_EN.md).

## User Manual and Game Rools
User guidelines and game rools are available on two languages: [Russian](/docs/USER_MANUAL_RU.md) and [English](/docs/USER_MANUAL_EN.md).