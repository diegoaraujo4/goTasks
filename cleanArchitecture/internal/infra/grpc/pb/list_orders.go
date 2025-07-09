package pb

// Temporary types for ListOrders functionality
type ListOrdersRequest struct{}

type ListOrdersResponse struct {
	Orders []*CreateOrderResponse
}
