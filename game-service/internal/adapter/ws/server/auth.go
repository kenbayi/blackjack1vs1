package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"game_svc/pkg/security"
)

// contextKey тип для ключей контекста, чтобы избежать коллизий.
type contextKey string

// UserIDKey ключ для хранения userID в контексте запроса.
const UserIDKey contextKey = "userID"

// AuthJWTMiddleware проверяет JWT токен.
// Если токен валиден, добавляет userID в контекст запроса.
// Теперь он будет использовать предоставленный JWTManager.
func AuthJWTMiddleware(next http.Handler, jwtManager *security.JWTManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("AuthJWTMiddleware: Attempting to authenticate WebSocket upgrade request")

		tokenStr := ""
		queryToken := r.URL.Query().Get("token") // Пытаемся извлечь из query-параметра "token"
		if queryToken != "" {
			tokenStr = queryToken
			log.Println("AuthJWTMiddleware: Token found in query parameter")
		} else {
			authHeader := r.Header.Get("Authorization") // Пытаемся извлечь из заголовка "Authorization"
			if authHeader != "" && strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
				tokenStr = authHeader[len("Bearer "):]
				log.Println("AuthJWTMiddleware: Token found in Authorization header")
			}
		}

		if tokenStr == "" {
			log.Println("AuthJWTMiddleware: Authentication token not found in request")
			http.Error(w, "Unauthorized: Missing token", http.StatusUnauthorized)
			return
		}

		// Используем JWTManager для верификации токена
		claims, err := jwtManager.Verify(tokenStr) //
		if err != nil {
			log.Printf("AuthJWTMiddleware: Invalid token: %v", err)
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}

		// Извлекаем userID из claims.
		// Твой JWTManager.GenerateAccessToken использует "user_id" и оно int64.
		// JWTManager.GenerateRefreshToken также использует "user_id" и оно int64.
		// Claims возвращаются как jwt.MapClaims, где значения - interface{}.
		var userIDStr string
		userIDClaim, okUserID := claims["user_id"]
		if !okUserID {
			log.Println("AuthJWTMiddleware: 'user_id' claim missing")
			http.Error(w, "Unauthorized: Invalid token claims (missing user_id)", http.StatusUnauthorized)
			return
		}

		// jwt.MapClaims может возвращать числа как float64
		switch userIDVal := userIDClaim.(type) {
		case float64: // Стандартно для чисел из JSON/JWT
			userIDStr = fmt.Sprintf("%.0f", userIDVal)
		case int64:
			userIDStr = strconv.FormatInt(userIDVal, 10)
		case string:
			userIDStr = userIDVal
		default:
			log.Printf("AuthJWTMiddleware: 'user_id' claim has unexpected type: %T", userIDClaim)
			http.Error(w, "Unauthorized: Invalid token claims (user_id type error)", http.StatusUnauthorized)
			return
		}

		if userIDStr == "" {
			log.Println("AuthJWTMiddleware: UserID claim ('user_id') is empty after type assertion/conversion")
			http.Error(w, "Unauthorized: Invalid token claims (empty user_id)", http.StatusUnauthorized)
			return
		}

		log.Printf("AuthJWTMiddleware: User %s authenticated successfully", userIDStr)
		ctxWithUser := context.WithValue(r.Context(), UserIDKey, userIDStr)
		next.ServeHTTP(w, r.WithContext(ctxWithUser))
	})
}
