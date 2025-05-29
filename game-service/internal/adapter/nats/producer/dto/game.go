package dto

import (
	eventsproto "game_svc/internal/adapter/grpc/server/frontend/proto/events"
	"game_svc/internal/model"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"strconv"
	"time"
)

func FromResult(standResult *model.Result, roomBet int64) *eventsproto.GameResult {
	if standResult == nil {
		log.Println("FromResult: received nil standResult, returning nil event")
		return nil
	}

	var p1Data *eventsproto.PlayerGameResult
	var p2Data *eventsproto.PlayerGameResult

	// We assume FinalScores (or FinalHands) contains entries for the two players.
	playerIDsFromGame := make([]string, 0, 2)
	if standResult.FinalScores != nil {
		for pid := range standResult.FinalScores {
			playerIDsFromGame = append(playerIDsFromGame, pid)
			if len(playerIDsFromGame) == 2 {
				break
			}
		}
	}

	if len(playerIDsFromGame) == 0 {
		log.Printf("FromResult: No player data found in StandResult.FinalScores for room %s. Cannot populate PlayerResult.", standResult.RoomID)
	}

	// Populate Player1 data
	if len(playerIDsFromGame) >= 1 {
		playerID1Str := playerIDsFromGame[0]
		score1, score1Ok := standResult.FinalScores[playerID1Str]
		hand1, hand1Ok := standResult.FinalHands[playerID1Str]

		if score1Ok && hand1Ok {
			p1Data = &eventsproto.PlayerGameResult{
				PlayerId:   toInt64(playerID1Str),
				FinalScore: int32(score1),
				FinalHand:  convertHandModelToStringSlice(hand1),
			}
		} else {
			log.Printf("FromResult: Missing score or hand for player1 ID %s in room %s", playerID1Str, standResult.RoomID)
			p1Data = &eventsproto.PlayerGameResult{PlayerId: toInt64(playerID1Str)}
		}
	}

	// Populate Player2 data
	if len(playerIDsFromGame) >= 2 {
		playerID2Str := playerIDsFromGame[1]
		score2, score2Ok := standResult.FinalScores[playerID2Str]
		hand2, hand2Ok := standResult.FinalHands[playerID2Str]

		if score2Ok && hand2Ok {
			p2Data = &eventsproto.PlayerGameResult{
				PlayerId:   toInt64(playerID2Str),
				FinalScore: int32(score2),
				FinalHand:  convertHandModelToStringSlice(hand2),
			}
		} else {
			log.Printf("FromResult: Missing score or hand for player2 ID %s in room %s", playerID2Str, standResult.RoomID)
			p2Data = &eventsproto.PlayerGameResult{PlayerId: toInt64(playerID2Str)} // At least ID
		}
	}

	// Construct the main event message
	event := &eventsproto.GameResult{
		RoomId:    standResult.RoomID,
		WinnerId:  toInt64(standResult.Winner),
		LoserId:   toInt64(standResult.Loser),
		Bet:       roomBet,
		CreatedAt: timestamppb.New(time.Now()),
		Player1:   p1Data,
		Player2:   p2Data,
	}

	return event
}

// Helper function to convert model.Card to string (as discussed before)
func cardModelToString(card model.Card) string {
	return card.Value + card.Suit
}

// Helper function to convert []model.Card to []string
func convertHandModelToStringSlice(modelHand []model.Card) []string {
	if modelHand == nil {
		return []string{}
	}
	stringHand := make([]string, len(modelHand))
	for i, card := range modelHand {
		stringHand[i] = cardModelToString(card)
	}
	return stringHand
}

// Helper function to safely convert string ID to int64, handling "0" or errors
func toInt64(playerIDStr string) int64 {
	if playerIDStr == "" || playerIDStr == "0" {
		return 0
	}
	id, err := strconv.ParseInt(playerIDStr, 10, 64)
	if err != nil {
		log.Printf("Warning: Could not parse player ID '%s' to int64: %v. Defaulting to 0.", playerIDStr, err)
		return 0
	}
	return id
}
