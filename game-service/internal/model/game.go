package model

import "time"

// Player представляет игрока в доменной логике.
type Player struct {
	ID         string
	IsReady    bool
	Score      int
	LastAction string
	Hand       []Card
}

// Card представляет игральную карту.
type Card struct {
	Value string // "A", "2", ..., "K"
	Suit  string // "H", "D", "C", "S" // Hearts, Diamonds, Clubs, Spades
}

// Room представляет игровую комнату в доменной логике.
type Room struct {
	ID                  string
	Status              string // "waiting", "in_progress", "finished"
	Bet                 int
	Players             []*Player // Список игроков в комнате
	Deck                []Card    // Игровая колода для этой комнаты (будет управляться GameUseCase)
	CurrentTurnPlayerID string    // ID игрока, чей сейчас ход (может быть пустым)
}

type PlayerReadyResult struct {
	UpdatedRoom         *Room
	GameJustStarted     bool
	PlayerIDReady       string
	IsPlayerNowReady    bool
	RoomRemovedFromList bool
}

type Result struct {
	RoomID             string
	PlayerID           string
	DealtCard          *Card
	NewScore           *int
	PlayerHand         *[]Card
	IsBusted           bool
	GameEnded          bool
	Winner             string
	Loser              string
	FinalScores        map[string]int
	FinalHands         map[string][]Card
	NextTurnPlayerID   string
	PlayerCurrentScore *int
	AllPlayerScores    *map[string]int
}

type User struct {
	ID        int64
	Username  string
	Email     string
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
	IsDeleted bool
	Nickname  *string
	Bio       *string
	Balance   *int64
	Rating    *int64
}

type Opponent struct {
	ID  string
	MMR int64
}

type Match struct {
	RoomID  string
	Players []string
}
