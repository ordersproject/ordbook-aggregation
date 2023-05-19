package mongo_service

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//var (
//	indexMap map[string]string
//)
//
//func init() {
//	indexMap = make(map[string]string)
//}

func CreateUniqueIndex(collection *mongo.Collection, indexName string)  {
	//if _, ok := indexMap[indexName]; ok {
	//	if indexMap[indexName] == collection.Name() {
	//		return
	//	}
	//}

	if indexItemMap != nil && indexItemMap.CheckIndexName(collection.Name(), indexName) {
		return
	}

	index, err := collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bson.M{
				indexName: 1,
			},
			Options: options.Index().SetUnique(true),
		},
	)
	_ = index
	if err != nil {
		fmt.Println("Create one unique index err:", err)
	}else {
		//fmt.Println("Create one unique index:", index)
	}

	if indexItemMap != nil {
		indexItemMap.Set(collection.Name(), indexName)
	}
	//indexMap[indexName] = collection.Name()
}


func CreateIndex(collection *mongo.Collection, indexName string)  {
	//if _, ok := indexMap[indexName]; ok {
	//	if indexMap[indexName] == collection.Name() {
	//		return
	//	}
	//}

	if indexItemMap != nil && indexItemMap.CheckIndexName(collection.Name(), indexName) {
		return
	}

	index, err := collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bson.M{
				indexName: 1,
			},
			Options: options.Index(),
		},
	)
	_ = index
	if err != nil {
		fmt.Println("Create one index err:", err)
	}else {
		//fmt.Println("Create one index:", index)
	}

	//indexMap[indexName] = collection.Name()
	if indexItemMap != nil {
		indexItemMap.Set(collection.Name(), indexName)
	}
}

//func CreateMultiIndex(collection *mongo.Collection, indexName... string)  {
//	if _, ok := indexMap[indexName]; ok {
//		if indexMap[indexName] == collection.Name() {
//			return
//		}
//	}
//
//	index, err := collection.Indexes().CreateOne(
//		context.Background(),
//		mongo.IndexModel{
//			Keys: bson.M{
//				indexName: 1,
//			},
//			Options: options.Index(),
//		},
//	)
//	if err != nil {
//		fmt.Println("Create one index err:", err)
//	}else {
//		fmt.Println("Create one index:", index)
//	}
//
//	indexMap[indexName] = collection.Name()
//}