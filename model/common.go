package model

const (
	STATE_EXIST   = 1
	STATE_DELETED = 2
)

type OrderType int

const (
	OrderTypeSell OrderType = 1
	OrderTypeBuy  OrderType = 2
)

type OrderState int

const (
	OrderStateCreate       OrderState = 1
	OrderStateFinish       OrderState = 2
	OrderStateCancel       OrderState = 3
	OrderStatePreCreate    OrderState = 4
	OrderStateTimeout      OrderState = 5
	OrderStateErr          OrderState = 6
	OrderStateFinishButErr OrderState = 8

	OrderStatePreAsk      OrderState = 7
	OrderStatePreClaim    OrderState = 9
	OrderStateFinishClaim OrderState = 10

	//100 all
	OrderStateAll OrderState = 100
)

type FreeState int

const (
	FreeStateYes   FreeState = 1
	FreeStateClaim FreeState = 2
)
