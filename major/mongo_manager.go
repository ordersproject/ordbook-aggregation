package major

import (
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	DsOrdbook = "ordbook"
)

func GetOrderbookDb()  (*mongo.Client, error) {
	return GetDBWith(DsOrdbook)
}

