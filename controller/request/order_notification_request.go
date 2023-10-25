package request

type OrderNotificationFetchReq struct {
	Net              string `json:"net"` //livenet/signet/testnet
	Tick             string `json:"tick"`
	Address          string `json:"address"`
	NotificationType int64  `json:"notificationType"`
}
