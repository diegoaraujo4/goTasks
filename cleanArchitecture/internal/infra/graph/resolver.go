package graph

import (
	"cleanarch/internal/entity"
	"cleanarch/internal/usecase"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	CreateOrderUseCase usecase.CreateOrderUseCase
	ListOrdersUseCase  usecase.ListOrdersUseCase
	OrderRepository    entity.OrderRepositoryInterface
}
