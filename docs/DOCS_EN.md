# Documentation for Dueljack

## 1. General Information

This documentation describes the Back-End part of the card game "Dueljack," developed in **Go**, using **PostgreSQL** for data storage, **Redis** for managing game sessions, and **WebSockets** for real-time player interactions.

------

## 2. Project Structure

The project consists of the following directories and files:

### Root Directory:

- **`server.go`** — Main server file.
- **`go.mod` and `go.sum`** — Go dependency management files.
- **`config/`** — Contains configuration files (e.g., database and Redis settings).
- **`dump.rdb`** — Redis database.
- **`src/`** — Project source code.

### `src/` Directory:

- **`models/`** — Contains data structures (users, game rooms, etc.).
- **`db/`** — Handles interaction with PostgreSQL and Redis.
- **`handlers/`** — API handlers and WebSocket connections.
- **`middlewares/`** — Utility functions (logging, authentication, etc.).

------

## 3. Data Models

### 3.1 User (**models/user.go**)

```go
type User struct {
	ID        int       `json:"id"`         // Unique identifier
	Username  string    `json:"username"`   // Username
	Password  string    `json:"-"`          // Hashed password (not exposed in JSON)
	CreatedAt time.Time `json:"created_at"` // Account creation timestamp
	Balance   int       `json:"balance"`    // User Balance
}
```

### 3.2 Game Room (**models/room.go**)

```go
type GameRoom struct {
	ID        int       `json:"id"`         // Unique identifier
	RoomID    string    `json:"room_id"`    // UUID for the room
	Player1ID int       `json:"player1_id"` // Player 1's ID
	Player2ID int       `json:"player2_id"` // Player 2's ID
	Status    string    `json:"status"`     // Room status: "finished"
	Winner    string    `json:"winner"`     // Winner of the room
	CreatedAt time.Time `json:"created_at"` // Room creation timestamp
}
```

------

## 4. API Endpoints

### 4.1 Authentication

#### `POST /register`

Register a new user.

- **Request Body:**

```json
{
  "username": "newPlayer",
  "password": "newPassword"
}
```

- **Response:**

```json
{
  "message": "User registered successfully"
}
```

- Errors:
  - `400 Bad Request` — Username already exists.

#### `POST /login`

Authenticate a user.

- **Request Body:**

```json
{
  "username": "player1",
  "password": "securepass"
}
```

- **Response:**

```json
{
  "token": "uuid-session-token"
}
```

- Errors:
  - `401 Unauthorized` — Invalid username or password.

#### `POST /logout`

Logout a user and invalidate session.

- **Request Body:**

```json
{
  "token": "uuid-session-token"
}
```

- **Response:**

```json
{
  "message": "User logged out successfully"
}
```

------

### 4.2 User Management

#### `GET /user/{username}`

Retrieve user details by username.

- **Response:**

```json
{
  "id": 1,
  "username": "player1",
  "created_at": "2024-02-20T15:30:00Z",
  "balance": 5000
}
```

#### `PUT /updProfile`

Update user profile details.

- **Request Body:**

```json
{
  "user_id": 1,
  "username": "newUsername"
}
```

- **Response:**

```json
{
  "message": "Profile updated successfully"
}
```

#### `PUT /updBalance`

Update user balance.

- **Request Body:**

```json
{
  "user_id": 1,
  "amount": 10000
}
```

- **Response:**

```json
{
  "message": "Balance updated successfully"
}
```

#### `DELETE /user/{id}`

Delete a user by ID.

- **Response:**

```json
{
  "message": "User deleted successfully"
}
```

------

### 4.3 Game Room Management

#### `GET /rooms`

Retrieve the list of available game rooms.

- **Response:**

```json
[
  {
    "roomID": "uuid-room1",
    "status": "waiting",
    "players": 1,
    "bet": 2000
  },
  {
    "roomID": "uuid-room2",
    "status": "in_progress",
    "players": 2,
    "bet": 4000
  }
]
```

#### `GET /history/{id}`

Retrieve game history by user ID.

- **Response:**

```json
{
  "id": 123,
  "room_id": "550e8400-e29b-41d4-a716-446655440000",
  "player1_id": 123,
  "player2_id": 456,
  "status": "finished",
  "winner": 456,
  "created_at": "2024-02-20T15:30:00Z"
}
```

------

### 4.4 Session Management

#### `GET /session`

Check the current session status.

- **Response:**

```json
{
  "user_id": 1,
  "message": "Session is valid"
}
```

------

### 4.5 WebSockets

#### `GET /ws`

Establish a WebSocket connection for real-time game interactions.

- **Description:** This endpoint is used to establish and manage real-time communication between players in a game room.

------

## 5. Web-Socket messages

The application uses WebSocket for real-time player interactions.

### 5.1 WebSocket Message Structure

```json
{
  "type": "hit",
  "content": {
    "roomID": "uuid-roomID"
  }
}
```

- **type** — Event type (`playerAction (hit, stand)`, `create_room`, `join_room`, `ready`).
- **content** — Message content (player ID, action, etc.).

### 5.1.2 Available Commands

- `ready` — Confirm readiness for the game. 
- `hit` — Take a card.
- `stand` — Pass the turn.
- `create_room` — To create room.
- `join_room` — To join existing room.
- `leave_room` — To kick player from the room, if he doesn't have enough balance for the room. 

### 5.2 Game Room Management

#### `create_room`

Creates a new game room.

- **Request Body:**

```json
{
	"type": "create_room",
	"content": {
		"bet": 2000
	}
}
```

- **Response:**

```json
{
  "room_id": "uuid-roomID"
}
```

- Errors:
  - `400 Bad Request` — Insufficient funds.
  - `500 Internal Server Error` — Error while creating the room.

#### `join_room`

Joins to a existing room.

- **Request Body:**

```json
{
	"type": "join_room",
	"content": {
		"roomID": "uuid-roomID",
		"bet": 2000
	}
}
```

------

## 6. Component Interaction

1. **User authentication** via API (`/login`, `/register`).
2. **Player creates a room** if they have enough funds (`/createRoom`).
3. **Another player joins** a room by finding it in the list (`/joinRoom`).
4. The gameplay is managed through WebSocket messages:
   - `playerAction` — Handles player actions (hit, stand, etc.).
   - `gameStart` — Starts the game.
   - `gameEnd` — Ends the game and records the result.
5. **After the game ends, results are stored in PostgreSQL**.

------

## 7. Database Management

### 7.1 PostgreSQL

Used for storing user and game data.

### 7.2 Redis

Used for storing game rooms and quickly accessing session information.

------

## 8. Running the Project

### 8.1 Installing Dependencies

```sh
go mod tidy
```

### 8.2 Starting the Server

```sh
go run server.go
```