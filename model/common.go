package model

const (
	STATE_EXIST = 1
	STATE_DELETED = 2
)

type OrderType int

const (
	OrderTypeSell OrderType = 1
	OrderTypeBuy  OrderType = 2
)

type OrderState int

const (
	OrderStateCreate    OrderState = 1
	OrderStateFinish    OrderState = 2
	OrderStateCancel    OrderState = 3
	OrderStatePreCreate OrderState = 4
	OrderStateTimeout   OrderState = 5
	OrderStateErr       OrderState = 6
)