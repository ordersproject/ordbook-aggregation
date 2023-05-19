package major

import (
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	DsOrderbook = "orderbook"
)

func GetOrderbookDb()  (*mongo.Client, error) {
	return GetDBWith(DsOrderbook)
}

