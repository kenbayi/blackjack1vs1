package dao

import (
	"strconv"
	"strings"
	"time"
	"user_svc/internal/model"
)

type User struct {
	ID        int64     `db:"id"`
	Username  string    `db:"username"`
	Email     string    `db:"email"`
	Role      string    `db:"role"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	IsDeleted bool      `db:"is_deleted"`
	Nickname  *string   `db:"nickname"`
	Bio       *string   `db:"bio"`
	Balance   *int64    `db:"balance"`
	Rating    *int64    `db:"rating"`
}

func ToUser(u User) model.User {
	return model.User{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		IsDeleted: u.IsDeleted,
		Nickname:  u.Nickname,
		Bio:       u.Bio,
		Balance:   u.Balance,
		Rating:    u.Rating,
	}
}

func FromUserFilter(filter model.UserFilter) (string, []interface{}) {
	var query string
	var args []interface{}
	var conditions []string
	argCounter := 1

	if filter.ID != nil {
		conditions = append(conditions, "id = $"+strconv.Itoa(argCounter))
		args = append(args, *filter.ID)
		argCounter++
	}
	if filter.Username != nil {
		conditions = append(conditions, "username = $"+strconv.Itoa(argCounter))
		args = append(args, *filter.Username)
		argCounter++
	}
	if filter.Email != nil {
		conditions = append(conditions, "email = $"+strconv.Itoa(argCounter))
		args = append(args, *filter.Email)
		argCounter++
	}
	if filter.IsDeleted != nil {
		conditions = append(conditions, "is_deleted = $"+strconv.Itoa(argCounter))
		args = append(args, *filter.IsDeleted)
		argCounter++
	}
	if filter.Nickname != nil {
		conditions = append(conditions, "nickname = $"+strconv.Itoa(argCounter))
		args = append(args, *filter.Nickname)
		argCounter++
	}
	if filter.Balance != nil {
		conditions = append(conditions, "balance = $"+strconv.Itoa(argCounter))
		args = append(args, *filter.Balance)
		argCounter++
	}
	if filter.Rating != nil {
		conditions = append(conditions, "rating = $"+strconv.Itoa(argCounter))
		args = append(args, *filter.Rating)
		argCounter++
	}

	if len(conditions) > 0 {
		query = " WHERE " + strings.Join(conditions, " AND ")
	}

	return query, args
}
