package user_entity

import (
	"auctionService/internal/internal_error"
	"context"
)

type User struct {
	Id   string
	Name string
}

type UserRepositoryInterface interface {
	FindUserById(
		ctx context.Context, userId string) (*User, *internal_error.InternalError)
}
